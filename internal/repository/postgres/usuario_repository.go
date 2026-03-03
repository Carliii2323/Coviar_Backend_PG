package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository"
)

type UsuarioRepository struct {
	db *sql.DB
}

func NewUsuarioRepository(db *sql.DB) repository.UsuarioRepository {
	return &UsuarioRepository{db: db}
}

func (r *UsuarioRepository) Create(ctx context.Context, usuario *domain.Usuario) (int, error) {
	query := `
		INSERT INTO usuarios (email, password_hash, nombre, apellido, rol, activo, fecha_registro)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id_usuario
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		usuario.Email, usuario.PasswordHash, usuario.Nombre, usuario.Apellido, usuario.Rol, usuario.Activo,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error creating usuario: %w", err)
	}

	usuario.IdUsuario = id
	return id, nil
}

func (r *UsuarioRepository) FindByID(ctx context.Context, id int) (*domain.Usuario, error) {
	query := `
		SELECT id_usuario, email, password_hash, nombre, apellido, rol, activo, fecha_registro, ultimo_acceso
		FROM usuarios WHERE id_usuario = $1
	`

	usuario := &domain.Usuario{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&usuario.IdUsuario, &usuario.Email, &usuario.PasswordHash, &usuario.Nombre, &usuario.Apellido,
		&usuario.Rol, &usuario.Activo, &usuario.FechaRegistro, &usuario.UltimoAcceso,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error finding usuario: %w", err)
	}

	return usuario, nil
}

func (r *UsuarioRepository) FindByEmail(ctx context.Context, email string) (*domain.Usuario, error) {
	query := `
		SELECT id_usuario, email, password_hash, nombre, apellido, rol, activo, fecha_registro, ultimo_acceso
		FROM usuarios WHERE email = $1
	`

	usuario := &domain.Usuario{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&usuario.IdUsuario, &usuario.Email, &usuario.PasswordHash, &usuario.Nombre, &usuario.Apellido,
		&usuario.Rol, &usuario.Activo, &usuario.FechaRegistro, &usuario.UltimoAcceso,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding usuario by email: %w", err)
	}

	return usuario, nil
}

func (r *UsuarioRepository) GetAll(ctx context.Context) ([]*domain.Usuario, error) {
	query := `
		SELECT id_usuario, email, password_hash, nombre, apellido, rol, activo, fecha_registro, ultimo_acceso
		FROM usuario ORDER BY id_usuario
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting all usuarios: %w", err)
	}
	defer rows.Close()

	var usuarios []*domain.Usuario
	for rows.Next() {
		usuario := &domain.Usuario{}
		err := rows.Scan(
			&usuario.IdUsuario, &usuario.Email, &usuario.PasswordHash, &usuario.Nombre, &usuario.Apellido,
			&usuario.Rol, &usuario.Activo, &usuario.FechaRegistro, &usuario.UltimoAcceso,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning usuario: %w", err)
		}
		usuarios = append(usuarios, usuario)
	}

	return usuarios, rows.Err()
}

func (r *UsuarioRepository) Update(ctx context.Context, usuario *domain.Usuario) error {
	query := `
		UPDATE usuario
		SET email = $1, nombre = $2, apellido = $3, rol = $4, activo = $5, ultimo_acceso = $6
		WHERE id_usuario = $7
	`

	_, err := r.db.ExecContext(ctx, query,
		usuario.Email, usuario.Nombre, usuario.Apellido, usuario.Rol, usuario.Activo, usuario.UltimoAcceso, usuario.IdUsuario,
	)

	if err != nil {
		return fmt.Errorf("error updating usuario: %w", err)
	}

	return nil
}

func (r *UsuarioRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM usuarioss WHERE id_usuario = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting usuario: %w", err)
	}

	return nil
}
