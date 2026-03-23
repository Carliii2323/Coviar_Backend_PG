package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository"
)

type UbicacionRepository struct {
	db *sql.DB
}

func NewUbicacionRepository(db *sql.DB) repository.UbicacionRepository {
	return &UbicacionRepository{db: db}
}

// ===== PROVINCIAS =====

func (r *UbicacionRepository) GetProvincias(ctx context.Context) ([]*domain.Provincia, error) {
	query := `SELECT id_provincia, nombre FROM provincias ORDER BY nombre`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting provincias: %w", err)
	}
	defer rows.Close()

	var provincias []*domain.Provincia
	for rows.Next() {
		provincia := &domain.Provincia{}
		if err := rows.Scan(&provincia.ID, &provincia.Nombre); err != nil {
			return nil, fmt.Errorf("error scanning provincia: %w", err)
		}
		provincias = append(provincias, provincia)
	}

	return provincias, rows.Err()
}

func (r *UbicacionRepository) GetProvinciaByID(ctx context.Context, id int) (*domain.Provincia, error) {
	query := `SELECT id_provincia, nombre FROM provincias WHERE id_provincia = $1`

	provincia := &domain.Provincia{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&provincia.ID, &provincia.Nombre)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error finding provincia: %w", err)
	}

	return provincia, nil
}

// ===== DEPARTAMENTOS =====

func (r *UbicacionRepository) GetDepartamentos(ctx context.Context) ([]*domain.Departamento, error) {
	query := `SELECT id_departamento, id_provincia, nombre FROM departamentos ORDER BY nombre`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting departamentos: %w", err)
	}
	defer rows.Close()

	var departamentos []*domain.Departamento
	for rows.Next() {
		departamento := &domain.Departamento{}
		if err := rows.Scan(&departamento.ID, &departamento.IDProvincia, &departamento.Nombre); err != nil {
			return nil, fmt.Errorf("error scanning departamento: %w", err)
		}
		departamentos = append(departamentos, departamento)
	}

	return departamentos, rows.Err()
}

func (r *UbicacionRepository) GetDepartamentoByID(ctx context.Context, id int) (*domain.Departamento, error) {
	query := `SELECT id_departamento, id_provincia, nombre FROM departamentos WHERE id_departamento = $1`

	departamento := &domain.Departamento{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&departamento.ID, &departamento.IDProvincia, &departamento.Nombre)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error finding departamento: %w", err)
	}

	return departamento, nil
}

func (r *UbicacionRepository) GetDepartamentosByProvinciaID(ctx context.Context, provinciaID int) ([]*domain.Departamento, error) {
	query := `SELECT id_departamento, id_provincia, nombre FROM departamentos WHERE id_provincia = $1 ORDER BY nombre`

	rows, err := r.db.QueryContext(ctx, query, provinciaID)
	if err != nil {
		return nil, fmt.Errorf("error getting departamentos: %w", err)
	}
	defer rows.Close()

	var departamentos []*domain.Departamento
	for rows.Next() {
		departamento := &domain.Departamento{}
		if err := rows.Scan(&departamento.ID, &departamento.IDProvincia, &departamento.Nombre); err != nil {
			return nil, fmt.Errorf("error scanning departamento: %w", err)
		}
		departamentos = append(departamentos, departamento)
	}

	return departamentos, rows.Err()
}

// ===== LOCALIDADES =====

func (r *UbicacionRepository) GetLocalidades(ctx context.Context) ([]*domain.Localidad, error) {
	query := `SELECT id_localidad, id_departamento, nombre FROM localidades ORDER BY nombre`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting localidades: %w", err)
	}
	defer rows.Close()

	var localidades []*domain.Localidad
	for rows.Next() {
		localidad := &domain.Localidad{}
		if err := rows.Scan(&localidad.ID, &localidad.IDDepartamento, &localidad.Nombre); err != nil {
			return nil, fmt.Errorf("error scanning localidad: %w", err)
		}
		localidades = append(localidades, localidad)
	}

	return localidades, rows.Err()
}

func (r *UbicacionRepository) GetLocalidadByID(ctx context.Context, id int) (*domain.Localidad, error) {
	query := `SELECT id_localidad, id_departamento, nombre FROM localidades WHERE id_localidad = $1`

	localidad := &domain.Localidad{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&localidad.ID, &localidad.IDDepartamento, &localidad.Nombre)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error finding localidad: %w", err)
	}

	return localidad, nil
}

func (r *UbicacionRepository) GetLocalidadesByDepartamentoID(ctx context.Context, departamentoID int) ([]*domain.Localidad, error) {
	query := `SELECT id_localidad, id_departamento, nombre FROM localidades WHERE id_departamento = $1 ORDER BY nombre`

	rows, err := r.db.QueryContext(ctx, query, departamentoID)
	if err != nil {
		return nil, fmt.Errorf("error getting localidades: %w", err)
	}
	defer rows.Close()

	var localidades []*domain.Localidad
	for rows.Next() {
		localidad := &domain.Localidad{}
		if err := rows.Scan(&localidad.ID, &localidad.IDDepartamento, &localidad.Nombre); err != nil {
			return nil, fmt.Errorf("error scanning localidad: %w", err)
		}
		localidades = append(localidades, localidad)
	}

	return localidades, rows.Err()
}
