package handler

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/middleware"
	"coviar_backend/internal/service"
	"coviar_backend/pkg/audit"
	"coviar_backend/pkg/httputil"
	"coviar_backend/pkg/jwt"
	"coviar_backend/pkg/ratelimit"
	"coviar_backend/pkg/router"
)

type CuentaHandler struct {
	service      *service.CuentaService
	jwtSecret    string
	isProduction bool
	limiter      *ratelimit.Limiter
	audit        *audit.Logger
}

func NewCuentaHandler(service *service.CuentaService, jwtSecret string, isProduction bool, auditLogger *audit.Logger) *CuentaHandler {
	return &CuentaHandler{
		service:      service,
		jwtSecret:    jwtSecret,
		isProduction: isProduction,
		limiter:      ratelimit.New(),
		audit:        auditLogger,
	}
}

func (h *CuentaHandler) Login(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔐 Login request recibido")

	ip := ratelimit.GetIP(r)
	if h.limiter.IsBlocked(ip) {
		log.Printf("🚫 IP bloqueada por exceso de intentos: %s", ip)
		httputil.RespondError(w, http.StatusTooManyRequests, "demasiados intentos fallidos, intentá de nuevo en 15 minutos")
		return
	}

	var req domain.CuentaRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		log.Printf("❌ Error decodificando JSON: %v", err)
		httputil.RespondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	log.Printf("🔐 Intento de login desde IP: %s", ip)

	cuenta, err := h.service.Login(r.Context(), &req)
	if err != nil {
		log.Printf("❌ Error en login: %v", err)
		h.limiter.RecordFailure(ip)
		h.audit.Log(r.Context(), audit.LoginFallido, nil, ip, "")
		httputil.HandleServiceError(w, err)
		return
	}

	log.Printf("✅ Login exitoso para cuenta ID: %d", cuenta.ID)
	h.limiter.RecordSuccess(ip)
	h.audit.Log(r.Context(), audit.LoginExitoso, &cuenta.ID, ip, "")

	bodegaID := 0
	if cuenta.Bodega != nil {
		bodegaID = cuenta.Bodega.ID
	}

	// Generar JWT token (válido por 24 horas)
	accessToken, err := jwt.GenerateToken(
		cuenta.ID,
		cuenta.EmailLogin,
		string(cuenta.Tipo),
		h.jwtSecret,
		24*time.Hour,
		bodegaID,
	)
	if err != nil {
		log.Printf("❌ Error generando access token: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Error generando token")
		return
	}

	// Generar refresh token (válido por 7 días)
	refreshToken, err := jwt.GenerateRefreshToken(
		cuenta.ID,
		cuenta.EmailLogin,
		string(cuenta.Tipo),
		h.jwtSecret,
		bodegaID,
	)
	if err != nil {
		log.Printf("❌ Error generando refresh token: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Error generando refresh token")
		return
	}

	// Establecer cookie de access token
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 horas en segundos
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteLaxMode,
	})

	// Establecer cookie de refresh token
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 días en segundos
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteLaxMode,
	})

	log.Printf("🍪 Cookies establecidas para cuenta ID: %d", cuenta.ID)

	// Responder con datos de la cuenta (sin incluir tokens en JSON)
	httputil.RespondJSON(w, http.StatusOK, cuenta)
}

func (h *CuentaHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	cuenta, err := h.service.GetByIDWithBodega(r.Context(), id)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, cuenta)
}

func (h *CuentaHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if err := h.service.UpdatePassword(r.Context(), id, req.Password); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	ip := ratelimit.GetIP(r)
	idCuenta, _ := r.Context().Value(middleware.UserIDKey).(int)
	h.audit.Log(r.Context(), audit.CambioPassword, &idCuenta, ip, "")

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"mensaje": "Contraseña actualizada"})
}
