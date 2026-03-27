package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"coviar_backend/pkg/httputil"
	"coviar_backend/pkg/ratelimit"
)

// reenvioDelays define la espera mínima entre reenvíos consecutivos.
// El índice representa el número de reenvío ya realizado (1-based).
//
//	Tras el 1er reenvío  → esperar 1 min
//	Tras el 2do          → esperar 3 min
//	Tras el 3ro          → esperar 5 min
//	Tras el 4to          → esperar 10 min
//	Tras el 5to          → esperar 30 min
//	6to intento          → bloqueado 24 horas
var reenvioDelays = []time.Duration{
	1 * time.Minute,
	3 * time.Minute,
	5 * time.Minute,
	10 * time.Minute,
	30 * time.Minute,
}

const bloqueoReenvio = 24 * time.Hour
const maxIntentosFallidos = 20

// Limiter de IP como primera línea de defensa contra bots
var verificacionLimiter = ratelimit.New()

// ─── DTOs ────────────────────────────────────────────────────────────────────

type verificarCorreoRequest struct {
	Email  string `json:"email"`
	Codigo string `json:"codigo"`
}

type verificarCorreoResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type reenviarCodigoRequest struct {
	Email string `json:"email"`
}

// reenvioResponse incluye segundos_espera para que el frontend
// pueda ajustar su temporizador dinámicamente.
type reenvioResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	SegsEspera int    `json:"segundos_espera"`
}

// ─── Utilidades ──────────────────────────────────────────────────────────────

func generateVerificationCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%06d", r.Intn(1_000_000))
}

// formatDuracion convierte una duración a texto en español para mensajes de error.
func formatDuracion(d time.Duration) string {
	if d >= time.Hour {
		return fmt.Sprintf("%d hora(s)", int(d.Hours()))
	}
	if d >= time.Minute {
		return fmt.Sprintf("%d minuto(s)", int(d.Minutes()))
	}
	return fmt.Sprintf("%d segundo(s)", int(d.Seconds()))
}

func sendVerificationEmail(email, codigo string) error {
	smtpHost := getEnvDefault("SMTP_HOST", "smtp.gmail.com")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	smtpPort := getEnvDefault("SMTP_PORT", "587")

	if smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("configuración SMTP incompleta: SMTP_USER y SMTP_PASSWORD son requeridos")
	}

	subject := "Subject: Verificación de correo - COVIAR\r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
    .wrapper { background-color: #f5f5f5; padding: 40px 20px; }
    .container { max-width: 560px; margin: 0 auto; background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
    .header { background-color: #880D1E; padding: 32px 40px; text-align: center; }
    .header h1 { color: #ffffff; margin: 0; font-size: 22px; font-weight: bold; }
    .header p { color: rgba(255,255,255,0.8); margin: 6px 0 0; font-size: 13px; }
    .body { padding: 36px 40px; }
    .body p { margin: 0 0 16px; font-size: 15px; color: #444; }
    .code-box { text-align: center; background-color: #f8f4f4; border: 2px solid #880D1E; border-radius: 8px; padding: 24px; margin: 24px 0; }
    .code { font-size: 42px; font-weight: bold; letter-spacing: 12px; color: #880D1E; font-family: monospace; }
    .expiry { font-size: 13px; color: #880D1E; font-weight: bold; text-align: center; margin: 8px 0 24px; }
    .footer { border-top: 1px solid #f0e8e8; padding: 20px 40px; text-align: center; font-size: 12px; color: #999; }
  </style>
</head>
<body>
  <div class="wrapper">
    <div class="container">
      <div class="header">
        <h1>Verificación de Correo</h1>
        <p>Corporación Vitivinícola Argentina</p>
      </div>
      <div class="body">
        <p>Ingresá el siguiente código de verificación para activar tu cuenta:</p>
        <div class="code-box">
          <div class="code">%s</div>
        </div>
        <p class="expiry">Este código expira en 5 minutos.</p>
        <p style="font-size:13px; color:#777;">Si no creaste una cuenta en COVIAR, podés ignorar este correo.</p>
      </div>
      <div class="footer">
        <p>&copy; Corporación Vitivinícola Argentina</p>
      </div>
    </div>
  </div>
</body>
</html>`, codigo)

	message := []byte(subject + mime + body)
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	if err := smtp.SendMail(addr, auth, smtpUser, []string{email}, message); err != nil {
		return fmt.Errorf("error enviando email de verificación: %w", err)
	}

	log.Printf("[EMAIL] Verificación enviada a %s", email)
	return nil
}

// SendVerificationCode genera un código, lo persiste y lo envía por email.
// Es exportada para ser invocada desde el hook del handler de registro.
func SendVerificationCode(db *sql.DB, cuentaID int, email string) error {
	if _, err := db.Exec("DELETE FROM verificacion_correo WHERE cuenta_id = $1", cuentaID); err != nil {
		log.Printf("[WARN] No se pudieron limpiar códigos anteriores para cuenta %d: %v", cuentaID, err)
	}

	codigo := generateVerificationCode()
	expiresAt := time.Now().UTC().Add(5 * time.Minute)

	if _, err := db.Exec(
		"INSERT INTO verificacion_correo (cuenta_id, codigo, expires_at) VALUES ($1, $2, $3)",
		cuentaID, codigo, expiresAt,
	); err != nil {
		return fmt.Errorf("error guardando código de verificación: %w", err)
	}

	return sendVerificationEmail(email, codigo)
}

// ─── Handlers ────────────────────────────────────────────────────────────────

// VerificarCorreo maneja POST /api/verificar-correo
func VerificarCorreo(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := ratelimit.GetIP(r)
		if verificacionLimiter.IsBlocked(ip) {
			log.Printf("[RATE LIMIT] IP bloqueada en verificación de correo: %s", ip)
			httputil.RespondJSON(w, http.StatusTooManyRequests, verificarCorreoResponse{
				false, "Demasiados intentos, esperá unos minutos e intentá nuevamente",
			})
			return
		}

		var req verificarCorreoRequest
		if err := httputil.DecodeJSON(r, &req); err != nil {
			httputil.RespondJSON(w, http.StatusBadRequest, verificarCorreoResponse{false, "Datos inválidos"})
			return
		}

		if req.Email == "" || len(req.Codigo) != 6 {
			httputil.RespondJSON(w, http.StatusBadRequest, verificarCorreoResponse{false, "Email y código son requeridos"})
			return
		}

		var cuentaID int
		var emailVerificado bool
		err := db.QueryRowContext(r.Context(),
			"SELECT id_cuenta, email_verificado FROM cuentas WHERE email_login = $1",
			req.Email,
		).Scan(&cuentaID, &emailVerificado)

		if err == sql.ErrNoRows {
			verificacionLimiter.RecordFailure(ip)
			httputil.RespondJSON(w, http.StatusBadRequest, verificarCorreoResponse{false, "Código inválido o expirado"})
			return
		}
		if err != nil {
			log.Printf("Error al buscar cuenta para verificación: %v", err)
			httputil.RespondJSON(w, http.StatusInternalServerError, verificarCorreoResponse{false, "Error interno"})
			return
		}

		if emailVerificado {
			httputil.RespondJSON(w, http.StatusOK, verificarCorreoResponse{true, "Correo ya verificado"})
			return
		}

		var dbCodigo string
		var expiresAt time.Time
		var yaVerificado bool
		var intentosFallidos int

		err = db.QueryRowContext(r.Context(),
			`SELECT codigo, expires_at, verificado, intentos_fallidos
			   FROM verificacion_correo
			  WHERE cuenta_id = $1
			  ORDER BY created_at DESC
			  LIMIT 1`,
			cuentaID,
		).Scan(&dbCodigo, &expiresAt, &yaVerificado, &intentosFallidos)

		if err == sql.ErrNoRows {
			verificacionLimiter.RecordFailure(ip)
			httputil.RespondJSON(w, http.StatusBadRequest, verificarCorreoResponse{
				false, "El código ingresado es inválido o ha expirado. Por favor, verifique e intente nuevamente.",
			})
			return
		}
		if err != nil {
			log.Printf("Error al verificar código: %v", err)
			httputil.RespondJSON(w, http.StatusInternalServerError, verificarCorreoResponse{false, "Error interno"})
			return
		}

		// Verificar límite de intentos fallidos por código
		if intentosFallidos >= maxIntentosFallidos {
			httputil.RespondJSON(w, http.StatusTooManyRequests, verificarCorreoResponse{
				false, "Demasiados intentos fallidos. Por favor, solicitá un nuevo código.",
			})
			return
		}

		// Validar código
		if yaVerificado || time.Now().UTC().After(expiresAt) || dbCodigo != req.Codigo {
			verificacionLimiter.RecordFailure(ip)
			// Incrementar intentos fallidos en BD (subquery necesario en PostgreSQL)
			db.ExecContext(r.Context(),
				`UPDATE verificacion_correo
				    SET intentos_fallidos = intentos_fallidos + 1
				  WHERE id = (
				      SELECT id FROM verificacion_correo
				       WHERE cuenta_id = $1 AND verificado = FALSE
				       ORDER BY created_at DESC
				       LIMIT 1
				  )`,
				cuentaID,
			)
			restantes := maxIntentosFallidos - (intentosFallidos + 1)
			var msg string
			if restantes <= 0 {
				msg = "Demasiados intentos fallidos. Por favor, solicitá un nuevo código."
			} else {
				msg = fmt.Sprintf("Código inválido o expirado. Te quedan %d intento(s).", restantes)
			}
			httputil.RespondJSON(w, http.StatusBadRequest, verificarCorreoResponse{false, msg})
			return
		}

		// Marcar código y cuenta como verificados
		db.ExecContext(r.Context(),
			"UPDATE verificacion_correo SET verificado = TRUE WHERE cuenta_id = $1 AND codigo = $2",
			cuentaID, req.Codigo,
		)
		db.ExecContext(r.Context(),
			"UPDATE cuentas SET email_verificado = TRUE WHERE id_cuenta = $1",
			cuentaID,
		)

		// Limpiar el historial de reenvíos al verificar exitosamente
		db.ExecContext(r.Context(),
			"DELETE FROM verificacion_reenvios WHERE cuenta_id = $1",
			cuentaID,
		)

		log.Printf("[VERIFICACION] Correo verificado para cuenta %d (%s)", cuentaID, req.Email)
		httputil.RespondJSON(w, http.StatusOK, verificarCorreoResponse{true, "Correo verificado correctamente"})
	}
}

// ReenviarCodigoVerificacion maneja POST /api/reenviar-codigo-verificacion
//
// Aplica esperas progresivas entre reenvíos y bloquea la cuenta 24 h
// tras superar el límite de intentos. La respuesta incluye segundos_espera
// para que el frontend ajuste su temporizador dinámicamente.
func ReenviarCodigoVerificacion(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := ratelimit.GetIP(r)
		if verificacionLimiter.IsBlocked(ip) {
			log.Printf("[RATE LIMIT] IP bloqueada en reenvío de código: %s", ip)
			httputil.RespondJSON(w, http.StatusTooManyRequests, reenvioResponse{
				false, "Demasiadas solicitudes, intentá de nuevo más tarde", 0,
			})
			return
		}

		var req reenviarCodigoRequest
		if err := httputil.DecodeJSON(r, &req); err != nil {
			httputil.RespondJSON(w, http.StatusBadRequest, reenvioResponse{false, "Datos inválidos", 0})
			return
		}
		if req.Email == "" {
			httputil.RespondJSON(w, http.StatusBadRequest, reenvioResponse{false, "Email es requerido", 0})
			return
		}

		// Buscar cuenta
		var cuentaID int
		var emailVerificado bool
		err := db.QueryRowContext(r.Context(),
			"SELECT id_cuenta, email_verificado FROM cuentas WHERE email_login = $1",
			req.Email,
		).Scan(&cuentaID, &emailVerificado)

		if err != nil {
			// Email no encontrado: penalizar IP y responder genérico (anti-enumeración)
			verificacionLimiter.RecordFailure(ip)
			httputil.RespondJSON(w, http.StatusOK, reenvioResponse{
				true, "Si el email está registrado y pendiente de verificación, recibirás un nuevo código", 0,
			})
			return
		}

		if emailVerificado {
			httputil.RespondJSON(w, http.StatusOK, reenvioResponse{true, "El correo ya fue verificado", 0})
			return
		}

		// Consultar estado de reenvíos en BD
		now := time.Now().UTC()
		var intentos int
		var proximoReenvio sql.NullTime
		var bloqueadoHasta sql.NullTime

		err = db.QueryRowContext(r.Context(),
			"SELECT intentos, proximo_reenvio, bloqueado_hasta FROM verificacion_reenvios WHERE cuenta_id = $1",
			cuentaID,
		).Scan(&intentos, &proximoReenvio, &bloqueadoHasta)

		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error consultando verificacion_reenvios para cuenta %d: %v", cuentaID, err)
			httputil.RespondJSON(w, http.StatusInternalServerError, reenvioResponse{false, "Error interno", 0})
			return
		}

		// ── Verificar bloqueo de 24 h ────────────────────────────────────────
		if bloqueadoHasta.Valid && now.Before(bloqueadoHasta.Time) {
			secsLeft := int(bloqueadoHasta.Time.Sub(now).Seconds())
			httputil.RespondJSON(w, http.StatusTooManyRequests, reenvioResponse{
				false,
				"Tu cuenta está bloqueada por demasiados intentos de reenvío. Intentá nuevamente en 24 horas.",
				secsLeft,
			})
			return
		}

		// ── Verificar cooldown entre reenvíos ────────────────────────────────
		if proximoReenvio.Valid && now.Before(proximoReenvio.Time) {
			remaining := proximoReenvio.Time.Sub(now)
			secsLeft := int(remaining.Seconds())
			msg := fmt.Sprintf("Debés esperar %s antes de reenviar el código.", formatDuracion(remaining))
			httputil.RespondJSON(w, http.StatusTooManyRequests, reenvioResponse{false, msg, secsLeft})
			return
		}

		// ── Calcular próximo estado ───────────────────────────────────────────
		newIntentos := intentos + 1

		if newIntentos > len(reenvioDelays) {
			// Supera el límite → bloquear 24 h, NO enviar código
			bloqueadoHastaTime := now.Add(bloqueoReenvio)
			db.ExecContext(r.Context(),
				`INSERT INTO verificacion_reenvios (cuenta_id, intentos, proximo_reenvio, bloqueado_hasta)
				 VALUES ($1, $2, NULL, $3)
				 ON CONFLICT (cuenta_id) DO UPDATE
				 SET intentos        = EXCLUDED.intentos,
				     proximo_reenvio = NULL,
				     bloqueado_hasta = EXCLUDED.bloqueado_hasta`,
				cuentaID, newIntentos, bloqueadoHastaTime,
			)
			log.Printf("[BLOQUEO] Cuenta %d bloqueada 24 h por exceso de reenvíos (intento %d)", cuentaID, newIntentos)
			httputil.RespondJSON(w, http.StatusTooManyRequests, reenvioResponse{
				false,
				"Demasiados intentos de reenvío. Tu cuenta está bloqueada por 24 horas.",
				int(bloqueoReenvio.Seconds()),
			})
			return
		}

		// Dentro del límite → calcular próxima espera y actualizar
		nextDelay := reenvioDelays[newIntentos-1]
		nextAllowed := now.Add(nextDelay)

		if _, err = db.ExecContext(r.Context(),
			`INSERT INTO verificacion_reenvios (cuenta_id, intentos, proximo_reenvio, bloqueado_hasta)
			 VALUES ($1, $2, $3, NULL)
			 ON CONFLICT (cuenta_id) DO UPDATE
			 SET intentos        = EXCLUDED.intentos,
			     proximo_reenvio = EXCLUDED.proximo_reenvio,
			     bloqueado_hasta = NULL`,
			cuentaID, newIntentos, nextAllowed,
		); err != nil {
			log.Printf("Error actualizando verificacion_reenvios para cuenta %d: %v", cuentaID, err)
			httputil.RespondJSON(w, http.StatusInternalServerError, reenvioResponse{false, "Error interno", 0})
			return
		}

		// Enviar el código
		if err := SendVerificationCode(db, cuentaID, req.Email); err != nil {
			log.Printf("Error enviando código a %s: %v", req.Email, err)
			httputil.RespondJSON(w, http.StatusInternalServerError, reenvioResponse{
				false, "No se pudo reenviar el código. Intentá nuevamente.", 0,
			})
			return
		}

		segsEspera := int(nextDelay.Seconds())
		log.Printf("[REENVIO] Código enviado a %s (intento %d/%d, próximo en %s)",
			req.Email, newIntentos, len(reenvioDelays), nextDelay)
		httputil.RespondJSON(w, http.StatusOK, reenvioResponse{
			true, "Código reenviado correctamente", segsEspera,
		})
	}
}
