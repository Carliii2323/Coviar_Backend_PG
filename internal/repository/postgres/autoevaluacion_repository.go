package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository"
)

type AutoevaluacionRepository struct {
	db *sql.DB
}

func NewAutoevaluacionRepository(db *sql.DB) repository.AutoevaluacionRepository {
	return &AutoevaluacionRepository{db: db}
}

func (r *AutoevaluacionRepository) Create(ctx context.Context, tx repository.Transaction, auto *domain.Autoevaluacion) (int, error) {
	query := `
		INSERT INTO autoevaluaciones (fecha_inicio, estado, id_bodega, id_responsable)
		VALUES (NOW(), $1, $2, $3)
		RETURNING id_autoevaluacion
	`

	var id int
	err := r.db.QueryRowContext(ctx, query, domain.EstadoPendiente, auto.IDBodega, auto.IDResponsable).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error creating autoevaluacion: %w", err)
	}

	auto.ID = id
	auto.FechaInicio = time.Now()
	auto.Estado = domain.EstadoPendiente
	return id, nil
}

/*func (r *AutoevaluacionRepository) FindByID(ctx context.Context, id int) (*domain.Autoevaluacion, error) {
	query := `
		SELECT id_autoevaluacion, fecha_inicio, fecha_fin, estado, id_bodega, id_segmento
		FROM autoevaluaciones WHERE id_autoevaluacion = $1
	`

	auto := &domain.Autoevaluacion{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&auto.ID, &auto.FechaInicio, &auto.FechaFin, &auto.Estado, &auto.IDBodega, &auto.IDSegmento,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error finding autoevaluacion: %w", err)
	}

	return auto, nil
}*/

func (r *AutoevaluacionRepository) FindByID(ctx context.Context, id int) (*domain.Autoevaluacion, error) {
	query := `
		SELECT id_autoevaluacion, fecha_inicio, fecha_fin, estado, id_bodega, id_segmento,
		       puntaje_final, id_nivel_sostenibilidad, id_responsable
		FROM autoevaluaciones WHERE id_autoevaluacion = $1
	`

	auto := &domain.Autoevaluacion{}
	var idResp sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&auto.ID, &auto.FechaInicio, &auto.FechaFin, &auto.Estado, &auto.IDBodega, &auto.IDSegmento,
		&auto.PuntajeFinal, &auto.IDNivelSostenibilidad, &idResp,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error finding autoevaluacion: %w", err)
	}
	if idResp.Valid {
		v := int(idResp.Int64)
		auto.IDResponsable = &v
	}

	return auto, nil
}

func (r *AutoevaluacionRepository) UpdateSegmento(ctx context.Context, id int, idSegmento int) error {
	query := `UPDATE autoevaluaciones SET id_segmento = $1 WHERE id_autoevaluacion = $2`

	_, err := r.db.ExecContext(ctx, query, idSegmento, id)
	if err != nil {
		return fmt.Errorf("error updating segmento: %w", err)
	}

	return nil
}

/*func (r *AutoevaluacionRepository) Complete(ctx context.Context, id int) error {
	query := `UPDATE autoevaluaciones SET estado = $1, fecha_fin = NOW() WHERE id_autoevaluacion = $2`

	_, err := r.db.ExecContext(ctx, query, domain.EstadoCompletada, id)
	if err != nil {
		return fmt.Errorf("error completing autoevaluacion: %w", err)
	}

	return nil
}*/

func (r *AutoevaluacionRepository) Complete(ctx context.Context, id int) error {
	query := `
		UPDATE autoevaluaciones 
		SET estado = $1, 
		    fecha_fin = NOW(),
		    puntaje_final = $2,
		    id_nivel_sostenibilidad = $3
		WHERE id_autoevaluacion = $4
	`

	_, err := r.db.ExecContext(ctx, query, domain.EstadoCompletada, id)
	if err != nil {
		return fmt.Errorf("error completing autoevaluacion: %w", err)
	}

	return nil
}

// CompleteWithScore completa la autoevaluación con el puntaje calculado y nivel de sostenibilidad
func (r *AutoevaluacionRepository) CompleteWithScore(ctx context.Context, id int, puntajeFinal int, idNivelSostenibilidad int) error {
	query := `
		UPDATE autoevaluaciones 
		SET estado = $1, 
		    fecha_fin = NOW(),
		    puntaje_final = $2,
		    id_nivel_sostenibilidad = $3
		WHERE id_autoevaluacion = $4
	`

	_, err := r.db.ExecContext(ctx, query, domain.EstadoCompletada, puntajeFinal, idNivelSostenibilidad, id)
	if err != nil {
		return fmt.Errorf("error completing autoevaluacion with score: %w", err)
	}

	return nil
}

/*func (r *AutoevaluacionRepository) FindPendienteByBodega(ctx context.Context, idBodega int) (*domain.Autoevaluacion, error) {
	query := `
		SELECT id_autoevaluacion, fecha_inicio, fecha_fin, estado, id_bodega, id_segmento
		FROM autoevaluaciones
		WHERE id_bodega = $1 AND estado = $2
	`

	auto := &domain.Autoevaluacion{}
	err := r.db.QueryRowContext(ctx, query, idBodega, domain.EstadoPendiente).Scan(
		&auto.ID, &auto.FechaInicio, &auto.FechaFin, &auto.Estado, &auto.IDBodega, &auto.IDSegmento,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No hay autoevaluación pendiente
		}
		return nil, fmt.Errorf("error finding pending autoevaluacion: %w", err)
	}

	return auto, nil
}*/

func (r *AutoevaluacionRepository) FindPendienteByBodega(ctx context.Context, idBodega int) (*domain.Autoevaluacion, error) {
	query := `
		SELECT id_autoevaluacion, fecha_inicio, fecha_fin, estado, id_bodega, id_segmento,
		       puntaje_final, id_nivel_sostenibilidad, id_responsable
		FROM autoevaluaciones
		WHERE id_bodega = $1 AND estado = $2
	`

	auto := &domain.Autoevaluacion{}
	var idResp sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, idBodega, domain.EstadoPendiente).Scan(
		&auto.ID, &auto.FechaInicio, &auto.FechaFin, &auto.Estado, &auto.IDBodega, &auto.IDSegmento,
		&auto.PuntajeFinal, &auto.IDNivelSostenibilidad, &idResp,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding pending autoevaluacion: %w", err)
	}
	if idResp.Valid {
		v := int(idResp.Int64)
		auto.IDResponsable = &v
	}

	return auto, nil
}

func (r *AutoevaluacionRepository) Cancel(ctx context.Context, id int) error {
	query := `UPDATE autoevaluaciones SET estado = $1, fecha_fin = NOW() WHERE id_autoevaluacion = $2`

	_, err := r.db.ExecContext(ctx, query, domain.EstadoCancelada, id)
	if err != nil {
		return fmt.Errorf("error canceling autoevaluacion: %w", err)
	}

	return nil
}

func (r *AutoevaluacionRepository) HasPendingByBodega(ctx context.Context, idBodega int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM autoevaluaciones WHERE id_bodega = $1 AND estado = $2)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, idBodega, domain.EstadoPendiente).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking pending autoevaluacion: %w", err)
	}

	return exists, nil
}

func (r *AutoevaluacionRepository) UpdateEvidenciaStatus(ctx context.Context, id int, estado domain.EstadoEvidencia) error {
	query := `UPDATE autoevaluaciones SET estado_evidencia = $1 WHERE id_autoevaluacion = $2`

	_, err := r.db.ExecContext(ctx, query, string(estado), id)
	if err != nil {
		return fmt.Errorf("error updating evidencia status: %w", err)
	}

	return nil
}

func (r *AutoevaluacionRepository) FindCompletadasByBodega(ctx context.Context, idBodega int) ([]*domain.Autoevaluacion, error) {
	query := `
		SELECT id_autoevaluacion, id_bodega, id_segmento, estado, puntaje_final,
		       id_nivel_sostenibilidad, fecha_inicio, fecha_fin, id_responsable
		FROM autoevaluaciones
		WHERE id_bodega = $1 AND estado = 'COMPLETADA'
		ORDER BY fecha_inicio DESC
	`

	rows, err := r.db.QueryContext(ctx, query, idBodega)
	if err != nil {
		return nil, fmt.Errorf("error finding completadas by bodega: %w", err)
	}
	defer rows.Close()

	var autoevaluaciones []*domain.Autoevaluacion
	for rows.Next() {
		auto := &domain.Autoevaluacion{}
		var idResp sql.NullInt64
		err := rows.Scan(
			&auto.ID, &auto.IDBodega, &auto.IDSegmento, &auto.Estado,
			&auto.PuntajeFinal, &auto.IDNivelSostenibilidad,
			&auto.FechaInicio, &auto.FechaFin, &idResp,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning autoevaluacion completada: %w", err)
		}
		if idResp.Valid {
			v := int(idResp.Int64)
			auto.IDResponsable = &v
		}
		autoevaluaciones = append(autoevaluaciones, auto)
	}

	return autoevaluaciones, rows.Err()
}

func (r *AutoevaluacionRepository) FindLastCompletadaByBodega(ctx context.Context, idBodega int) (*domain.Autoevaluacion, error) {
	query := `
		SELECT id_autoevaluacion, id_bodega, id_segmento, estado, puntaje_final,
		       id_nivel_sostenibilidad, fecha_inicio, fecha_fin, id_responsable
		FROM autoevaluaciones
		WHERE id_bodega = $1 AND estado = 'COMPLETADA'
		ORDER BY fecha_fin DESC
		LIMIT 1
	`

	auto := &domain.Autoevaluacion{}
	var idResp sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, idBodega).Scan(
		&auto.ID, &auto.IDBodega, &auto.IDSegmento, &auto.Estado,
		&auto.PuntajeFinal, &auto.IDNivelSostenibilidad,
		&auto.FechaInicio, &auto.FechaFin, &idResp,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding last completada: %w", err)
	}
	if idResp.Valid {
		v := int(idResp.Int64)
		auto.IDResponsable = &v
	}

	return auto, nil
}
