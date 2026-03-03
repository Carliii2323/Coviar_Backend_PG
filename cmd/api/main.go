package main

import (
	"log"
	"net/http"

	"coviar_backend/internal/handler"
	"coviar_backend/internal/middleware"
	"coviar_backend/internal/repository/postgres"
	"coviar_backend/internal/service"
	"coviar_backend/pkg/config"
	"coviar_backend/pkg/database"
	"coviar_backend/pkg/httputil"
	"coviar_backend/pkg/router"
)

func main() {
	// 1. Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Error cargando configuración: %v", err)
	}
	log.Println("✓ Configuración cargada")

	// 2. Conectar a PostgreSQL local
	db, err := database.ConnectPostgres(cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name)
	if err != nil {
		log.Fatalf("❌ Error conectando a PostgreSQL: %v", err)
	}
	defer db.Close()
	log.Println("✓ Conexión a PostgreSQL local establecida")

	// 3. Inicializar repositorios
	bodegaRepo := postgres.NewBodegaRepository(db.DB)
	cuentaRepo := postgres.NewCuentaRepository(db.DB)
	responsableRepo := postgres.NewResponsableRepository(db.DB)
	ubicacionRepo := postgres.NewUbicacionRepository(db.DB)
	autoevaluacionRepo := postgres.NewAutoevaluacionRepository(db.DB)
	segmentoRepo := postgres.NewSegmentoRepository(db.DB)
	capituloRepo := postgres.NewCapituloRepository(db.DB)
	indicadorRepo := postgres.NewIndicadorRepository(db.DB)
	nivelRespuestaRepo := postgres.NewNivelRespuestaRepository(db.DB)
	respuestaRepo := postgres.NewRespuestaRepository(db.DB)
	txManager := postgres.NewTransactionManager(db.DB)
	evidenciaRepo := postgres.NewEvidenciaRepository(db.DB)
	adminRepo := postgres.NewAdminRepository(db.DB)

	log.Println("✓ Repositorios inicializados")

	// 4. Inicializar servicios
	registroService := service.NewRegistroService(bodegaRepo, cuentaRepo, responsableRepo, txManager)
	ubicacionService := service.NewUbicacionService(ubicacionRepo)
	cuentaService := service.NewCuentaService(cuentaRepo, bodegaRepo)
	bodegaService := service.NewBodegaService(bodegaRepo)
	responsableService := service.NewResponsableService(responsableRepo, cuentaRepo, autoevaluacionRepo)
	autoevaluacionService := service.NewAutoevaluacionService(autoevaluacionRepo, segmentoRepo, capituloRepo, indicadorRepo, nivelRespuestaRepo, respuestaRepo, evidenciaRepo, responsableRepo)
	evidenciaService := service.NewEvidenciaService(evidenciaRepo, respuestaRepo, autoevaluacionRepo, bodegaRepo, indicadorRepo)
	adminService := service.NewAdminService(adminRepo)

	log.Println("✓ Servicios inicializados")

	// 5. Inicializar handlers (con JWT secret para autenticación)
	registroHandler := handler.NewRegistroHandler(registroService)
	ubicacionHandler := handler.NewUbicacionHandler(ubicacionService)
	cuentaHandler := handler.NewCuentaHandler(cuentaService, cfg.JWT.Secret)
	bodegaHandler := handler.NewBodegaHandler(bodegaService, autoevaluacionService)
	responsableHandler := handler.NewResponsableHandler(responsableService)
	autoevaluacionHandler := handler.NewAutoevaluacionHandler(autoevaluacionService)
	evidenciaHandler := handler.NewEvidenciaHandler(evidenciaService)
	adminHandler := handler.NewAdminHandler(adminService)

	log.Println("✓ Handlers inicializados")

	// 6. Configurar router
	r := router.New()

	// Middlewares globales
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS)

	// ===== RUTAS PÚBLICAS =====

	// Registro y autenticación (no requieren autenticación)
	r.POST("/api/registro", registroHandler.RegistrarBodega)
	r.POST("/api/login", cuentaHandler.Login)

	// Logout (elimina cookies)
	r.POST("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🔓 Logout request recibido")

		// Eliminar cookie auth_token
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1, // Eliminar cookie
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		// Eliminar cookie refresh_token
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1, // Eliminar cookie
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		log.Printf("✅ Cookies eliminadas")
		httputil.RespondJSON(w, http.StatusOK, map[string]string{
			"mensaje": "Logout exitoso",
		})
	})

	// Ubicaciones (públicas - necesarias para registro)
	r.GET("/api/provincias", ubicacionHandler.GetProvincias)
	r.GET("/api/departamentos", ubicacionHandler.GetDepartamentos)
	r.GET("/api/localidades", ubicacionHandler.GetLocalidades)

	// Recuperación de contraseña (públicas)
	r.POST("/api/recuperar-password", RequestPasswordReset(db.DB))
	r.POST("/api/restablecer-password", ResetPassword(db.DB))

	// Iniciar limpieza de tokens expirados en background
	go cleanExpiredTokens(db.DB)

	// Health check
	r.GET("/health", func(w http.ResponseWriter, r *http.Request) {
		httputil.RespondJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"version": "2.0.0",
			"message": "Coviar Backend - Integrado y Funcional con JWT",
		})
	})

	// ===== RUTAS PROTEGIDAS (requieren autenticación) =====

	authMiddleware := middleware.AuthMiddleware(cfg.JWT.Secret)

	// Helper para convertir http.Handler a http.HandlerFunc
	protect := func(handler http.HandlerFunc) http.HandlerFunc {
		return authMiddleware(handler).ServeHTTP
	}

	// Cuentas (protegidas)
	r.GET("/api/cuentas/{id}", protect(cuentaHandler.GetByID))
	r.PUT("/api/cuentas/{id}", protect(cuentaHandler.UpdatePassword))

	// Bodegas (protegidas) - static routes BEFORE dynamic
	r.GET("/api/bodegas/{id}/resultados-autoevaluacion", protect(bodegaHandler.GetResultadosAutoevaluacion))
	r.GET("/api/bodegas/{id}", protect(bodegaHandler.GetByID))
	r.PUT("/api/bodegas/{id}", protect(bodegaHandler.Update))

	// Responsables (protegidas)
	r.GET("/api/responsables/{id}", protect(responsableHandler.GetByID))
	r.PUT("/api/responsables/{id}", protect(responsableHandler.Update))
	r.POST("/api/responsables/{id}/baja", protect(responsableHandler.DarDeBaja))
	r.GET("/api/cuentas/{cuenta_id}/responsables", protect(responsableHandler.GetByCuentaID))
	r.POST("/api/cuentas/{cuenta_id}/responsables", protect(responsableHandler.Create))

	// Admin (protegidas)
	r.GET("/api/admin/stats", protect(adminHandler.GetStats))
	r.GET("/api/admin/evaluaciones", protect(adminHandler.GetAllEvaluaciones))
	r.GET("/api/admin/bodegas", protect(bodegaHandler.GetAll))
	r.POST("/api/admin/bodegas/{id}/cambiar-password", protect(AdminCambiarPasswordBodega(db.DB)))

	// Autoevaluaciones (protegidas) - static routes BEFORE dynamic
	r.GET("/api/autoevaluaciones/historial", protect(autoevaluacionHandler.GetHistorialAutoevaluaciones))
	r.POST("/api/autoevaluaciones", protect(autoevaluacionHandler.CreateAutoevaluacion))
	r.GET("/api/autoevaluaciones/{id_autoevaluacion}/resultados", protect(autoevaluacionHandler.GetResultadosAutoevaluacion))
	r.GET("/api/autoevaluaciones/{id_autoevaluacion}/segmentos", protect(autoevaluacionHandler.GetSegmentos))
	r.PUT("/api/autoevaluaciones/{id_autoevaluacion}/segmento", protect(autoevaluacionHandler.SeleccionarSegmento))
	r.GET("/api/autoevaluaciones/{id_autoevaluacion}/estructura", protect(autoevaluacionHandler.GetEstructura))
	r.POST("/api/autoevaluaciones/{id_autoevaluacion}/respuestas", protect(autoevaluacionHandler.GuardarRespuestas))
	r.POST("/api/autoevaluaciones/{id_autoevaluacion}/completar", protect(autoevaluacionHandler.CompletarAutoevaluacion))
	r.POST("/api/autoevaluaciones/{id_autoevaluacion}/cancelar", protect(autoevaluacionHandler.CancelarAutoevaluacion))
	r.POST("/api/autoevaluaciones/{id_autoevaluacion}/respuestas/{id_respuesta}/evidencias", protect(evidenciaHandler.AgregarEvidencia))
	r.GET("/api/autoevaluaciones/{id_autoevaluacion}/respuestas/{id_respuesta}/evidencia", protect(evidenciaHandler.ObtenerEvidencia))
	r.GET("/api/autoevaluaciones/{id_autoevaluacion}/evidencias", protect(evidenciaHandler.ObtenerEvidenciasPorAutoevaluacion))
	r.GET("/api/autoevaluaciones/{id_autoevaluacion}/respuestas/{id_respuesta}/evidencia/descargar", protect(evidenciaHandler.DescargarEvidencia))
	r.GET("/api/autoevaluaciones/{id_autoevaluacion}/evidencias/descargar", protect(evidenciaHandler.DescargarTodasEvidencias))
	r.DELETE("/api/autoevaluaciones/{id_autoevaluacion}/respuestas/{id_respuesta}/evidencia", protect(evidenciaHandler.EliminarEvidencia))
	r.PUT("/api/autoevaluaciones/{id_autoevaluacion}/respuestas/{id_respuesta}/evidencia", protect(evidenciaHandler.CambiarEvidencia))

	// 7. Iniciar servidor
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("🚀 Servidor iniciando en http://%s", addr)
	log.Printf("📍 Entorno: %s", cfg.App.Environment)
	log.Printf("🗄️  Base de datos: %s@%s:%s/%s", cfg.DB.User, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)
	log.Printf("🔐 JWT Secret configurado: %s", maskSecret(cfg.JWT.Secret))
	log.Printf("🍪 Autenticación basada en cookies HttpOnly habilitada")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("❌ Error iniciando servidor: %v", err)
	}
}

// maskSecret enmascara el secret para no mostrarlo completo en logs
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}
