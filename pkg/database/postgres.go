package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)


type DB struct {
	*sql.DB
}

// ConnectPostgres establece conexión con PostgreSQL local
func ConnectPostgres(host, port, user, password, dbname string) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=10",
		host, port, user, password, dbname)

	fmt.Printf("Conectando a PostgreSQL local...\n")
	fmt.Printf("   Host: %s:%s\n", host, port)
	fmt.Printf("   Usuario: %s / BD: %s\n", user, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error abriendo conexión: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("error conectando a PostgreSQL: %w", err)
	}

	fmt.Printf("Conexión exitosa a PostgreSQL local\n")
	return &DB{db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}
