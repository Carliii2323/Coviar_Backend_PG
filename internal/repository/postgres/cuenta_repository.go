package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository"
)

type CuentaRepository struct {
	db *sql.DB
}

func NewCuentaRepository(db *sql.DB) repository.CuentaRepository {
	return &CuentaRepository{db: db}
}

func (r *CuentaRepository) Create(ctx context.Context, tx repository.Transaction, cuenta *domain.Cuenta) (int, error) {
	// Validar restricciones antes de insertar si es necesario
	// - tipo debe ser 'BODEGA' o 'ADMINISTRADOR_APP'
	// - Si tipo = 'BODEGA', id_bodega no puede ser NULL
	// - Si tipo = 'ADMINISTRADOR_APP', id_bodega debe ser NULL
	// - email_login único, id_bodega único
	query := `
	       INSERT INTO cuentas (tipo, id_bodega, email_login, password_hash, fecha_registro)
	       VALUES ($1, $2, $3, $4, NOW())
	       RETURNING id_cuenta
       `

	var id int
	err := r.db.QueryRowContext(ctx, query, cuenta.Tipo, cuenta.IDBodega, cuenta.EmailLogin, cuenta.PasswordHash).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error creating cuenta: %w", err)
	}

	cuenta.ID = id
	cuenta.FechaRegistro = time.Now()
	return id, nil
}

func (r *CuentaRepository) FindByID(ctx context.Context, id int) (*domain.Cuenta, error) {
	query := `
		SELECT id_cuenta, tipo, id_bodega, email_login, password_hash, fecha_registro
		FROM cuentas WHERE id_cuenta = $1
	`

	cuenta := &domain.Cuenta{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cuenta.ID, &cuenta.Tipo, &cuenta.IDBodega, &cuenta.EmailLogin, &cuenta.PasswordHash, &cuenta.FechaRegistro,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error finding cuenta: %w", err)
	}

	return cuenta, nil
}

func (r *CuentaRepository) FindByEmail(ctx context.Context, email string) (*domain.Cuenta, error) {
	query := `
		SELECT id_cuenta, tipo, id_bodega, email_login, password_hash, fecha_registro
		FROM cuentas WHERE email_login = $1
	`

	cuenta := &domain.Cuenta{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&cuenta.ID, &cuenta.Tipo, &cuenta.IDBodega, &cuenta.EmailLogin, &cuenta.PasswordHash, &cuenta.FechaRegistro,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding cuenta by email: %w", err)
	}

	return cuenta, nil
}

func (r *CuentaRepository) Update(ctx context.Context, tx repository.Transaction, cuenta *domain.Cuenta) error {
	query := `
		UPDATE cuentas
		SET tipo = $1, id_bodega = $2, email_login = $3, password_hash = $4
		WHERE id_cuenta = $5
	`

	_, err := r.db.ExecContext(ctx, query, cuenta.Tipo, cuenta.IDBodega, cuenta.EmailLogin, cuenta.PasswordHash, cuenta.ID)

	if err != nil {
		return fmt.Errorf("error updating cuenta: %w", err)
	}

	return nil
}

func (r *CuentaRepository) Delete(ctx context.Context, tx repository.Transaction, id int) error {
	query := `DELETE FROM cuentas WHERE id_cuenta = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting cuenta: %w", err)
	}

	return nil
}
