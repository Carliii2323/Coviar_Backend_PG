package handler

import (
	"net/http"
	"strconv"

	"coviar_backend/internal/service"
	"coviar_backend/pkg/httputil"
	"coviar_backend/pkg/router"
)

type UbicacionHandler struct {
	service *service.UbicacionService
}

func NewUbicacionHandler(service *service.UbicacionService) *UbicacionHandler {
	return &UbicacionHandler{service: service}
}

// ===== PROVINCIAS =====

func (h *UbicacionHandler) GetProvincias(w http.ResponseWriter, r *http.Request) {
	provincias, err := h.service.GetProvincias(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, provincias)
}

func (h *UbicacionHandler) GetProvinciaByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(router.GetParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	provincia, err := h.service.GetProvinciaByID(r.Context(), id)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, provincia)
}

// ===== DEPARTAMENTOS =====

func (h *UbicacionHandler) GetDepartamentos(w http.ResponseWriter, r *http.Request) {
	provinciaIDStr := r.URL.Query().Get("provincia")
	if provinciaIDStr != "" {
		provinciaID, err := strconv.Atoi(provinciaIDStr)
		if err != nil {
			httputil.RespondError(w, http.StatusBadRequest, "ID de provincia inválido")
			return
		}

		departamentos, err := h.service.GetDepartamentosByProvinciaID(r.Context(), provinciaID)
		if err != nil {
			httputil.HandleServiceError(w, err)
			return
		}
		httputil.RespondJSON(w, http.StatusOK, departamentos)
		return
	}

	departamentos, err := h.service.GetDepartamentos(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, departamentos)
}

func (h *UbicacionHandler) GetDepartamentoByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(router.GetParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	departamento, err := h.service.GetDepartamentoByID(r.Context(), id)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, departamento)
}

// ===== LOCALIDADES =====

func (h *UbicacionHandler) GetLocalidades(w http.ResponseWriter, r *http.Request) {
	departamentoIDStr := r.URL.Query().Get("departamento")
	if departamentoIDStr != "" {
		departamentoID, err := strconv.Atoi(departamentoIDStr)
		if err != nil {
			httputil.RespondError(w, http.StatusBadRequest, "ID de departamento inválido")
			return
		}

		localidades, err := h.service.GetLocalidadesByDepartamentoID(r.Context(), departamentoID)
		if err != nil {
			httputil.HandleServiceError(w, err)
			return
		}
		httputil.RespondJSON(w, http.StatusOK, localidades)
		return
	}

	localidades, err := h.service.GetLocalidades(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, localidades)
}

func (h *UbicacionHandler) GetLocalidadByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(router.GetParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	localidad, err := h.service.GetLocalidadByID(r.Context(), id)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, localidad)
}
