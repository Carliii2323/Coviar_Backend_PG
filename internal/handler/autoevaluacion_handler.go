package handler

import (
	"net/http"
	"strconv"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/service"
	"coviar_backend/pkg/httputil"
	"coviar_backend/pkg/router"
)

type AutoevaluacionHandler struct {
	service *service.AutoevaluacionService
}

func NewAutoevaluacionHandler(service *service.AutoevaluacionService) *AutoevaluacionHandler {
	return &AutoevaluacionHandler{service: service}
}

// CreateAutoevaluacion POST /api/autoevaluaciones
func (h *AutoevaluacionHandler) CreateAutoevaluacion(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateAutoevaluacionRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	response, err := h.service.CreateAutoevaluacion(r.Context(), req.IDBodega)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	// Si hay una autoevaluación pendiente, retornar con código 200
	// Si se creó una nueva, retornar con código 201
	if response.AutoevaluacionPendiente != nil && response.AutoevaluacionPendiente.Estado == domain.EstadoPendiente && len(response.Respuestas) > 0 {
		httputil.RespondJSON(w, http.StatusOK, response)
	} else {
		httputil.RespondJSON(w, http.StatusCreated, response)
	}
}

// GetSegmentos GET /api/autoevaluaciones/{id_autoevaluacion}/segmentos
func (h *AutoevaluacionHandler) GetSegmentos(w http.ResponseWriter, r *http.Request) {
	segmentos, err := h.service.GetSegmentos(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, segmentos)
}

// SeleccionarSegmento PUT /api/autoevaluaciones/{id_autoevaluacion}/segmento
func (h *AutoevaluacionHandler) SeleccionarSegmento(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id_autoevaluacion")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req domain.SeleccionarSegmentoRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if err := h.service.SeleccionarSegmento(r.Context(), id, req.IDSegmento); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"mensaje": "Segmento seleccionado correctamente"})
}

// GetEstructura GET /api/autoevaluaciones/{id_autoevaluacion}/estructura
func (h *AutoevaluacionHandler) GetEstructura(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id_autoevaluacion")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	estructura, err := h.service.GetEstructura(r.Context(), id)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, estructura)
}

// GuardarRespuestas POST /api/autoevaluaciones/{id_autoevaluacion}/respuestas
func (h *AutoevaluacionHandler) GuardarRespuestas(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id_autoevaluacion")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req domain.GuardarRespuestasRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	respuestasGuardadas, err := h.service.GuardarRespuestas(r.Context(), id, req.Respuestas)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"mensaje":    "Respuestas guardadas correctamente",
		"respuestas": respuestasGuardadas,
	})
}

// CompletarAutoevaluacion POST /api/autoevaluaciones/{id_autoevaluacion}/completar
func (h *AutoevaluacionHandler) CompletarAutoevaluacion(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id_autoevaluacion")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.service.CompletarAutoevaluacion(r.Context(), id); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"mensaje": "Autoevaluación completada correctamente"})
}

// CancelarAutoevaluacion POST /api/autoevaluaciones/{id_autoevaluacion}/cancelar
func (h *AutoevaluacionHandler) CancelarAutoevaluacion(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id_autoevaluacion")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.service.CancelarAutoevaluacion(r.Context(), id); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"mensaje": "Autoevaluación cancelada correctamente"})
}

// GetHistorialAutoevaluaciones GET /api/autoevaluaciones/historial?id_bodega=X
func (h *AutoevaluacionHandler) GetHistorialAutoevaluaciones(w http.ResponseWriter, r *http.Request) {
	idBodegaStr := r.URL.Query().Get("id_bodega")
	if idBodegaStr == "" {
		httputil.RespondError(w, http.StatusBadRequest, "id_bodega es requerido")
		return
	}

	idBodega, err := strconv.Atoi(idBodegaStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "id_bodega inválido")
		return
	}

	historial, err := h.service.GetHistorialAutoevaluaciones(r.Context(), idBodega)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, historial)
}

// GetResultadosAutoevaluacion GET /api/autoevaluaciones/{id_autoevaluacion}/resultados
func (h *AutoevaluacionHandler) GetResultadosAutoevaluacion(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id_autoevaluacion")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	resultado, err := h.service.GetResultadosDetallados(r.Context(), id)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, resultado)
}
