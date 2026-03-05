package service

import (
	"context"
	"fmt"
	"os"
	"strings"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository"
)

type AutoevaluacionService struct {
	autoevaluacionRepo repository.AutoevaluacionRepository
	segmentoRepo       repository.SegmentoRepository
	capituloRepo       repository.CapituloRepository
	indicadorRepo      repository.IndicadorRepository
	nivelRespuestaRepo repository.NivelRespuestaRepository
	respuestaRepo      repository.RespuestaRepository
	evidenciaRepo      repository.EvidenciaRepository
	responsableRepo    repository.ResponsableRepository
}

func NewAutoevaluacionService(
	autoevaluacionRepo repository.AutoevaluacionRepository,
	segmentoRepo repository.SegmentoRepository,
	capituloRepo repository.CapituloRepository,
	indicadorRepo repository.IndicadorRepository,
	nivelRespuestaRepo repository.NivelRespuestaRepository,
	respuestaRepo repository.RespuestaRepository,
	evidenciaRepo repository.EvidenciaRepository,
	responsableRepo repository.ResponsableRepository,
) *AutoevaluacionService {
	return &AutoevaluacionService{
		autoevaluacionRepo: autoevaluacionRepo,
		segmentoRepo:       segmentoRepo,
		capituloRepo:       capituloRepo,
		indicadorRepo:      indicadorRepo,
		nivelRespuestaRepo: nivelRespuestaRepo,
		respuestaRepo:      respuestaRepo,
		evidenciaRepo:      evidenciaRepo,
		responsableRepo:    responsableRepo,
	}
}

// CreateAutoevaluacion crea una nueva autoevaluación para una bodega
/*func (s *AutoevaluacionService) CreateAutoevaluacion(ctx context.Context, idBodega int) (*domain.Autoevaluacion, error) {
	auto := &domain.Autoevaluacion{
		IDBodega: idBodega,
	}

	id, err := s.autoevaluacionRepo.Create(ctx, nil, auto)
	if err != nil {
		return nil, fmt.Errorf("error creating autoevaluacion: %w", err)
	}

	auto.ID = id
	return auto, nil
}*/

// CreateAutoevaluacion crea una nueva autoevaluación para una bodega o retorna la pendiente
func (s *AutoevaluacionService) CreateAutoevaluacion(ctx context.Context, idBodega int) (*domain.AutoevaluacionPendienteResponse, error) {
	// Verificar que haya un responsable activo asignado a la bodega
	responsable, err := s.responsableRepo.FindActivoByBodega(ctx, idBodega)
	if err != nil {
		return nil, fmt.Errorf("error checking responsable: %w", err)
	}
	if responsable == nil {
		return nil, domain.ErrSinResponsable
	}

	// Verificar si ya existe una autoevaluación pendiente para esta bodega
	autoPendiente, err := s.autoevaluacionRepo.FindPendienteByBodega(ctx, idBodega)
	if err != nil {
		return nil, fmt.Errorf("error checking for pending autoevaluacion: %w", err)
	}

	// Si existe una autoevaluación pendiente, retornarla con sus respuestas
	if autoPendiente != nil {
		// Obtener las respuestas existentes
		respuestas, err := s.respuestaRepo.FindByAutoevaluacion(ctx, autoPendiente.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting respuestas: %w", err)
		}

		// Convertir respuestas al formato DTO
		respuestasDTO := make([]domain.GuardarRespuestaRequest, len(respuestas))
		for i, resp := range respuestas {
			respuestasDTO[i] = domain.GuardarRespuestaRequest{
				IDRespuesta:      resp.ID,
				IDIndicador:      resp.IDIndicador,
				IDNivelRespuesta: resp.IDNivelRespuesta,
			}
		}

		return &domain.AutoevaluacionPendienteResponse{
			AutoevaluacionPendiente: autoPendiente,
			Respuestas:              respuestasDTO,
			Mensaje:                 "Ya existe una autoevaluación pendiente. Para crear una nueva, primero debe cancelar la actual.",
		}, nil
	}

	// No existe autoevaluación pendiente, crear una nueva
	auto := &domain.Autoevaluacion{
		IDBodega:      idBodega,
		IDResponsable: &responsable.ID,
	}

	id, err := s.autoevaluacionRepo.Create(ctx, nil, auto)
	if err != nil {
		return nil, fmt.Errorf("error creating autoevaluacion: %w", err)
	}

	auto.ID = id
	return &domain.AutoevaluacionPendienteResponse{
		AutoevaluacionPendiente: auto,
		Respuestas:              []domain.GuardarRespuestaRequest{},
		Mensaje:                 "Autoevaluación creada correctamente",
	}, nil
}

// GetSegmentos obtiene todos los segmentos disponibles
func (s *AutoevaluacionService) GetSegmentos(ctx context.Context) ([]*domain.Segmento, error) {
	segmentos, err := s.segmentoRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting segmentos: %w", err)
	}
	return segmentos, nil
}

// SeleccionarSegmento selecciona un segmento para la autoevaluación
func (s *AutoevaluacionService) SeleccionarSegmento(ctx context.Context, idAutoevaluacion int, idSegmento int) error {
	// Verificar que la autoevaluación existe
	auto, err := s.autoevaluacionRepo.FindByID(ctx, idAutoevaluacion)
	if err != nil {
		return fmt.Errorf("error finding autoevaluacion: %w", err)
	}

	if auto == nil {
		return domain.ErrNotFound
	}

	// Verificar que el segmento existe
	seg, err := s.segmentoRepo.FindByID(ctx, idSegmento)
	if err != nil {
		return fmt.Errorf("error finding segmento: %w", err)
	}

	if seg == nil {
		return domain.ErrNotFound
	}

	// Actualizar la autoevaluación con el segmento
	err = s.autoevaluacionRepo.UpdateSegmento(ctx, idAutoevaluacion, idSegmento)
	if err != nil {
		return fmt.Errorf("error selecting segmento: %w", err)
	}

	// Inicializar estado_evidencia a SIN_EVIDENCIA
	if err := s.autoevaluacionRepo.UpdateEvidenciaStatus(ctx, idAutoevaluacion, domain.EstadoSinEvidencia); err != nil {
		return fmt.Errorf("error initializing evidencia status: %w", err)
	}

	return nil
}

// GetEstructura obtiene la estructura del cuestionario con indicadores habilitados según el segmento
func (s *AutoevaluacionService) GetEstructura(ctx context.Context, idAutoevaluacion int) (*domain.EstructuraAutoevaluacion, error) {
	// Obtener autoevaluación
	auto, err := s.autoevaluacionRepo.FindByID(ctx, idAutoevaluacion)
	if err != nil {
		return nil, fmt.Errorf("error finding autoevaluacion: %w", err)
	}

	if auto == nil || auto.IDSegmento == nil {
		return nil, fmt.Errorf("autoevaluacion not found or segmento not selected")
	}

	// Obtener indicadores habilitados para este segmento
	habilitadosIds, err := s.indicadorRepo.FindBySegmento(ctx, *auto.IDSegmento)
	if err != nil {
		return nil, fmt.Errorf("error getting enabled indicators: %w", err)
	}

	// Convertir a mapa para búsqueda rápida
	habilitadosMap := make(map[int]bool)
	for _, id := range habilitadosIds {
		habilitadosMap[id] = true
	}

	// Obtener todos los capítulos
	capitulos, err := s.capituloRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting capitulos: %w", err)
	}

	estructura := &domain.EstructuraAutoevaluacion{
		Capitulos: make([]*domain.CapituloEstructura, 0),
	}

	for _, cap := range capitulos {
		// Obtener indicadores del capítulo
		indicadores, err := s.indicadorRepo.FindByCapitulo(ctx, cap.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting indicadores: %w", err)
		}

		capEstructura := &domain.CapituloEstructura{
			Capitulo:    cap,
			Indicadores: make([]*domain.IndicadorConHabilitacion, 0),
		}

		for _, ind := range indicadores {
			// Obtener niveles de respuesta
			niveles, err := s.nivelRespuestaRepo.FindByIndicador(ctx, ind.ID)
			if err != nil {
				return nil, fmt.Errorf("error getting niveles_respuesta: %w", err)
			}

			habilitado := habilitadosMap[ind.ID]

			indConHab := &domain.IndicadorConHabilitacion{
				Indicador:        ind,
				NivelesRespuesta: niveles,
				Habilitado:       habilitado,
			}

			capEstructura.Indicadores = append(capEstructura.Indicadores, indConHab)
		}

		estructura.Capitulos = append(estructura.Capitulos, capEstructura)
	}

	return estructura, nil
}

// GuardarRespuestas guarda las respuestas de la autoevaluación y retorna las respuestas con sus IDs
func (s *AutoevaluacionService) GuardarRespuestas(ctx context.Context, idAutoevaluacion int, respuestas []domain.GuardarRespuestaRequest) ([]*domain.Respuesta, error) {
	// Verificar que la autoevaluación existe
	auto, err := s.autoevaluacionRepo.FindByID(ctx, idAutoevaluacion)
	if err != nil {
		return nil, fmt.Errorf("error finding autoevaluacion: %w", err)
	}

	if auto == nil {
		return nil, domain.ErrNotFound
	}

	// Obtener respuestas existentes para detectar cambios de nivel
	existentes, err := s.respuestaRepo.FindByAutoevaluacion(ctx, idAutoevaluacion)
	if err != nil {
		return nil, fmt.Errorf("error getting existing respuestas: %w", err)
	}

	// Mapa de id_indicador -> respuesta existente
	existentesMap := make(map[int]*domain.Respuesta)
	for _, r := range existentes {
		existentesMap[r.IDIndicador] = r
	}

	// Guardar nuevas respuestas o actualizar existentes
	resultado := make([]*domain.Respuesta, 0, len(respuestas))
	for _, respReq := range respuestas {
		// Si la respuesta ya existía con un nivel diferente, eliminar su evidencia
		if existente, ok := existentesMap[respReq.IDIndicador]; ok {
			if existente.IDNivelRespuesta != respReq.IDNivelRespuesta {
				evidencia, err := s.evidenciaRepo.FindByRespuesta(ctx, existente.ID)
				if err == nil && evidencia != nil {
					os.Remove(evidencia.Ubicacion)
					s.evidenciaRepo.Delete(ctx, nil, evidencia.ID)
				}
			}
		}

		respuesta := &domain.Respuesta{
			IDNivelRespuesta: respReq.IDNivelRespuesta,
			IDIndicador:      respReq.IDIndicador,
			IDAutoevaluacion: idAutoevaluacion,
		}

		id, err := s.respuestaRepo.Upsert(ctx, nil, respuesta)
		if err != nil {
			return nil, fmt.Errorf("error guarding respuesta: %w", err)
		}
		respuesta.ID = id
		resultado = append(resultado, respuesta)
	}

	return resultado, nil
}

// CompletarAutoevaluacion marca la autoevaluación como completada
/*func (s *AutoevaluacionService) CompletarAutoevaluacion(ctx context.Context, idAutoevaluacion int) error {
	// Verificar que la autoevaluación existe
	auto, err := s.autoevaluacionRepo.FindByID(ctx, idAutoevaluacion)
	if err != nil {
		return fmt.Errorf("error finding autoevaluacion: %w", err)
	}

	if auto == nil {
		return domain.ErrNotFound
	}

	// Obtener respuestas para validar que todas las preguntas fueron respondidas
	respuestas, err := s.respuestaRepo.FindByAutoevaluacion(ctx, idAutoevaluacion)
	if err != nil {
		return fmt.Errorf("error getting respuestas: %w", err)
	}

	// Validación básica: debe haber al menos una respuesta
	if len(respuestas) == 0 {
		return fmt.Errorf("autoevaluacion must have at least one respuesta")
	}

	// Marcar como completada
	err = s.autoevaluacionRepo.Complete(ctx, idAutoevaluacion)
	if err != nil {
		return fmt.Errorf("error completing autoevaluacion: %w", err)
	}

	return nil
}*/

// CompletarAutoevaluacion marca la autoevaluación como completada
func (s *AutoevaluacionService) CompletarAutoevaluacion(ctx context.Context, idAutoevaluacion int) error {
	auto, err := s.autoevaluacionRepo.FindByID(ctx, idAutoevaluacion)
	if err != nil {
		return fmt.Errorf("error finding autoevaluacion: %w", err)
	}

	if auto == nil {
		return domain.ErrNotFound
	}

	// Verificar que tenga segmento seleccionado
	if auto.IDSegmento == nil {
		return fmt.Errorf("autoevaluacion must have segmento selected")
	}

	// Obtener respuestas para validar que todas las preguntas fueron respondidas
	respuestas, err := s.respuestaRepo.FindByAutoevaluacion(ctx, idAutoevaluacion)
	if err != nil {
		return fmt.Errorf("error getting respuestas: %w", err)
	}

	// Validación básica: debe haber al menos una respuesta
	if len(respuestas) == 0 {
		return fmt.Errorf("autoevaluacion must have at least one respuesta")
	}

	// === VALIDACIÓN ESTRICTA DE COMPLETITUD ===
	// Obtener los indicadores requeridos para este segmento
	requiredIndicators, err := s.indicadorRepo.FindBySegmento(ctx, *auto.IDSegmento)
	if err != nil {
		return fmt.Errorf("error getting required indicators: %w", err)
	}

	// Verificar que coincida la cantidad
	if len(respuestas) != len(requiredIndicators) {
		return fmt.Errorf("autoevaluacion incomplete: expected %d answers, got %d", len(requiredIndicators), len(respuestas))
	}
	// ==========================================

	// Calcular puntaje total
	puntajeTotal, err := s.respuestaRepo.CalculateTotalScore(ctx, idAutoevaluacion)
	if err != nil {
		return fmt.Errorf("error calculating total score: %w", err)
	}

	// Obtener niveles de sostenibilidad para el segmento
	niveles, err := s.segmentoRepo.FindNivelesSostenibilidadBySegmento(ctx, *auto.IDSegmento)
	if err != nil {
		return fmt.Errorf("error getting niveles sostenibilidad: %w", err)
	}

	// Si no hay niveles configurados, completar sin asignar nivel
	if len(niveles) == 0 {
		// Marcar como completada solo con puntaje (sin nivel de sostenibilidad)
		err = s.autoevaluacionRepo.CompleteWithScore(ctx, idAutoevaluacion, puntajeTotal, 0)
		if err != nil {
			return fmt.Errorf("error completing autoevaluacion: %w", err)
		}
		return nil
	}

	// Determinar el nivel de sostenibilidad según el puntaje
	var nivelAsignado *domain.NivelSostenibilidad
	for _, nivel := range niveles {
		if puntajeTotal >= nivel.MinPuntaje && puntajeTotal <= nivel.MaxPuntaje {
			nivelAsignado = nivel
			break
		}
	}

	// Si no encontró nivel exacto, asignar el más cercano
	if nivelAsignado == nil {
		// Buscar el nivel más alto si el puntaje supera todos los rangos
		for _, nivel := range niveles {
			if puntajeTotal >= nivel.MinPuntaje {
				nivelAsignado = nivel
			}
		}
		// Si aún no hay nivel, usar el primero (puntaje menor al mínimo)
		if nivelAsignado == nil && len(niveles) > 0 {
			nivelAsignado = niveles[0]
		}
	}

	var idNivelSostenibilidad int
	if nivelAsignado != nil {
		idNivelSostenibilidad = nivelAsignado.ID
	}

	// Marcar como completada con puntaje y nivel de sostenibilidad
	err = s.autoevaluacionRepo.CompleteWithScore(ctx, idAutoevaluacion, puntajeTotal, idNivelSostenibilidad)
	if err != nil {
		return fmt.Errorf("error completing autoevaluacion: %w", err)
	}

	return nil
}

// CancelarAutoevaluacion marca la autoevaluación como cancelada
func (s *AutoevaluacionService) CancelarAutoevaluacion(ctx context.Context, idAutoevaluacion int) error {
	// Verificar que la autoevaluación existe
	auto, err := s.autoevaluacionRepo.FindByID(ctx, idAutoevaluacion)
	if err != nil {
		return fmt.Errorf("error finding autoevaluacion: %w", err)
	}

	if auto == nil {
		return domain.ErrNotFound
	}

	// Verificar que esté en estado PENDIENTE
	if auto.Estado != domain.EstadoPendiente {
		return fmt.Errorf("solo se pueden cancelar autoevaluaciones en estado PENDIENTE")
	}

	// Marcar como cancelada
	err = s.autoevaluacionRepo.Cancel(ctx, idAutoevaluacion)
	if err != nil {
		return fmt.Errorf("error canceling autoevaluacion: %w", err)
	}

	return nil
}

// GetHistorialAutoevaluaciones obtiene el historial completo de autoevaluaciones completadas para una bodega
func (s *AutoevaluacionService) GetHistorialAutoevaluaciones(ctx context.Context, idBodega int) ([]domain.HistorialItemResponse, error) {
	autoevaluaciones, err := s.autoevaluacionRepo.FindCompletadasByBodega(ctx, idBodega)
	if err != nil {
		return nil, fmt.Errorf("error getting completadas: %w", err)
	}

	var historial []domain.HistorialItemResponse
	for _, auto := range autoevaluaciones {
		item := domain.HistorialItemResponse{
			IDAutoevaluacion:      auto.ID,
			FechaInicio:           auto.FechaInicio.UTC().Format("2006-01-02T15:04:05Z"),
			Estado:                strings.ToLower(string(auto.Estado)),
			IDBodega:              auto.IDBodega,
			IDSegmento:            auto.IDSegmento,
			PuntajeFinal:          auto.PuntajeFinal,
			IDNivelSostenibilidad: auto.IDNivelSostenibilidad,
		}

		if auto.FechaFin != nil {
			item.FechaFinalizacion = auto.FechaFin.UTC().Format("2006-01-02T15:04:05Z")
		}

		// Obtener nombre del segmento
		if auto.IDSegmento != nil {
			segmento, err := s.segmentoRepo.FindByID(ctx, *auto.IDSegmento)
			if err == nil && segmento != nil {
				item.NombreSegmento = segmento.Nombre
			}

			// Calcular puntaje máximo
			maxPuntos, err := s.nivelRespuestaRepo.FindMaxPuntosBySegmento(ctx, *auto.IDSegmento)
			if err == nil {
				totalMax := 0
				for _, mp := range maxPuntos {
					totalMax += mp
				}
				item.PuntajeMaximo = totalMax

				// Calcular porcentaje
				if totalMax > 0 && auto.PuntajeFinal != nil {
					item.Porcentaje = float64(*auto.PuntajeFinal) / float64(totalMax) * 100
				}
			}
		}

		// Obtener nivel de sostenibilidad
		if auto.IDNivelSostenibilidad != nil && auto.IDSegmento != nil {
			niveles, err := s.segmentoRepo.FindNivelesSostenibilidadBySegmento(ctx, *auto.IDSegmento)
			if err == nil {
				for _, nivel := range niveles {
					if nivel.ID == *auto.IDNivelSostenibilidad {
						item.NivelSostenibilidad = &domain.NivelSostenibilidadInfo{
							ID:     nivel.ID,
							Nombre: nivel.Nombre,
						}
						break
					}
				}
			}
		}

		// Contar indicadores respondidos y total del segmento
		if auto.IDSegmento != nil {
			habilitadosIds, err := s.indicadorRepo.FindBySegmento(ctx, *auto.IDSegmento)
			if err == nil {
				item.IndicadoresTotal = len(habilitadosIds)
			}
		}
		respuestas, err := s.respuestaRepo.FindByAutoevaluacion(ctx, auto.ID)
		if err == nil {
			item.IndicadoresRespondidos = len(respuestas)
		}

		historial = append(historial, item)
	}

	if historial == nil {
		historial = []domain.HistorialItemResponse{}
	}

	return historial, nil
}

// GetResultadosDetallados obtiene los resultados detallados de una autoevaluación con capítulos e indicadores
func (s *AutoevaluacionService) GetResultadosDetallados(ctx context.Context, idAutoevaluacion int) (*domain.ResultadoDetalladoResponse, error) {
	auto, err := s.autoevaluacionRepo.FindByID(ctx, idAutoevaluacion)
	if err != nil {
		return nil, err
	}

	if auto == nil {
		return nil, domain.ErrNotFound
	}

	// Construir la info de la autoevaluación
	autoInfo := domain.HistorialItemResponse{
		IDAutoevaluacion:      auto.ID,
		FechaInicio:           auto.FechaInicio.UTC().Format("2006-01-02T15:04:05Z"),
		Estado:                strings.ToLower(string(auto.Estado)),
		IDBodega:              auto.IDBodega,
		IDSegmento:            auto.IDSegmento,
		PuntajeFinal:          auto.PuntajeFinal,
		IDNivelSostenibilidad: auto.IDNivelSostenibilidad,
	}

	if auto.FechaFin != nil {
		autoInfo.FechaFinalizacion = auto.FechaFin.UTC().Format("2006-01-02T15:04:05Z")
	}

	// Obtener segmento info
	if auto.IDSegmento != nil {
		segmento, err := s.segmentoRepo.FindByID(ctx, *auto.IDSegmento)
		if err == nil && segmento != nil {
			autoInfo.NombreSegmento = segmento.Nombre
		}
	}

	// Calcular puntaje máximo
	var maxPuntosMap map[int]int
	if auto.IDSegmento != nil {
		maxPuntosMap, err = s.nivelRespuestaRepo.FindMaxPuntosBySegmento(ctx, *auto.IDSegmento)
		if err == nil {
			totalMax := 0
			for _, mp := range maxPuntosMap {
				totalMax += mp
			}
			autoInfo.PuntajeMaximo = totalMax
			if totalMax > 0 && auto.PuntajeFinal != nil {
				autoInfo.Porcentaje = float64(*auto.PuntajeFinal) / float64(totalMax) * 100
			}
		}
	}

	// Nivel de sostenibilidad
	if auto.IDNivelSostenibilidad != nil && auto.IDSegmento != nil {
		niveles, err := s.segmentoRepo.FindNivelesSostenibilidadBySegmento(ctx, *auto.IDSegmento)
		if err == nil {
			for _, nivel := range niveles {
				if nivel.ID == *auto.IDNivelSostenibilidad {
					autoInfo.NivelSostenibilidad = &domain.NivelSostenibilidadInfo{
						ID:     nivel.ID,
						Nombre: nivel.Nombre,
					}
					break
				}
			}
		}
	}

	// Obtener respuestas
	respuestas, err := s.respuestaRepo.FindByAutoevaluacion(ctx, idAutoevaluacion)
	if err != nil {
		return nil, fmt.Errorf("error getting respuestas: %w", err)
	}

	// Mapa de id_indicador -> respuesta
	respuestaMap := make(map[int]*domain.Respuesta)
	for _, r := range respuestas {
		respuestaMap[r.IDIndicador] = r
	}

	// Obtener evidencias
	evidencias, _ := s.evidenciaRepo.FindByAutoevaluacion(ctx, idAutoevaluacion)
	evidenciaMap := make(map[int]bool)
	for _, ev := range evidencias {
		evidenciaMap[ev.IDRespuesta] = true
	}

	// Indicadores habilitados para el segmento
	var habilitadosMap map[int]bool
	if auto.IDSegmento != nil {
		habilitadosIds, err := s.indicadorRepo.FindBySegmento(ctx, *auto.IDSegmento)
		if err == nil {
			habilitadosMap = make(map[int]bool)
			for _, id := range habilitadosIds {
				habilitadosMap[id] = true
			}
		}
	}

	// Construir capítulos con indicadores
	capitulos, err := s.capituloRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting capitulos: %w", err)
	}

	var capitulosDetallados []domain.ResultadoCapituloDetallado
	for _, cap := range capitulos {
		indicadores, err := s.indicadorRepo.FindByCapitulo(ctx, cap.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting indicadores for capitulo %d: %w", cap.ID, err)
		}

		capDetallado := domain.ResultadoCapituloDetallado{
			IDCapitulo:  cap.ID,
			Nombre:      cap.Nombre,
			Indicadores: []domain.ResultadoIndicadorDetalle{},
		}

		puntajeObtenido := 0
		puntajeMaximo := 0
		indicadoresCompletados := 0
		indicadoresTotal := 0

		for _, ind := range indicadores {
			// Solo incluir indicadores habilitados para el segmento
			if habilitadosMap != nil && !habilitadosMap[ind.ID] {
				continue
			}

			indicadoresTotal++

			indDetalle := domain.ResultadoIndicadorDetalle{
				IDIndicador: ind.ID,
				Nombre:      ind.Nombre,
				Descripcion: ind.Descripcion,
				Orden:       ind.Orden,
			}

			// Puntaje máximo del indicador
			if maxPuntosMap != nil {
				indDetalle.PuntajeMaximo = maxPuntosMap[ind.ID]
				puntajeMaximo += maxPuntosMap[ind.ID]
			}

			// Respuesta seleccionada
			if resp, ok := respuestaMap[ind.ID]; ok {
				indicadoresCompletados++
				indDetalle.IDRespuesta = &resp.ID
				indDetalle.TieneEvidencia = evidenciaMap[resp.ID]

				// Obtener datos del nivel de respuesta
				nivelResp, err := s.nivelRespuestaRepo.FindByID(ctx, resp.IDNivelRespuesta)
				if err == nil && nivelResp != nil {
					indDetalle.RespuestaNombre = nivelResp.Nombre
					indDetalle.RespuestaDescripcion = nivelResp.Descripcion
					indDetalle.RespuestaPuntos = nivelResp.Puntos
					puntajeObtenido += nivelResp.Puntos
				}
			}

			capDetallado.Indicadores = append(capDetallado.Indicadores, indDetalle)
		}

		capDetallado.PuntajeObtenido = puntajeObtenido
		capDetallado.PuntajeMaximo = puntajeMaximo
		capDetallado.IndicadoresCompletados = indicadoresCompletados
		capDetallado.IndicadoresTotal = indicadoresTotal

		if puntajeMaximo > 0 {
			capDetallado.Porcentaje = float64(puntajeObtenido) / float64(puntajeMaximo) * 100
		}

		// Solo incluir capítulos que tienen indicadores asignados al segmento
		if indicadoresTotal > 0 {
			capitulosDetallados = append(capitulosDetallados, capDetallado)
		}
	}

	// Calcular totales de indicadores respondidos
	totalIndicadoresRespondidos := 0
	totalIndicadores := 0
	for _, cap := range capitulosDetallados {
		totalIndicadoresRespondidos += cap.IndicadoresCompletados
		totalIndicadores += cap.IndicadoresTotal
	}
	autoInfo.IndicadoresRespondidos = totalIndicadoresRespondidos
	autoInfo.IndicadoresTotal = totalIndicadores

	// Obtener el responsable que realizó esta evaluación (guardado al momento de crearla)
	var respInfo *domain.ResponsableInfo
	var resp *domain.Responsable
	if auto.IDResponsable != nil {
		resp, err = s.responsableRepo.FindByID(ctx, *auto.IDResponsable)
	} else {
		// Fallback para evaluaciones antiguas sin responsable guardado
		resp, err = s.responsableRepo.FindActivoByBodega(ctx, auto.IDBodega)
	}
	if err == nil && resp != nil {
		respInfo = &domain.ResponsableInfo{
			Nombre:   resp.Nombre,
			Apellido: resp.Apellido,
			Cargo:    resp.Cargo,
			DNI:      resp.DNI,
		}
	}

	return &domain.ResultadoDetalladoResponse{
		Autoevaluacion: autoInfo,
		Capitulos:      capitulosDetallados,
		Responsable:    respInfo,
	}, nil
}

// GetResultadosBodega obtiene los resultados de la última autoevaluación completada de una bodega
func (s *AutoevaluacionService) GetResultadosBodega(ctx context.Context, idBodega int) (*domain.ResultadosBodegaResponse, error) {
	auto, err := s.autoevaluacionRepo.FindLastCompletadaByBodega(ctx, idBodega)
	if err != nil {
		return nil, fmt.Errorf("error finding last completada: %w", err)
	}
	if auto == nil {
		return nil, domain.ErrNotFound
	}

	response := &domain.ResultadosBodegaResponse{}

	// Info de la autoevaluación
	if auto.FechaFin != nil {
		response.Autoevaluacion.FechaFin = auto.FechaFin.UTC().Format("2006-01-02T15:04:05Z")
	}
	if auto.PuntajeFinal != nil {
		response.Autoevaluacion.PuntajeFinal = *auto.PuntajeFinal
	}

	// Info del segmento
	if auto.IDSegmento != nil {
		segmento, err := s.segmentoRepo.FindByID(ctx, *auto.IDSegmento)
		if err == nil && segmento != nil {
			response.Segmento.Nombre = segmento.Nombre
		}
	}

	// Nivel de sostenibilidad
	if auto.IDNivelSostenibilidad != nil && auto.IDSegmento != nil {
		niveles, err := s.segmentoRepo.FindNivelesSostenibilidadBySegmento(ctx, *auto.IDSegmento)
		if err == nil {
			for _, nivel := range niveles {
				if nivel.ID == *auto.IDNivelSostenibilidad {
					response.NivelSustentabilidad.Nombre = nivel.Nombre
					break
				}
			}
		}
	}

	// Obtener respuestas con sus niveles agrupados por capítulo
	respuestas, err := s.respuestaRepo.FindByAutoevaluacion(ctx, auto.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting respuestas: %w", err)
	}

	respuestaMap := make(map[int]*domain.Respuesta)
	for _, r := range respuestas {
		respuestaMap[r.IDIndicador] = r
	}

	capitulos, err := s.capituloRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting capitulos: %w", err)
	}

	for _, cap := range capitulos {
		capResult := domain.CapituloResultadoSimple{
			Nombre: cap.Nombre,
			Orden:  cap.Orden,
		}

		indicadores, err := s.indicadorRepo.FindByCapitulo(ctx, cap.ID)
		if err != nil {
			continue
		}

		for _, ind := range indicadores {
			indResult := domain.IndicadorResultadoSimple{
				Nombre:      ind.Nombre,
				Descripcion: ind.Descripcion,
				Orden:       ind.Orden,
			}

			// Obtener niveles de respuesta
			niveles, err := s.nivelRespuestaRepo.FindByIndicador(ctx, ind.ID)
			if err == nil {
				for _, nr := range niveles {
					indResult.NivelesRespuesta = append(indResult.NivelesRespuesta, domain.NivelRespuestaResultado{
						Nombre:      nr.Nombre,
						Descripcion: nr.Descripcion,
						Puntos:      nr.Puntos,
					})
				}
			}

			capResult.Indicadores = append(capResult.Indicadores, indResult)
		}

		response.Capitulos = append(response.Capitulos, capResult)
	}

	return response, nil
}
