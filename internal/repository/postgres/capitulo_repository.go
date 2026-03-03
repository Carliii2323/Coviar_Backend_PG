package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository"
)

type CapituloRepository struct {
	db *sql.DB
}

func NewCapituloRepository(db *sql.DB) repository.CapituloRepository {
	return &CapituloRepository{db: db}
}

func (r *CapituloRepository) FindAll(ctx context.Context) ([]*domain.Capitulo, error) {
	query := `SELECT id_capitulo, nombre, descripcion, orden FROM capitulos ORDER BY orden`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying capitulos: %w", err)
	}
	defer rows.Close()

	var capitulos []*domain.Capitulo
	for rows.Next() {
		cap := &domain.Capitulo{}
		if err := rows.Scan(&cap.ID, &cap.Nombre, &cap.Descripcion, &cap.Orden); err != nil {
			return nil, fmt.Errorf("error scanning capitulo: %w", err)
		}
		capitulos = append(capitulos, cap)
	}

	return capitulos, rows.Err()
}
