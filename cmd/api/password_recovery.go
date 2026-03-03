package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"coviar_backend/pkg/httputil"
	"coviar_backend/pkg/router"

	"golang.org/x/crypto/bcrypt"
)

type passwordResetResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type passwordResetRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

func generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func sendResetEmail(email, token string) error {
	smtpHost := getEnvDefault("SMTP_HOST", "smtp.gmail.com")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	smtpPort := getEnvDefault("SMTP_PORT", "587")
	frontendURL := getEnvDefault("FRONTEND_URL", "http://localhost:3000")

	if smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("configuración SMTP incompleta: SMTP_USER y SMTP_PASSWORD son requeridos")
	}

	resetURL := fmt.Sprintf("%s/actualizar-contrasena?token=%s", frontendURL, token)

	subject := "Subject: Recuperación de Contraseña\r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"

	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
				.wrapper { background-color: #f5f5f5; padding: 40px 20px; }
				.container { max-width: 580px; margin: 0 auto; background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
				.header { background-color: #880D1E; padding: 32px 40px; text-align: center; }
				.header h1 { color: #ffffff; margin: 0; font-size: 22px; font-weight: bold; letter-spacing: 0.5px; }
				.header p { color: rgba(255,255,255,0.8); margin: 6px 0 0; font-size: 13px; }
				.body { padding: 36px 40px; }
				.body p { margin: 0 0 16px; font-size: 15px; color: #444; }
				.button-wrap { text-align: center; margin: 28px 0; }
				.button {
					display: inline-block;
					padding: 14px 36px;
					background-color: #880D1E;
					color: #ffffff !important;
					text-decoration: none;
					border-radius: 6px;
					font-size: 15px;
					font-weight: bold;
					letter-spacing: 0.3px;
				}
				.link-box { background-color: #f8f4f4; border: 1px solid #e8d5d5; border-radius: 4px; padding: 12px 16px; margin: 16px 0; word-break: break-all; font-size: 12px; color: #880D1E; }
				.expiry { font-size: 13px; color: #880D1E; font-weight: bold; text-align: center; margin: 8px 0 24px; }
				.footer { border-top: 1px solid #f0e8e8; padding: 20px 40px; text-align: center; font-size: 12px; color: #999; }
			</style>
		</head>
		<body>
			<div class="wrapper">
				<div class="container">
					<div class="header">
						<h1>Recuperación de Contraseña</h1>
						<p>Corporación Vitivinícola Argentina</p>
					</div>
					<div class="body">
						<p>Has solicitado restablecer la contraseña de tu cuenta en COVIAR.</p>
						<p>Hacé clic en el siguiente botón para crear una nueva contraseña:</p>
						<div class="button-wrap">
							<a href="%s" class="button">Restablecer Contraseña</a>
						</div>
						<p style="font-size:13px; color:#777;">Si el botón no funciona, copiá y pegá este enlace en tu navegador:</p>
						<div class="link-box">%s</div>
						<p class="expiry">⏱ Este enlace expirará en 1 hora.</p>
					</div>
					<div class="footer">
						<p>Si no solicitaste este cambio, podés ignorar este correo.</p>
						<p style="margin-top:8px;">&copy; Corporación Vitivinícola Argentina</p>
					</div>
				</div>
			</div>
		</body>
		</html>
	`, resetURL, resetURL)

	message := []byte(subject + mime + body)
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	if err := smtp.SendMail(addr, auth, smtpUser, []string{email}, message); err != nil {
		return fmt.Errorf("error enviando email: %v", err)
	}

	log.Printf("✅ Email de recuperación enviado a %s", email)
	return nil
}

// RequestPasswordReset maneja POST /api/request-password-reset
func RequestPasswordReset(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req passwordResetRequest
		if err := httputil.DecodeJSON(r, &req); err != nil {
			log.Printf("Error decodificando request: %v", err)
			httputil.RespondJSON(w, http.StatusBadRequest, passwordResetResponse{false, "Datos inválidos"})
			return
		}

		log.Printf("Solicitud de recuperación para email: %s", req.Email)

		// Buscar usuario por email en tabla 'cuentas'
		var userID int
		err := db.QueryRow("SELECT id_cuenta FROM cuentas WHERE email_login = $1", req.Email).Scan(&userID)
		if err != nil {
			httputil.RespondJSON(w, http.StatusNotFound, passwordResetResponse{false, "El correo ingresado no está registrado en el sistema"})
			return
		}

		// Limpiar tokens antiguos del usuario
		if _, err = db.Exec("DELETE FROM restaurar_contrasenas WHERE user_id = $1", userID); err != nil {
			log.Printf("Error al limpiar tokens antiguos: %v", err)
		}

		// Generar token
		token, err := generateResetToken()
		if err != nil {
			log.Printf("Error generando token: %v", err)
			httputil.RespondJSON(w, http.StatusInternalServerError, passwordResetResponse{false, "Error al generar token"})
			return
		}

		expiresAt := time.Now().UTC().Add(1 * time.Hour)

		// Guardar token en BD
		_, err = db.Exec(
			"INSERT INTO restaurar_contrasenas (user_id, token, expires_at, used) VALUES ($1, $2, $3, $4)",
			userID, token, expiresAt, false,
		)
		if err != nil {
			log.Printf("Error al guardar token: %v", err)
			httputil.RespondJSON(w, http.StatusInternalServerError, passwordResetResponse{false, "Error al procesar solicitud"})
			return
		}

		// Enviar email
		if err := sendResetEmail(req.Email, token); err != nil {
			log.Printf("Error enviando email: %v", err)
			httputil.RespondJSON(w, http.StatusInternalServerError, passwordResetResponse{false, "Error al enviar email"})
			return
		}

		httputil.RespondJSON(w, http.StatusOK, passwordResetResponse{true, "Si el email existe, recibirás un correo de recuperación"})
	}
}

// ResetPassword maneja POST /api/reset-password
func ResetPassword(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req resetPasswordRequest
		if err := httputil.DecodeJSON(r, &req); err != nil {
			log.Printf("Error decodificando request: %v", err)
			httputil.RespondJSON(w, http.StatusBadRequest, passwordResetResponse{false, "Datos inválidos"})
			return
		}

		if len(req.NewPassword) < 6 {
			httputil.RespondJSON(w, http.StatusBadRequest, passwordResetResponse{false, "La contraseña debe tener al menos 6 caracteres"})
			return
		}

		var userID int
		var expiresAt time.Time
		var used bool

		// Verificar token
		err := db.QueryRow(
			"SELECT user_id, expires_at, used FROM restaurar_contrasenas WHERE token = $1",
			req.Token,
		).Scan(&userID, &expiresAt, &used)

		if err == sql.ErrNoRows {
			httputil.RespondJSON(w, http.StatusBadRequest, passwordResetResponse{false, "Token inválido"})
			return
		}
		if err != nil {
			log.Printf("Error al verificar token: %v", err)
			httputil.RespondJSON(w, http.StatusInternalServerError, passwordResetResponse{false, "Error al verificar token"})
			return
		}

		if used {
			httputil.RespondJSON(w, http.StatusBadRequest, passwordResetResponse{false, "Este token ya fue utilizado"})
			return
		}

		if time.Now().UTC().After(expiresAt) {
			httputil.RespondJSON(w, http.StatusBadRequest, passwordResetResponse{false, "El token ha expirado"})
			return
		}

		// Hashear nueva contraseña
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hasheando contraseña: %v", err)
			httputil.RespondJSON(w, http.StatusInternalServerError, passwordResetResponse{false, "Error al procesar contraseña"})
			return
		}

		// Actualizar contraseña en tabla 'cuentas'
		if _, err = db.Exec("UPDATE cuentas SET password_hash = $1 WHERE id_cuenta = $2", string(hashedPassword), userID); err != nil {
			log.Printf("Error al actualizar contraseña: %v", err)
			httputil.RespondJSON(w, http.StatusInternalServerError, passwordResetResponse{false, "Error al actualizar contraseña"})
			return
		}

		// Marcar token como usado
		if _, err = db.Exec("UPDATE restaurar_contrasenas SET used = TRUE WHERE token = $1", req.Token); err != nil {
			log.Printf("Error al marcar token como usado: %v", err)
		}

		httputil.RespondJSON(w, http.StatusOK, passwordResetResponse{true, "Contraseña actualizada exitosamente"})
	}
}

type adminCambiarPasswordRequest struct {
	NuevaPassword string `json:"nueva_password"`
}

// AdminCambiarPasswordBodega maneja POST /api/admin/bodegas/{id}/cambiar-password
// Permite al admin establecer directamente una nueva contraseña para la cuenta de una bodega
func AdminCambiarPasswordBodega(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := router.GetParam(r, "id")
		idBodega, err := strconv.Atoi(idStr)
		if err != nil {
			httputil.RespondJSON(w, http.StatusBadRequest, passwordResetResponse{false, "ID de bodega inválido"})
			return
		}

		var req adminCambiarPasswordRequest
		if err := httputil.DecodeJSON(r, &req); err != nil {
			httputil.RespondJSON(w, http.StatusBadRequest, passwordResetResponse{false, "Datos inválidos"})
			return
		}

		if len(req.NuevaPassword) < 6 {
			httputil.RespondJSON(w, http.StatusBadRequest, passwordResetResponse{false, "La contraseña debe tener al menos 6 caracteres"})
			return
		}

		// Buscar cuenta asociada a la bodega
		var userID int
		err = db.QueryRowContext(r.Context(),
			"SELECT id_cuenta FROM cuentas WHERE id_bodega = $1 AND tipo != 'ADMINISTRADOR_APP' LIMIT 1",
			idBodega,
		).Scan(&userID)

		if err == sql.ErrNoRows {
			httputil.RespondJSON(w, http.StatusNotFound, passwordResetResponse{false, "No se encontró cuenta asociada a esta bodega"})
			return
		}
		if err != nil {
			log.Printf("Error al buscar cuenta de bodega %d: %v", idBodega, err)
			httputil.RespondJSON(w, http.StatusInternalServerError, passwordResetResponse{false, "Error al buscar la cuenta"})
			return
		}

		// Hashear nueva contraseña
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NuevaPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hasheando contraseña: %v", err)
			httputil.RespondJSON(w, http.StatusInternalServerError, passwordResetResponse{false, "Error al procesar contraseña"})
			return
		}

		// Actualizar contraseña directamente
		if _, err = db.ExecContext(r.Context(),
			"UPDATE cuentas SET password_hash = $1 WHERE id_cuenta = $2",
			string(hashedPassword), userID,
		); err != nil {
			log.Printf("Error al actualizar contraseña de bodega %d: %v", idBodega, err)
			httputil.RespondJSON(w, http.StatusInternalServerError, passwordResetResponse{false, "Error al actualizar contraseña"})
			return
		}

		log.Printf("✅ Contraseña cambiada por admin para bodega %d (cuenta %d)", idBodega, userID)
		httputil.RespondJSON(w, http.StatusOK, passwordResetResponse{true, "Contraseña actualizada correctamente"})
	}
}

// cleanExpiredTokens se ejecuta en background y elimina tokens expirados o ya usados cada hora
func cleanExpiredTokens(db *sql.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		result, err := db.Exec("DELETE FROM restaurar_contrasenas WHERE expires_at < NOW() OR used = TRUE")
		if err != nil {
			log.Printf("Error limpiando tokens expirados: %v", err)
			continue
		}
		if rows, _ := result.RowsAffected(); rows > 0 {
			log.Printf("Tokens expirados eliminados: %d", rows)
		}
	}
}

func getEnvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
