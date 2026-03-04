package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"coviar_backend/pkg/jwt"
)

// ContextKey es el tipo para claves de contexto
type ContextKey string

const (
	// UserIDKey es la clave para almacenar el ID de usuario en el contexto
	UserIDKey ContextKey = "user_id"
	// UserEmailKey es la clave para almacenar el email del usuario en el contexto
	UserEmailKey ContextKey = "user_email"
	// UserTipoKey es la clave para almacenar el tipo de cuenta en el contexto
	UserTipoKey ContextKey = "user_tipo"
)

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
		frontendURL := os.Getenv("FRONTEND_URL")
		allowed := []string{"http://localhost:3000", "http://localhost:8080"}
		if frontendURL != "" {
			allowed = append(allowed, frontendURL)
		}
		for _, o := range allowed {
			if origin == o {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		// IMPORTANTE: Requerido para cookies
		w.Header().Set("Access-Control-Allow-Credentials", "true")

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

// AuthMiddleware verifica el JWT token desde la cookie
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Obtener cookie de auth_token
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				log.Printf("❌ No se encontró cookie auth_token: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"no autenticado"}`))
				return
			}

			// Validar token
			claims, err := jwt.ValidateToken(cookie.Value, jwtSecret)
			if err != nil {
				log.Printf("❌ Token inválido: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"token inválido o expirado"}`))
				return
			}

			// Agregar claims al contexto
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserTipoKey, claims.TipoCuenta)

			log.Printf("✅ Usuario autenticado: ID=%d, Email=%s", claims.UserID, claims.Email)

			// Continuar con la petición
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
