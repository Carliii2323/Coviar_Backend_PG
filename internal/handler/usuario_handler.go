package handler

import (
	"net/http"
	"strconv"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/service"
	"coviar_backend/pkg/httputil"
	"coviar_backend/pkg/router"
)

type UsuarioHandler struct {
	service *service.UsuarioService
}

func NewUsuarioHandler(service *service.UsuarioService) *UsuarioHandler {
	return &UsuarioHandler{service: service}
}

// Create maneja POST /api/usuarios
func (h *UsuarioHandler) Create(w http.ResponseWriter, r *http.Request) {
	var dto domain.UsuarioDTO
	if err := httputil.DecodeJSON(r, &dto); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	usuario, err := h.service.Create(r.Context(), &dto)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, usuario.ToPublic())
}

// Login maneja POST /api/usuarios/login
func (h *UsuarioHandler) Login(w http.ResponseWriter, r *http.Request) {
	var login domain.UsuarioLogin
	if err := httputil.DecodeJSON(r, &login); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	usuario, err := h.service.Verify(r.Context(), &login)
	if err != nil {
		httputil.RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	httputil.RespondJSON(w, http.StatusOK, usuario.ToPublic())
}

// GetAll maneja GET /api/usuarios
func (h *UsuarioHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	usuarios, err := h.service.GetAll(r.Context())
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "Error obteniendo usuarios")
		return
	}

	// Limpiar password hashes
	for _, u := range usuarios {
		u.PasswordHash = ""
	}

	httputil.RespondJSON(w, http.StatusOK, usuarios)
}

// GetByID maneja GET /api/usuarios/{id}
func (h *UsuarioHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	usuario, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		httputil.RespondError(w, http.StatusNotFound, "Usuario no encontrado")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, usuario.ToPublic())
}

// Delete maneja DELETE /api/usuarios/{id}
func (h *UsuarioHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	err = h.service.Delete(r.Context(), id)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "Error eliminando usuario")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
