package service

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository"
)

type EvidenciaService struct {
	evidenciaRepo      repository.EvidenciaRepository
	respuestaRepo      repository.RespuestaRepository
	autoevaluacionRepo repository.AutoevaluacionRepository
	bodegaRepo         repository.BodegaRepository
	indicadorRepo      repository.IndicadorRepository
}

func NewEvidenciaService(
	evidenciaRepo repository.EvidenciaRepository,
	respuestaRepo repository.RespuestaRepository,
	autoevaluacionRepo repository.AutoevaluacionRepository,
	bodegaRepo repository.BodegaRepository,
	indicadorRepo repository.IndicadorRepository,
) *EvidenciaService {
	return &EvidenciaService{
		evidenciaRepo:      evidenciaRepo,
		respuestaRepo:      respuestaRepo,
		autoevaluacionRepo: autoevaluacionRepo,
		bodegaRepo:         bodegaRepo,
		indicadorRepo:      indicadorRepo,
	}
}

// AgregarEvidencia agrega una evidencia a una respuesta
func (s *EvidenciaService) AgregarEvidencia(
	ctx context.Context,
	idAutoevaluacion int,
	idRespuesta int,
	nombreArchivo string,
	file io.Reader,
) (*domain.Evidencia, error) {
	// 1. Validar que la respuesta existe y pertenece a la autoevaluación
	_, err := s.getAndValidateRespuesta(ctx, idRespuesta, idAutoevaluacion)
	if err != nil {
		return nil, err
	}

	// 2. Verificar que no exista ya una evidencia para esta respuesta (relación 1:1)
	existente, err := s.evidenciaRepo.FindByRespuesta(ctx, idRespuesta)
	if err != nil {
		return nil, fmt.Errorf("error checking existing evidencia: %w", err)
	}
	if existente != nil {
		return nil, fmt.Errorf("ya existe una evidencia para esta respuesta, use el endpoint de cambio (PUT) para reemplazarla")
	}

	// 3. Obtener bodegaId desde la autoevaluación
	auto, err := s.autoevaluacionRepo.FindByID(ctx, idAutoevaluacion)
	if err != nil {
		return nil, fmt.Errorf("error getting autoevaluacion: %w", err)
	}
	bodegaId := auto.IDBodega

	// 2. Validar que el archivo sea PDF y no exceda 2MB
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if err := s.validatePDFFile(fileBytes, nombreArchivo); err != nil {
		return nil, err
	}

	// 3. Crear directorio si no existe
	basePath := "evidencias"
	bodegaPath := filepath.Join(basePath, fmt.Sprintf("%d", bodegaId))
	if err := os.MkdirAll(bodegaPath, 0755); err != nil {
		return nil, fmt.Errorf("error creating directory: %w", err)
	}

	// 4. Generar nombre del archivo único con timestamp
	timestamp := time.Now().Format("20060102_150405")
	cleanFileName := nombreArchivo
	if len(cleanFileName) > 4 {
		cleanFileName = cleanFileName[:len(cleanFileName)-4]
	}
	fileName := fmt.Sprintf("%s_ae%d_r%d_%s.pdf", cleanFileName, idAutoevaluacion, idRespuesta, timestamp)
	filePath := filepath.Join(bodegaPath, fileName)

	// 5. Guardar archivo
	if err := os.WriteFile(filePath, fileBytes, 0644); err != nil {
		return nil, fmt.Errorf("error saving file: %w", err)
	}

	// 6. Guardar en base de datos
	// Nombre: nombre generado con patrón {original}_ae{id_auto}_r{id_resp}_{timestamp}.pdf
	// Ubicacion: ruta completa donde se guarda el archivo
	evidencia := &domain.Evidencia{
		IDRespuesta: idRespuesta,
		Nombre:      fileName, // Nombre generado internamente (patrón único)
		Ubicacion:   filePath, // Ruta completa en el servidor
	}

	_, err = s.evidenciaRepo.Create(ctx, nil, evidencia)
	if err != nil {
		os.Remove(filePath)
		return nil, err
	}

	// 7. Actualizar estado_evidencia de la autoevaluación
	if err := s.updateAutoevaluacionEvidenciaStatus(ctx, idAutoevaluacion); err != nil {
		return nil, err
	}

	return evidencia, nil
}

// EliminarEvidencia elimina la evidencia de una respuesta (archivo + registro DB)
func (s *EvidenciaService) EliminarEvidencia(ctx context.Context, idAutoevaluacion int, idRespuesta int) error {
	// 1. Validar que la respuesta pertenece a la autoevaluación
	_, err := s.getAndValidateRespuesta(ctx, idRespuesta, idAutoevaluacion)
	if err != nil {
		return err
	}

	// 2. Buscar la evidencia existente
	evidencia, err := s.evidenciaRepo.FindByRespuesta(ctx, idRespuesta)
	if err != nil {
		return fmt.Errorf("error getting evidencia: %w", err)
	}
	if evidencia == nil {
		return fmt.Errorf("no existe evidencia para esta respuesta")
	}

	// 3. Eliminar archivo del disco
	os.Remove(evidencia.Ubicacion)

	// 4. Eliminar registro de la base de datos
	if err := s.evidenciaRepo.Delete(ctx, nil, evidencia.ID); err != nil {
		return fmt.Errorf("error deleting evidencia: %w", err)
	}

	// 5. Actualizar estado_evidencia de la autoevaluación
	if err := s.updateAutoevaluacionEvidenciaStatus(ctx, idAutoevaluacion); err != nil {
		return err
	}

	return nil
}

// CambiarEvidencia reemplaza la evidencia de una respuesta (elimina la anterior y sube la nueva)
func (s *EvidenciaService) CambiarEvidencia(ctx context.Context, idAutoevaluacion int, idRespuesta int, nombreArchivo string, file io.Reader) (*domain.Evidencia, error) {
	// 1. Eliminar la evidencia anterior si existe
	evidenciaAnterior, err := s.evidenciaRepo.FindByRespuesta(ctx, idRespuesta)
	if err != nil {
		return nil, fmt.Errorf("error getting evidencia anterior: %w", err)
	}
	if evidenciaAnterior != nil {
		os.Remove(evidenciaAnterior.Ubicacion)
		if err := s.evidenciaRepo.Delete(ctx, nil, evidenciaAnterior.ID); err != nil {
			return nil, fmt.Errorf("error deleting evidencia anterior: %w", err)
		}
	}

	// 2. Subir la nueva evidencia (reutiliza AgregarEvidencia)
	return s.AgregarEvidencia(ctx, idAutoevaluacion, idRespuesta, nombreArchivo, file)
}

// VerificarPropietario verifica que la autoevaluación pertenezca a la bodega indicada.
func (s *EvidenciaService) VerificarPropietario(ctx context.Context, idAutoevaluacion, idBodega int) error {
	auto, err := s.autoevaluacionRepo.FindByID(ctx, idAutoevaluacion)
	if err != nil {
		return domain.ErrNotFound
	}
	if auto.IDBodega != idBodega {
		return domain.ErrNoAutorizado
	}
	return nil
}

// ObtenerEvidencia obtiene la evidencia de una respuesta
func (s *EvidenciaService) ObtenerEvidencia(ctx context.Context, idRespuesta int) (*domain.Evidencia, error) {
	evidencia, err := s.evidenciaRepo.FindByRespuesta(ctx, idRespuesta)
	if err != nil {
		return nil, err
	}

	return evidencia, nil
}

// ObtenerEvidenciasPorAutoevaluacion obtiene todas las evidencias de una autoevaluación
func (s *EvidenciaService) ObtenerEvidenciasPorAutoevaluacion(ctx context.Context, idAutoevaluacion int) ([]*domain.Evidencia, error) {
	evidencias, err := s.evidenciaRepo.FindByAutoevaluacion(ctx, idAutoevaluacion)
	if err != nil {
		return nil, err
	}

	return evidencias, nil
}

// DescargarEvidencia devuelve el contenido del archivo de una evidencia
func (s *EvidenciaService) DescargarEvidencia(ctx context.Context, evidencia *domain.Evidencia) ([]byte, string, error) {
	fileData, err := os.ReadFile(evidencia.Ubicacion)
	if err != nil {
		return nil, "", fmt.Errorf("error reading evidence file: %w", err)
	}

	return fileData, evidencia.Nombre, nil
}

// DescargarTodasEvidenciasZip retorna un ZIP con todas las evidencias de una autoevaluación
func (s *EvidenciaService) DescargarTodasEvidenciasZip(ctx context.Context, idAutoevaluacion int) ([]byte, error) {
	evidencias, err := s.evidenciaRepo.FindByAutoevaluacion(ctx, idAutoevaluacion)
	if err != nil {
		return nil, fmt.Errorf("error getting evidencias: %w", err)
	}

	if len(evidencias) == 0 {
		return nil, fmt.Errorf("no evidencias found for this autoevaluacion")
	}

	zipBuffer := new(writeCounter)
	zipWriter := zip.NewWriter(zipBuffer)

	for _, ev := range evidencias {
		fileData, err := os.ReadFile(ev.Ubicacion)
		if err != nil {
			zipWriter.Close()
			return nil, fmt.Errorf("error reading evidence file %s: %w", ev.Nombre, err)
		}

		zipEntry, err := zipWriter.Create(ev.Nombre)
		if err != nil {
			zipWriter.Close()
			return nil, fmt.Errorf("error creating zip entry for %s: %w", ev.Nombre, err)
		}

		if _, err := zipEntry.Write(fileData); err != nil {
			zipWriter.Close()
			return nil, fmt.Errorf("error writing to zip entry for %s: %w", ev.Nombre, err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("error closing zip: %w", err)
	}

	return zipBuffer.Bytes(), nil
}

// validatePDFFile valida que el archivo sea un PDF real (magic bytes + extensión) y no exceda 2MB
func (s *EvidenciaService) validatePDFFile(fileBytes []byte, fileName string) error {
	// Validar extensión (case-insensitive)
	if !strings.HasSuffix(strings.ToLower(fileName), ".pdf") {
		return domain.ErrArchivoInvalido
	}

	// Validar magic bytes: todo PDF real comienza con "%PDF-"
	if len(fileBytes) < 5 || string(fileBytes[:5]) != "%PDF-" {
		return domain.ErrArchivoInvalido
	}

	// Validar tamaño máximo 2MB
	const maxSize = 2 * 1024 * 1024
	if len(fileBytes) > maxSize {
		return domain.ErrArchivoDemasiadoGrande
	}

	return nil
}

// getAndValidateRespuesta obtiene y valida una respuesta
func (s *EvidenciaService) getAndValidateRespuesta(ctx context.Context, idRespuesta int, idAutoevaluacion int) (*domain.Respuesta, error) {
	respuestas, err := s.respuestaRepo.FindByAutoevaluacion(ctx, idAutoevaluacion)
	if err != nil {
		return nil, fmt.Errorf("error getting respuestas: %w", err)
	}

	for _, r := range respuestas {
		if r.ID == idRespuesta {
			return r, nil
		}
	}

	return nil, fmt.Errorf("respuesta not found in autoevaluacion")
}

// updateAutoevaluacionEvidenciaStatus actualiza el estado de evidencia de la autoevaluación
func (s *EvidenciaService) updateAutoevaluacionEvidenciaStatus(ctx context.Context, idAutoevaluacion int) error {
	auto, err := s.autoevaluacionRepo.FindByID(ctx, idAutoevaluacion)
	if err != nil {
		return fmt.Errorf("error getting autoevaluacion: %w", err)
	}

	if auto.IDSegmento == nil {
		return fmt.Errorf("segmento not selected in autoevaluacion")
	}

	indicadores, err := s.indicadorRepo.FindBySegmento(ctx, *auto.IDSegmento)
	if err != nil {
		return fmt.Errorf("error getting indicadores: %w", err)
	}

	respuestas, err := s.respuestaRepo.FindByAutoevaluacion(ctx, idAutoevaluacion)
	if err != nil {
		return fmt.Errorf("error getting respuestas: %w", err)
	}

	// Contar evidencias de respuestas que pertenecen a indicadores del segmento
	evidenciasSegmento := 0
	for _, r := range respuestas {
		for _, idIndicador := range indicadores {
			if r.IDIndicador == idIndicador {
				evidencia, err := s.evidenciaRepo.FindByRespuesta(ctx, r.ID)
				if err == nil && evidencia != nil {
					evidenciasSegmento++
				}
				break
			}
		}
	}

	// Comparar contra el total de indicadores habilitados del segmento
	totalIndicadores := len(indicadores)

	var estado domain.EstadoEvidencia
	if evidenciasSegmento == 0 {
		estado = domain.EstadoSinEvidencia
	} else if evidenciasSegmento == totalIndicadores && totalIndicadores > 0 {
		estado = domain.EstadoCompleta
	} else {
		estado = domain.EstadoParcial
	}

	return s.autoevaluacionRepo.UpdateEvidenciaStatus(ctx, idAutoevaluacion, estado)
}

type writeCounter struct {
	data []byte
}

func (w *writeCounter) Write(p []byte) (n int, err error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

func (w *writeCounter) Bytes() []byte {
	return w.data
}
