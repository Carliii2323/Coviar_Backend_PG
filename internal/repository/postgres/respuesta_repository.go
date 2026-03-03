package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository"
)

type RespuestaRepository struct {
	db *sql.DB
}

func NewRespuestaRepository(db *sql.DB) repository.RespuestaRepository {
	return &RespuestaRepository{db: db}
}

func (r *RespuestaRepository) Create(ctx context.Context, tx repository.Transaction, respuesta *domain.Respuesta) (int, error) {
	query := `
		INSERT INTO respuestas (id_nivel_respuesta, id_indicador, id_autoevaluacion)
		VALUES ($1, $2, $3)
		RETURNING id_respuesta
	`

	var id int
	err := r.db.QueryRowContext(ctx, query, respuesta.IDNivelRespuesta, respuesta.IDIndicador, respuesta.IDAutoevaluacion).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error creating respuesta: %w", err)
	}

	respuesta.ID = id
	return id, nil
}

func (r *RespuestaRepository) Upsert(ctx context.Context, tx repository.Transaction, respuesta *domain.Respuesta) (int, error) {
	query := `
		INSERT INTO respuestas (id_nivel_respuesta, id_indicador, id_autoevaluacion)
		VALUES ($1, $2, $3)
		ON CONFLICT (id_autoevaluacion, id_indicador) 
		DO UPDATE SET id_nivel_respuesta = EXCLUDED.id_nivel_respuesta
		RETURNING id_respuesta
	`

	var id int
	err := r.db.QueryRowContext(ctx, query, respuesta.IDNivelRespuesta, respuesta.IDIndicador, respuesta.IDAutoevaluacion).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error upserting respuesta: %w", err)
	}

	respuesta.ID = id
	return id, nil
}

func (r *RespuestaRepository) FindByAutoevaluacion(ctx context.Context, idAutoevaluacion int) ([]*domain.Respuesta, error) {
	query := `
		SELECT id_respuesta, id_nivel_respuesta, id_indicador, id_autoevaluacion
		FROM respuestas
		WHERE id_autoevaluacion = $1
	`

	rows, err := r.db.QueryContext(ctx, query, idAutoevaluacion)
	if err != nil {
		return nil, fmt.Errorf("error querying respuestas: %w", err)
	}
	defer rows.Close()

	var respuestas []*domain.Respuesta
	for rows.Next() {
		resp := &domain.Respuesta{}
		if err := rows.Scan(&resp.ID, &resp.IDNivelRespuesta, &resp.IDIndicador, &resp.IDAutoevaluacion); err != nil {
			return nil, fmt.Errorf("error scanning respuesta: %w", err)
		}
		respuestas = append(respuestas, resp)
	}

	return respuestas, rows.Err()
}

func (r *RespuestaRepository) DeleteByAutoevaluacion(ctx context.Context, idAutoevaluacion int) error {
	query := `DELETE FROM respuestas WHERE id_autoevaluacion = $1`

	_, err := r.db.ExecContext(ctx, query, idAutoevaluacion)
	if err != nil {
		return fmt.Errorf("error deleting respuestas: %w", err)
	}

	return nil
}

func (r *RespuestaRepository) CalculateTotalScore(ctx context.Context, idAutoevaluacion int) (int, error) {
	query := `
		SELECT COALESCE(SUM(nr.puntos), 0) as total_puntos
		FROM respuestas r
		INNER JOIN niveles_respuesta nr ON r.id_nivel_respuesta = nr.id_nivel_respuesta
		WHERE r.id_autoevaluacion = $1
	`

	var totalPuntos int
	err := r.db.QueryRowContext(ctx, query, idAutoevaluacion).Scan(&totalPuntos)
	if err != nil {
		return 0, fmt.Errorf("error calculating total score: %w", err)
	}

	return totalPuntos, nil
}
