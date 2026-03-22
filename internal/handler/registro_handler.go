package handler

import (
	"net/http"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/service"
	"coviar_backend/pkg/httputil"
)

type RegistroHandler struct {
	service     *service.RegistroService
	emailSender func(cuentaID int, email string)
}

func NewRegistroHandler(service *service.RegistroService) *RegistroHandler {
	return &RegistroHandler{service: service}
}

// SetEmailSender configura el hook que se invoca tras un registro exitoso
// para enviar el email de verificación. Se llama de forma asíncrona.
func (h *RegistroHandler) SetEmailSender(fn func(cuentaID int, email string)) {
	h.emailSender = fn
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

	// Disparar envío de email de verificación en segundo plano
	if h.emailSender != nil {
		go h.emailSender(resp.IDCuenta, req.Cuenta.EmailLogin)
	}

	httputil.RespondJSON(w, http.StatusCreated, resp)
}
