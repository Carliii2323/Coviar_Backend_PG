package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository"
)

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) repository.AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) GetStats(ctx context.Context) (*domain.AdminStatsResponse, error) {
	var stats domain.AdminStatsResponse

	// Total de bodegas
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM bodegas`).Scan(&stats.TotalBodegas)
	if err != nil {
		return nil, fmt.Errorf("error counting bodegas: %w", err)
	}

	// Evaluaciones completadas
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM autoevaluaciones WHERE estado = 'COMPLETADA'`).Scan(&stats.EvaluacionesCompletadas)
	if err != nil {
		return nil, fmt.Errorf("error counting evaluaciones: %w", err)
	}

	// Promedio de sostenibilidad y distribución por nivel
	// Obtener todas las evaluaciones completadas con su nivel de sostenibilidad
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.puntaje_final, a.id_nivel_sostenibilidad, a.id_segmento,
		       COALESCE(ns.nombre, '') as nivel_nombre
		FROM autoevaluaciones a
		LEFT JOIN niveles_sostenibilidad ns ON a.id_nivel_sostenibilidad = ns.id_nivel_sostenibilidad
		WHERE a.estado = 'COMPLETADA' AND a.puntaje_final IS NOT NULL
	`)
	if err != nil {
		return nil, fmt.Errorf("error querying evaluaciones for stats: %w", err)
	}
	defer rows.Close()

	var totalPuntaje int
	var count int
	dist := domain.DistribucionNiveles{}

	for rows.Next() {
		var puntaje sql.NullInt64
		var idNivel sql.NullInt64
		var idSegmento sql.NullInt64
		var nivelNombre string

		if err := rows.Scan(&puntaje, &idNivel, &idSegmento, &nivelNombre); err != nil {
			return nil, fmt.Errorf("error scanning stats row: %w", err)
		}

		if puntaje.Valid {
			totalPuntaje += int(puntaje.Int64)
			count++
		}

		nivelLower := strings.ToLower(nivelNombre)
		if strings.Contains(nivelLower, "mínimo") || strings.Contains(nivelLower, "minimo") {
			dist.Minimo++
		} else if strings.Contains(nivelLower, "medio") {
			dist.Medio++
		} else if strings.Contains(nivelLower, "alto") {
			dist.Alto++
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stats rows: %w", err)
	}

	stats.DistribucionNiveles = dist

	if count > 0 {
		stats.PromedioSostenibilidad = float64(totalPuntaje) / float64(count)
	}

	// Determinar nivel promedio
	totalConNivel := dist.Minimo + dist.Medio + dist.Alto
	if totalConNivel > 0 {
		if dist.Alto >= dist.Medio && dist.Alto >= dist.Minimo {
			stats.NivelPromedio = "Nivel alto de sostenibilidad"
		} else if dist.Medio >= dist.Minimo {
			stats.NivelPromedio = "Nivel medio de sostenibilidad"
		} else {
			stats.NivelPromedio = "Nivel mínimo de sostenibilidad"
		}
	} else {
		stats.NivelPromedio = "Sin datos"
	}

	// Distribución por segmento: última evaluación completada de cada bodega
	segRows, err := r.db.QueryContext(ctx, `
		WITH ultima_ae AS (
			SELECT DISTINCT ON (a.id_bodega)
				a.id_bodega,
				a.id_segmento,
				a.id_nivel_sostenibilidad
			FROM autoevaluaciones a
			WHERE a.estado = 'COMPLETADA'
			  AND a.id_segmento IS NOT NULL
			  AND a.id_nivel_sostenibilidad IS NOT NULL
			ORDER BY a.id_bodega, a.fecha_fin DESC
		)
		SELECT
			s.id_segmento,
			s.nombre AS nombre_segmento,
			COUNT(CASE WHEN LOWER(ns.nombre) LIKE '%alto%' THEN 1 END)::int AS alto,
			COUNT(CASE WHEN LOWER(ns.nombre) LIKE '%medio%' OR LOWER(ns.nombre) LIKE '%media%' THEN 1 END)::int AS medio,
			COUNT(CASE WHEN LOWER(ns.nombre) LIKE '%mínimo%' OR LOWER(ns.nombre) LIKE '%minimo%' OR LOWER(ns.nombre) LIKE '%bajo%' THEN 1 END)::int AS minimo
		FROM segmentos s
		LEFT JOIN ultima_ae uae ON uae.id_segmento = s.id_segmento
		LEFT JOIN niveles_sostenibilidad ns ON uae.id_nivel_sostenibilidad = ns.id_nivel_sostenibilidad
		GROUP BY s.id_segmento, s.nombre
		ORDER BY s.id_segmento
	`)
	if err != nil {
		return nil, fmt.Errorf("error querying distribucion por segmento: %w", err)
	}
	defer segRows.Close()

	stats.DistribucionPorSegmento = []domain.SegmentoDistribucion{}
	for segRows.Next() {
		var seg domain.SegmentoDistribucion
		if err := segRows.Scan(&seg.IDSegmento, &seg.NombreSegmento, &seg.Alto, &seg.Medio, &seg.Minimo); err != nil {
			return nil, fmt.Errorf("error scanning segmento distribucion: %w", err)
		}
		stats.DistribucionPorSegmento = append(stats.DistribucionPorSegmento, seg)
	}
	if err := segRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating segmento rows: %w", err)
	}

	return &stats, nil
}

func (r *AdminRepository) GetAllEvaluaciones(ctx context.Context, estado string, idBodega int) ([]domain.EvaluacionListItem, error) {
	// Precalcular responsable activo por bodega en una sola query
	responsableMap := make(map[int]string)
	respRows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT ON (c.id_bodega) c.id_bodega, r.nombre || ' ' || r.apellido
		FROM responsables r
		JOIN cuentas c ON r.id_cuenta = c.id_cuenta
		WHERE r.activo = true AND c.id_bodega IS NOT NULL
		ORDER BY c.id_bodega, r.id_responsable
	`)
	if err == nil {
		defer respRows.Close()
		for respRows.Next() {
			var idBod int
			var nombre string
			if respRows.Scan(&idBod, &nombre) == nil {
				responsableMap[idBod] = nombre
			}
		}
	}

	// Query principal sin subqueries
	query := `
		SELECT
			a.id_autoevaluacion,
			a.id_bodega,
			b.nombre_fantasia,
			b.razon_social,
			a.estado,
			a.fecha_inicio,
			a.fecha_fin
		FROM autoevaluaciones a
		JOIN bodegas b ON a.id_bodega = b.id_bodega
		WHERE 1=1
	`

	args := []interface{}{}
	argIdx := 1

	if estado != "" {
		query += fmt.Sprintf(" AND a.estado = $%d", argIdx)
		args = append(args, estado)
		argIdx++
	}

	if idBodega > 0 {
		query += fmt.Sprintf(" AND a.id_bodega = $%d", argIdx)
		args = append(args, idBodega)
		argIdx++
	}

	query += " ORDER BY a.fecha_inicio DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying evaluaciones: %w", err)
	}
	defer rows.Close()

	var evaluaciones []domain.EvaluacionListItem
	for rows.Next() {
		var item domain.EvaluacionListItem
		var fechaFin sql.NullTime
		var fechaInicio sql.NullTime

		err := rows.Scan(
			&item.IDAutoevaluacion,
			&item.IDBodega,
			&item.NombreBodega,
			&item.RazonSocial,
			&item.Estado,
			&fechaInicio,
			&fechaFin,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning evaluacion: %w", err)
		}

		if fechaInicio.Valid {
			item.FechaInicio = fechaInicio.Time.UTC().Format("2006-01-02T15:04:05Z")
		}

		if fechaFin.Valid {
			fin := fechaFin.Time.UTC().Format("2006-01-02T15:04:05Z")
			item.FechaFin = &fin
		}

		// Responsable desde el mapa precalculado
		if nombre, ok := responsableMap[item.IDBodega]; ok {
			item.Responsable = nombre
		} else {
			item.Responsable = "N/A"
		}

		evaluaciones = append(evaluaciones, item)
	}

	if evaluaciones == nil {
		evaluaciones = []domain.EvaluacionListItem{}
	}

	return evaluaciones, rows.Err()
}
