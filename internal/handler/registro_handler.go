package handler

import (
	"net/http"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/service"
	"coviar_backend/pkg/httputil"
)

type RegistroHandler struct {
	service *service.RegistroService
}

func NewRegistroHandler(service *service.RegistroService) *RegistroHandler {
	return &RegistroHandler{service: service}
}

func (h *RegistroHandler) RegistrarBodega(w http.ResponseWriter, r *http.Request) {
	var req domain.RegistroRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	resp, err := h.service.RegistrarBodega(r.Context(), &req)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, resp)
}
