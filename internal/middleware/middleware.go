package middleware

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"coviar_backend/pkg/jwt"
	"coviar_backend/pkg/ratelimit"
	"coviar_backend/pkg/tokenblacklist"
)

var (
	// 200 requests/min por IP para cualquier endpoint
	globalLimiter = ratelimit.NewWindowLimiter(200, time.Minute)
	// 20 requests/min por IP para operaciones de escritura (POST/PUT/DELETE)
	writeLimiter = ratelimit.NewWindowLimiter(20, time.Minute)
)

// ContextKey es el tipo para claves de contexto
type ContextKey string

const (
	UserIDKey    ContextKey = "user_id"
	UserEmailKey ContextKey = "user_email"
	UserTipoKey  ContextKey = "user_tipo"
	BodegaIDKey  ContextKey = "bodega_id"
)

// allowedOrigins devuelve la lista de orígenes permitidos según configuración.
func allowedOrigins() []string {
	origins := []string{"http://localhost:3000", "http://localhost:8080"}
	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		origins = append(origins, frontendURL)
	}
	return origins
}

// CSRFProtect verifica que las peticiones mutantes (POST/PUT/DELETE/PATCH)
// provengan de un origen permitido, usando el header Origin o Referer como respaldo.
// Es una defensa en profundidad sobre SameSite=Lax.
func CSRFProtect(next http.Handler) http.Handler {
	mutating := map[string]bool{
		http.MethodPost:   true,
		http.MethodPut:    true,
		http.MethodDelete: true,
		http.MethodPatch:  true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !mutating[r.Method] {
			next.ServeHTTP(w, r)
			return
		}

		allowed := allowedOrigins()

		// Verificar header Origin primero
		if origin := r.Header.Get("Origin"); origin != "" {
			for _, o := range allowed {
				if origin == o {
					next.ServeHTTP(w, r)
					return
				}
			}
			log.Printf("[CSRF] Bloqueado — Origin no permitido: %s", origin)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"origen no permitido"}`))
			return
		}

		// Sin Origin: verificar Referer como respaldo
		if referer := r.Header.Get("Referer"); referer != "" {
			parsed, err := url.Parse(referer)
			if err != nil {
				log.Printf("[CSRF] Bloqueado — Referer inválido: %s", referer)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error":"origen no permitido"}`))
				return
			}
			refOrigin := strings.TrimRight(parsed.Scheme+"://"+parsed.Host, "/")
			for _, o := range allowed {
				if refOrigin == o {
					next.ServeHTTP(w, r)
					return
				}
			}
			log.Printf("[CSRF] Bloqueado — Referer no permitido: %s", refOrigin)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"origen no permitido"}`))
			return
		}

		// Sin Origin ni Referer: petición directa (curl, server-to-server, mismo origen)
		// Se permite para no bloquear herramientas legítimas y tests
		next.ServeHTTP(w, r)
	})
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[%s] %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s - completado en %v", r.Method, r.RequestURI, time.Since(start))
	})
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, o := range allowedOrigins() {
			if origin == o {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				// Solo orígenes permitidos reciben credenciales (cookies)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				break
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cookie")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAdmin verifica que el usuario autenticado sea ADMINISTRADOR_APP.
// Debe usarse después de AuthMiddleware.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tipo, _ := r.Context().Value(UserTipoKey).(string)
		if tipo != "ADMINISTRADOR_APP" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"acceso restringido a administradores"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RateLimit aplica límites de requests por IP: 200/min global, 20/min para escrituras.
func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := ratelimit.GetIP(r)

		if !globalLimiter.Allow(ip) {
			log.Printf("[RATE LIMIT] Global excedido por IP: %s", ip)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"demasiadas solicitudes, intentá de nuevo en un minuto"}`))
			return
		}

		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			if !writeLimiter.Allow(ip) {
				log.Printf("[RATE LIMIT] Escritura excedida por IP: %s", ip)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"demasiadas operaciones de escritura, intentá de nuevo en un minuto"}`))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"error interno del servidor"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware verifica el JWT token desde la cookie y que no haya sido revocado.
func AuthMiddleware(jwtSecret string, bl *tokenblacklist.Blacklist) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Obtener cookie de auth_token
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				log.Printf("[AUTH] Cookie auth_token no encontrada: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"no autenticado"}`))
				return
			}

			// Verificar si el token fue revocado (logout previo)
			if bl.IsRevoked(cookie.Value) {
				log.Printf("[AUTH] Token revocado (sesión cerrada)")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"sesión cerrada, iniciá sesión nuevamente"}`))
				return
			}

			// Verificar inactividad: si pasó más de 1h sin requests, cerrar sesión
			if bl.IsInactive(cookie.Value) {
				log.Printf("[AUTH] Sesión expirada por inactividad")
				bl.Revoke(cookie.Value, time.Now().Add(24*time.Hour)) // revocar para limpiar
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"sesión expirada por inactividad, iniciá sesión nuevamente"}`))
				return
			}

			// Validar token
			claims, err := jwt.ValidateToken(cookie.Value, jwtSecret)
			if err != nil {
				log.Printf("[AUTH] Token inválido: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"token inválido o expirado"}`))
				return
			}

			// Agregar claims al contexto
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserTipoKey, claims.TipoCuenta)
			ctx = context.WithValue(ctx, BodegaIDKey, claims.BodegaID)

			log.Printf("[AUTH] Usuario autenticado: ID=%d, Email=%s", claims.UserID, claims.Email)

			// Actualizar timestamp de actividad (reinicia el contador de inactividad)
			bl.RecordActivity(cookie.Value)

			// Continuar con la petición
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
