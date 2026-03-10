package audit

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// Accion representa el tipo de evento auditado.
type Accion string

const (
	LoginExitoso            Accion = "LOGIN_EXITOSO"
	LoginFallido            Accion = "LOGIN_FALLIDO"
	Logout                  Accion = "LOGOUT"
	CambioPassword          Accion = "CAMBIO_PASSWORD"
	AdminCambioPassword     Accion = "ADMIN_CAMBIO_PASSWORD"
	CrearResponsable        Accion = "CREAR_RESPONSABLE"
	BajaResponsable         Accion = "BAJA_RESPONSABLE"
	CrearAutoevaluacion     Accion = "CREAR_AUTOEVALUACION"
	CompletarAutoevaluacion Accion = "COMPLETAR_AUTOEVALUACION"
	CancelarAutoevaluacion  Accion = "CANCELAR_AUTOEVALUACION"
)

// Logger escribe eventos de auditoría en la base de datos.
type Logger struct {
	db *sql.DB
}

func New(db *sql.DB) *Logger {
	return &Logger{db: db}
}

// CreateTable crea la tabla audit_log si no existe. Llamar una vez al inicio.
func CreateTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_log (
			id         BIGSERIAL PRIMARY KEY,
			timestamp  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			accion     VARCHAR(50) NOT NULL,
			id_cuenta  INTEGER,
			ip         VARCHAR(45),
			detalle    TEXT
		)
	`)
	return err
}

// Log registra un evento de auditoría. No bloquea: si falla, loguea el error internamente.
func (l *Logger) Log(ctx context.Context, accion Accion, idCuenta *int, ip, detalle string) {
	_, err := l.db.ExecContext(ctx,
		`INSERT INTO audit_log (accion, id_cuenta, ip, detalle, timestamp) VALUES ($1, $2, $3, $4, $5)`,
		string(accion), idCuenta, ip, detalle, time.Now().UTC(),
	)
	if err != nil {
		log.Printf("[AUDIT ERROR] No se pudo registrar acción %s: %v", accion, err)
	}
}
