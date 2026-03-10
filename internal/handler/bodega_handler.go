package handler

import (
	"net/http"
	"strconv"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/middleware"
	"coviar_backend/internal/service"
	"coviar_backend/pkg/httputil"
	"coviar_backend/pkg/router"
)

// esPropietario verifica que el usuario autenticado sea dueño de la bodega,
// o que sea administrador. Retorna false y escribe 403 si no tiene acceso.
func esPropietario(w http.ResponseWriter, r *http.Request, idBodega int) bool {
	tipo, _ := r.Context().Value(middleware.UserTipoKey).(string)
	if tipo == string(domain.TipoCuentaAdministradorApp) {
		return true
	}
	bodegaID, _ := r.Context().Value(middleware.BodegaIDKey).(int)
	if bodegaID != idBodega {
		httputil.RespondError(w, http.StatusForbidden, "no tenés permiso para acceder a esta bodega")
		return false
	}
	return true
}

type BodegaHandler struct {
	service              *service.BodegaService
	autoevaluacionService *service.AutoevaluacionService
}

func NewBodegaHandler(service *service.BodegaService, autoevaluacionService *service.AutoevaluacionService) *BodegaHandler {
	return &BodegaHandler{service: service, autoevaluacionService: autoevaluacionService}
}

func (h *BodegaHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	bodegas, err := h.service.GetAllAdmin(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	if bodegas == nil {
		bodegas = []*domain.BodegaAdminItem{}
	}

	httputil.RespondJSON(w, http.StatusOK, bodegas)
}

func (h *BodegaHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if !esPropietario(w, r, id) {
		return
	}

	bodega, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, bodega)
}

func (h *BodegaHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if !esPropietario(w, r, id) {
		return
	}

	var req domain.BodegaUpdateDTO
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if err := h.service.Update(r.Context(), id, &req); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"mensaje": "Bodega actualizada"})
}

// GetResultadosAutoevaluacion GET /api/bodegas/{id}/resultados-autoevaluacion
func (h *BodegaHandler) GetResultadosAutoevaluacion(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if !esPropietario(w, r, id) {
		return
	}

	resultado, err := h.autoevaluacionService.GetResultadosBodega(r.Context(), id)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, resultado)
}
