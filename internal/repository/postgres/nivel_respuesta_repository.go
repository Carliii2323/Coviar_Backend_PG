package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository"
)

type NivelRespuestaRepository struct {
	db *sql.DB
}

func NewNivelRespuestaRepository(db *sql.DB) repository.NivelRespuestaRepository {
	return &NivelRespuestaRepository{db: db}
}

func (r *NivelRespuestaRepository) FindByIndicador(ctx context.Context, idIndicador int) ([]*domain.NivelRespuesta, error) {
	query := `
		SELECT id_nivel_respuesta, id_indicador, nombre, COALESCE(descripcion, '') as descripcion, puntos, COALESCE(posicion, 0) as posicion
		FROM niveles_respuesta
		WHERE id_indicador = $1
		ORDER BY posicion ASC
	`

	rows, err := r.db.QueryContext(ctx, query, idIndicador)
	if err != nil {
		return nil, fmt.Errorf("error querying niveles_respuesta: %w", err)
	}
	defer rows.Close()

	var niveles []*domain.NivelRespuesta
	for rows.Next() {
		nivel := &domain.NivelRespuesta{}
		if err := rows.Scan(&nivel.ID, &nivel.IDIndicador, &nivel.Nombre, &nivel.Descripcion, &nivel.Puntos, &nivel.Posicion); err != nil {
			return nil, fmt.Errorf("error scanning nivel_respuesta: %w", err)
		}
		niveles = append(niveles, nivel)
	}

	return niveles, rows.Err()
}

func (r *NivelRespuestaRepository) FindByID(ctx context.Context, id int) (*domain.NivelRespuesta, error) {
	query := `
		SELECT id_nivel_respuesta, id_indicador, nombre, COALESCE(descripcion, '') as descripcion, puntos, COALESCE(posicion, 0) as posicion
		FROM niveles_respuesta
		WHERE id_nivel_respuesta = $1
	`

	nivel := &domain.NivelRespuesta{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&nivel.ID, &nivel.IDIndicador, &nivel.Nombre, &nivel.Descripcion, &nivel.Puntos, &nivel.Posicion,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error finding nivel_respuesta: %w", err)
	}

	return nivel, nil
}

func (r *NivelRespuestaRepository) FindMaxPuntosBySegmento(ctx context.Context, idSegmento int) (map[int]int, error) {
	query := `
		SELECT i.id_indicador, MAX(nr.puntos) as max_puntos
		FROM indicadores i
		INNER JOIN niveles_respuesta nr ON i.id_indicador = nr.id_indicador
		INNER JOIN segmento_indicador si ON i.id_indicador = si.id_indicador
		WHERE si.id_segmento = $1
		GROUP BY i.id_indicador
	`

	rows, err := r.db.QueryContext(ctx, query, idSegmento)
	if err != nil {
		return nil, fmt.Errorf("error querying max puntos by segmento: %w", err)
	}
	defer rows.Close()

	result := make(map[int]int)
	for rows.Next() {
		var idIndicador, maxPuntos int
		if err := rows.Scan(&idIndicador, &maxPuntos); err != nil {
			return nil, fmt.Errorf("error scanning max puntos: %w", err)
		}
		result[idIndicador] = maxPuntos
	}

	return result, rows.Err()
}
