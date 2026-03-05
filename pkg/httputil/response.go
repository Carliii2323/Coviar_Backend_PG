package httputil

import (
	"encoding/json"
	"errors"
	"net/http"

	"coviar_backend/internal/domain"
	"coviar_backend/pkg/validator"
)

type ErrorResponse struct {
	Error   string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
}

type SuccessResponse struct {
	Data interface{} `json:"data,omitempty"`
}

// RespondJSON envía una respuesta JSON
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// RespondError envía un error como JSON
func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, ErrorResponse{Error: message})
}

// DecodeJSON decodifica JSON del request
func DecodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// HandleServiceError maneja errores del servicio de forma centralizada
func HandleServiceError(w http.ResponseWriter, err error) {
	// Log del error para debugging
	println("ERROR:", err.Error())

	// Validación
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "errores de validación",
			Details: validationErrs,
		})
		return
	}

	// Errores de dominio
	switch {
	case errors.Is(err, domain.ErrNotFound):
		RespondError(w, http.StatusNotFound, "recurso no encontrado")
	case errors.Is(err, domain.ErrEmailYaRegistrado):
		RespondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrCUITYaRegistrado):
		RespondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrNoAutorizado):
		RespondError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, domain.ErrAutoevaluacionesPendientes):
		RespondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrResponsableYaDadoDeBaja):
		RespondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrSinResponsable):
		RespondError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, domain.ErrValidation):
		RespondError(w, http.StatusBadRequest, "error de validación")
	case errors.Is(err, domain.ErrInvalidCredentials):
		RespondError(w, http.StatusUnauthorized, err.Error())
	default:
		RespondError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
