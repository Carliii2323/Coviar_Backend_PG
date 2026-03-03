package postgres

import (
	"context"
	"database/sql"

	"coviar_backend/internal/repository"
)

// PostgresTransaction implementa la interfaz Transaction
type PostgresTransaction struct {
	tx *sql.Tx
}

func (pt *PostgresTransaction) Commit() error {
	return pt.tx.Commit()
}

func (pt *PostgresTransaction) Rollback() error {
	return pt.tx.Rollback()
}

// TransactionManager para PostgreSQL
type TransactionManager struct {
	db *sql.DB
}

func NewTransactionManager(db *sql.DB) repository.TransactionManager {
	return &TransactionManager{db: db}
}

func (tm *TransactionManager) BeginTx(ctx context.Context) (repository.Transaction, error) {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &PostgresTransaction{tx: tx}, nil
}
