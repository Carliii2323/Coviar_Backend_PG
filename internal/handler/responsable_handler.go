package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/middleware"
	"coviar_backend/internal/service"
	"coviar_backend/pkg/httputil"
	"coviar_backend/pkg/router"
)

type ResponsableHandler struct {
	service *service.ResponsableService
}

func NewResponsableHandler(service *service.ResponsableService) *ResponsableHandler {
	return &ResponsableHandler{service: service}
}

func (h *ResponsableHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// Verificar permisos
	userID := r.Context().Value(middleware.UserIDKey).(int)
	userTipo := r.Context().Value(middleware.UserTipoKey).(string)

	canAccess, err := h.service.CanAccess(r.Context(), id, userID, userTipo)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	if !canAccess {
		httputil.RespondError(w, http.StatusForbidden, "no tiene permisos para acceder a este recurso")
		return
	}

	responsable, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, responsable)
}

func (h *ResponsableHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// Verificar permisos
	userID := r.Context().Value(middleware.UserIDKey).(int)
	userTipo := r.Context().Value(middleware.UserTipoKey).(string)

	canAccess, err := h.service.CanAccess(r.Context(), id, userID, userTipo)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	if !canAccess {
		httputil.RespondError(w, http.StatusForbidden, "no tiene permisos para modificar este recurso")
		return
	}

	var req domain.ResponsableUpdateDTO
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if err := h.service.Update(r.Context(), id, &req); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"mensaje": "Responsable actualizado"})
}

func (h *ResponsableHandler) DarDeBaja(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// Verificar permisos
	userID := r.Context().Value(middleware.UserIDKey).(int)
	userTipo := r.Context().Value(middleware.UserTipoKey).(string)

	canAccess, err := h.service.CanAccess(r.Context(), id, userID, userTipo)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	if !canAccess {
		httputil.RespondError(w, http.StatusForbidden, "no tiene permisos para dar de baja este recurso")
		return
	}

	var body struct {
		CancelarPendientes bool `json:"cancelar_pendientes"`
	}
	json.NewDecoder(r.Body).Decode(&body) //nolint:errcheck

	if err := h.service.DarDeBaja(r.Context(), id, body.CancelarPendientes); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"mensaje": "Responsable dado de baja exitosamente"})
}

func (h *ResponsableHandler) GetByCuentaID(w http.ResponseWriter, r *http.Request) {
	cuentaIDStr := router.GetParam(r, "cuenta_id")
	cuentaID, err := strconv.Atoi(cuentaIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID de cuenta inválido")
		return
	}

	// Verificar que el usuario autenticado solo acceda a sus propios responsables
	userID := r.Context().Value(middleware.UserIDKey).(int)
	userTipo := r.Context().Value(middleware.UserTipoKey).(string)

	// Solo puede acceder si es admin o es la misma cuenta
	if userTipo != string(domain.TipoCuentaAdministradorApp) && userID != cuentaID {
		httputil.RespondError(w, http.StatusForbidden, "no tiene permisos para acceder a este recurso")
		return
	}

	responsables, err := h.service.GetByCuentaID(r.Context(), cuentaID)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, responsables)
}

func (h *ResponsableHandler) Create(w http.ResponseWriter, r *http.Request) {
	cuentaIDStr := router.GetParam(r, "cuenta_id")
	cuentaID, err := strconv.Atoi(cuentaIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID de cuenta inválido")
		return
	}

	// Verificar que el usuario autenticado solo pueda crear responsables para su cuenta
	userID := r.Context().Value(middleware.UserIDKey).(int)
	userTipo := r.Context().Value(middleware.UserTipoKey).(string)

	// Solo puede crear si es admin o es la misma cuenta
	if userTipo != string(domain.TipoCuentaAdministradorApp) && userID != cuentaID {
		httputil.RespondError(w, http.StatusForbidden, "no tiene permisos para crear responsables en esta cuenta")
		return
	}

	var req domain.ResponsableUpdateDTO
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	responsable, err := h.service.Create(r.Context(), cuentaID, &req)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, responsable)
}
