package handler

import (
	"net/http"
	"path/filepath"
	"strconv"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/middleware"
	"coviar_backend/internal/service"
	"coviar_backend/pkg/httputil"
	"coviar_backend/pkg/router"
)

type EvidenciaHandler struct {
	service *service.EvidenciaService
}

func NewEvidenciaHandler(service *service.EvidenciaService) *EvidenciaHandler {
	return &EvidenciaHandler{
		service: service,
	}
}

// verificarOwnership verifica que la autoevaluación pertenezca al usuario autenticado.
// Los administradores tienen acceso irrestricto. Retorna false y escribe 403 si no tiene acceso.
func (h *EvidenciaHandler) verificarOwnership(w http.ResponseWriter, r *http.Request, idAutoevaluacion int) bool {
	tipo, _ := r.Context().Value(middleware.UserTipoKey).(string)
	if tipo == string(domain.TipoCuentaAdministradorApp) {
		return true
	}
	bodegaID, _ := r.Context().Value(middleware.BodegaIDKey).(int)
	if err := h.service.VerificarPropietario(r.Context(), idAutoevaluacion, bodegaID); err != nil {
		httputil.RespondError(w, http.StatusForbidden, "no tenés permiso para acceder a esta autoevaluación")
		return false
	}
	return true
}

// AgregarEvidencia POST /api/autoevaluaciones/{id_autoevaluacion}/respuestas/{id_respuesta}/evidencias
func (h *EvidenciaHandler) AgregarEvidencia(w http.ResponseWriter, r *http.Request) {
	idAutoevaluacionStr := router.GetParam(r, "id_autoevaluacion")
	idRespuestaStr := router.GetParam(r, "id_respuesta")

	idAutoevaluacion, err := strconv.Atoi(idAutoevaluacionStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID autoevaluación inválido")
		return
	}

	idRespuesta, err := strconv.Atoi(idRespuestaStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID respuesta inválido")
		return
	}

	if !h.verificarOwnership(w, r, idAutoevaluacion) {
		return
	}

	if err := r.ParseMultipartForm(2621440); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Error parseando formulario: "+err.Error())
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	evidencia, err := h.service.AgregarEvidencia(
		r.Context(),
		idAutoevaluacion,
		idRespuesta,
		handler.Filename,
		file,
	)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"mensaje":   "Evidencia agregada correctamente",
		"evidencia": evidencia,
	})
}

// ObtenerEvidencia GET /api/autoevaluaciones/{id_autoevaluacion}/respuestas/{id_respuesta}/evidencia
func (h *EvidenciaHandler) ObtenerEvidencia(w http.ResponseWriter, r *http.Request) {
	idAutoevaluacionStr := router.GetParam(r, "id_autoevaluacion")
	idRespuestaStr := router.GetParam(r, "id_respuesta")

	idAutoevaluacion, err := strconv.Atoi(idAutoevaluacionStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID autoevaluación inválido")
		return
	}

	idRespuesta, err := strconv.Atoi(idRespuestaStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID respuesta inválido")
		return
	}

	if !h.verificarOwnership(w, r, idAutoevaluacion) {
		return
	}

	evidencia, err := h.service.ObtenerEvidencia(r.Context(), idRespuesta)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	if evidencia == nil {
		httputil.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"mensaje":   "No hay evidencia para esta respuesta",
			"evidencia": nil,
		})
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"evidencia": evidencia,
	})
}

// ObtenerEvidenciasPorAutoevaluacion GET /api/autoevaluaciones/{id_autoevaluacion}/evidencias
func (h *EvidenciaHandler) ObtenerEvidenciasPorAutoevaluacion(w http.ResponseWriter, r *http.Request) {
	idAutoevaluacionStr := router.GetParam(r, "id_autoevaluacion")

	idAutoevaluacion, err := strconv.Atoi(idAutoevaluacionStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID autoevaluación inválido")
		return
	}

	if !h.verificarOwnership(w, r, idAutoevaluacion) {
		return
	}

	evidencias, err := h.service.ObtenerEvidenciasPorAutoevaluacion(r.Context(), idAutoevaluacion)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	if evidencias == nil {
		evidencias = make([]*domain.Evidencia, 0)
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"evidencias": evidencias,
		"total":      len(evidencias),
	})
}

// DescargarEvidencia GET /api/autoevaluaciones/{id_autoevaluacion}/respuestas/{id_respuesta}/evidencia/descargar
func (h *EvidenciaHandler) DescargarEvidencia(w http.ResponseWriter, r *http.Request) {
	idAutoevaluacionStr := router.GetParam(r, "id_autoevaluacion")
	idRespuestaStr := router.GetParam(r, "id_respuesta")

	idAutoevaluacion, err := strconv.Atoi(idAutoevaluacionStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID autoevaluación inválido")
		return
	}

	idRespuesta, err := strconv.Atoi(idRespuestaStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID respuesta inválido")
		return
	}

	if !h.verificarOwnership(w, r, idAutoevaluacion) {
		return
	}

	evidencia, err := h.service.ObtenerEvidencia(r.Context(), idRespuesta)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	if evidencia == nil {
		httputil.RespondError(w, http.StatusNotFound, "Evidencia no encontrada")
		return
	}

	fileData, fileName, err := h.service.DescargarEvidencia(r.Context(), evidencia)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(fileName)+"\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(fileData)))

	_, err = w.Write(fileData)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "Error descargando archivo")
		return
	}
}

// DescargarTodasEvidencias GET /api/autoevaluaciones/{id_autoevaluacion}/evidencias/descargar
func (h *EvidenciaHandler) DescargarTodasEvidencias(w http.ResponseWriter, r *http.Request) {
	idAutoevaluacionStr := router.GetParam(r, "id_autoevaluacion")

	idAutoevaluacion, err := strconv.Atoi(idAutoevaluacionStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID autoevaluación inválido")
		return
	}

	if !h.verificarOwnership(w, r, idAutoevaluacion) {
		return
	}

	zipData, err := h.service.DescargarTodasEvidenciasZip(r.Context(), idAutoevaluacion)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"evidencias_ae"+idAutoevaluacionStr+".zip\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(zipData)))

	_, err = w.Write(zipData)
	if err != nil {
		httputil.RespondError(w, http.StatusInternalServerError, "Error descargando archivo")
		return
	}
}

// EliminarEvidencia DELETE /api/autoevaluaciones/{id_autoevaluacion}/respuestas/{id_respuesta}/evidencia
func (h *EvidenciaHandler) EliminarEvidencia(w http.ResponseWriter, r *http.Request) {
	idAutoevaluacionStr := router.GetParam(r, "id_autoevaluacion")
	idRespuestaStr := router.GetParam(r, "id_respuesta")

	idAutoevaluacion, err := strconv.Atoi(idAutoevaluacionStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID autoevaluación inválido")
		return
	}

	idRespuesta, err := strconv.Atoi(idRespuestaStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID respuesta inválido")
		return
	}

	if !h.verificarOwnership(w, r, idAutoevaluacion) {
		return
	}

	if err := h.service.EliminarEvidencia(r.Context(), idAutoevaluacion, idRespuesta); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{
		"mensaje": "Evidencia eliminada correctamente",
	})
}

// CambiarEvidencia PUT /api/autoevaluaciones/{id_autoevaluacion}/respuestas/{id_respuesta}/evidencia
func (h *EvidenciaHandler) CambiarEvidencia(w http.ResponseWriter, r *http.Request) {
	idAutoevaluacionStr := router.GetParam(r, "id_autoevaluacion")
	idRespuestaStr := router.GetParam(r, "id_respuesta")

	idAutoevaluacion, err := strconv.Atoi(idAutoevaluacionStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID autoevaluación inválido")
		return
	}

	idRespuesta, err := strconv.Atoi(idRespuestaStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "ID respuesta inválido")
		return
	}

	if !h.verificarOwnership(w, r, idAutoevaluacion) {
		return
	}

	if err := r.ParseMultipartForm(2621440); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Error parseando formulario: "+err.Error())
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	evidencia, err := h.service.CambiarEvidencia(
		r.Context(),
		idAutoevaluacion,
		idRespuesta,
		handler.Filename,
		file,
	)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"mensaje":   "Evidencia reemplazada correctamente",
		"evidencia": evidencia,
	})
}
