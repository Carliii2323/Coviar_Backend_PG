// Tests unitarios para AutoevaluacionService.
//
// Cada función de test sigue la estructura Arrange-Act-Assert (AAA):
//   - Arrange: se preparan los mocks y el servicio
//   - Act: se llama al método del servicio
//   - Assert: se verifica el resultado y que los mocks fueron llamados correctamente
//
// Para correr solo estos tests:
//
//	cd Backend && go test ./Tests/ -run TestAutoevaluacion -v
package tests

import (
	"context"
	"errors"
	"testing"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/service"
	"coviar_backend/Tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newAutoevaluacionService crea el servicio con todos sus mocks para los tests.
func newAutoevaluacionService(
	autoRepo *mocks.MockAutoevaluacionRepository,
	segRepo *mocks.MockSegmentoRepository,
	capRepo *mocks.MockCapituloRepository,
	indRepo *mocks.MockIndicadorRepository,
	nivelRepo *mocks.MockNivelRespuestaRepository,
	respRepo *mocks.MockRespuestaRepository,
	evidRepo *mocks.MockEvidenciaRepository,
	respableRepo *mocks.MockResponsableRepository,
) *service.AutoevaluacionService {
	return service.NewAutoevaluacionService(
		autoRepo, segRepo, capRepo, indRepo, nivelRepo, respRepo, evidRepo, respableRepo,
	)
}

// ─────────────────────────────────────────────
// CreateAutoevaluacion
// ─────────────────────────────────────────────

// TestCreateAutoevaluacion_SinResponsableActivo verifica que si no hay responsable
// activo asignado a la bodega, el servicio retorna ErrSinResponsable.
func TestCreateAutoevaluacion_SinResponsableActivo(t *testing.T) {
	// Arrange
	autoRepo := new(mocks.MockAutoevaluacionRepository)
	respableRepo := new(mocks.MockResponsableRepository)

	// El repo devuelve nil: no hay responsable activo
	respableRepo.On("FindActivoByBodega", mock.Anything, 1).Return(nil, nil)

	svc := newAutoevaluacionService(
		autoRepo,
		new(mocks.MockSegmentoRepository),
		new(mocks.MockCapituloRepository),
		new(mocks.MockIndicadorRepository),
		new(mocks.MockNivelRespuestaRepository),
		new(mocks.MockRespuestaRepository),
		new(mocks.MockEvidenciaRepository),
		respableRepo,
	)

	// Act
	result, err := svc.CreateAutoevaluacion(context.Background(), 1)

	// Assert
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrSinResponsable)
	respableRepo.AssertExpectations(t)
}

// TestCreateAutoevaluacion_RetornaPendienteExistente verifica que si ya existe una
// autoevaluación pendiente para la bodega, el servicio la retorna con sus respuestas
// en lugar de crear una nueva.
func TestCreateAutoevaluacion_RetornaPendienteExistente(t *testing.T) {
	// Arrange
	autoRepo := new(mocks.MockAutoevaluacionRepository)
	respableRepo := new(mocks.MockResponsableRepository)
	respRepo := new(mocks.MockRespuestaRepository)

	responsable := &domain.Responsable{ID: 5, IDCuenta: 10, Activo: true}
	autoPendiente := &domain.Autoevaluacion{ID: 99, IDBodega: 1, Estado: domain.EstadoPendiente}
	respuestasExistentes := []*domain.Respuesta{
		{ID: 1, IDIndicador: 10, IDNivelRespuesta: 2, IDAutoevaluacion: 99},
	}

	respableRepo.On("FindActivoByBodega", mock.Anything, 1).Return(responsable, nil)
	autoRepo.On("FindPendienteByBodega", mock.Anything, 1).Return(autoPendiente, nil)
	respRepo.On("FindByAutoevaluacion", mock.Anything, 99).Return(respuestasExistentes, nil)

	svc := newAutoevaluacionService(
		autoRepo,
		new(mocks.MockSegmentoRepository),
		new(mocks.MockCapituloRepository),
		new(mocks.MockIndicadorRepository),
		new(mocks.MockNivelRespuestaRepository),
		respRepo,
		new(mocks.MockEvidenciaRepository),
		respableRepo,
	)

	// Act
	result, err := svc.CreateAutoevaluacion(context.Background(), 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, autoPendiente, result.AutoevaluacionPendiente)
	assert.Len(t, result.Respuestas, 1)
	// No debe haberse llamado a Create porque ya existía una pendiente
	autoRepo.AssertNotCalled(t, "Create")
	respableRepo.AssertExpectations(t)
	autoRepo.AssertExpectations(t)
}

// TestCreateAutoevaluacion_CreaNueva verifica que cuando no hay autoevaluación pendiente
// y hay responsable activo, se crea una nueva autoevaluación correctamente.
func TestCreateAutoevaluacion_CreaNueva(t *testing.T) {
	// Arrange
	autoRepo := new(mocks.MockAutoevaluacionRepository)
	respableRepo := new(mocks.MockResponsableRepository)
	respRepo := new(mocks.MockRespuestaRepository)

	idResponsable := 5
	responsable := &domain.Responsable{ID: idResponsable, IDCuenta: 10, Activo: true}

	respableRepo.On("FindActivoByBodega", mock.Anything, 1).Return(responsable, nil)
	autoRepo.On("FindPendienteByBodega", mock.Anything, 1).Return(nil, nil)
	// Create devuelve ID 42 para la nueva autoevaluación
	autoRepo.On("Create", mock.Anything, nil, mock.AnythingOfType("*domain.Autoevaluacion")).Return(42, nil)

	svc := newAutoevaluacionService(
		autoRepo,
		new(mocks.MockSegmentoRepository),
		new(mocks.MockCapituloRepository),
		new(mocks.MockIndicadorRepository),
		new(mocks.MockNivelRespuestaRepository),
		respRepo,
		new(mocks.MockEvidenciaRepository),
		respableRepo,
	)

	// Act
	result, err := svc.CreateAutoevaluacion(context.Background(), 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 42, result.AutoevaluacionPendiente.ID)
	assert.Empty(t, result.Respuestas)
	autoRepo.AssertExpectations(t)
	respableRepo.AssertExpectations(t)
}

// ─────────────────────────────────────────────
// SeleccionarSegmento
// ─────────────────────────────────────────────

// TestSeleccionarSegmento_AutoNoExiste verifica que se retorna ErrNotFound
// cuando la autoevaluación no existe.
func TestSeleccionarSegmento_AutoNoExiste(t *testing.T) {
	// Arrange
	autoRepo := new(mocks.MockAutoevaluacionRepository)
	autoRepo.On("FindByID", mock.Anything, 99).Return(nil, nil)

	svc := newAutoevaluacionService(
		autoRepo,
		new(mocks.MockSegmentoRepository),
		new(mocks.MockCapituloRepository),
		new(mocks.MockIndicadorRepository),
		new(mocks.MockNivelRespuestaRepository),
		new(mocks.MockRespuestaRepository),
		new(mocks.MockEvidenciaRepository),
		new(mocks.MockResponsableRepository),
	)

	// Act
	err := svc.SeleccionarSegmento(context.Background(), 99, 1)

	// Assert
	assert.ErrorIs(t, err, domain.ErrNotFound)
	autoRepo.AssertExpectations(t)
}

// TestSeleccionarSegmento_SegmentoNoExiste verifica que se retorna ErrNotFound
// cuando el segmento indicado no existe.
func TestSeleccionarSegmento_SegmentoNoExiste(t *testing.T) {
	// Arrange
	autoRepo := new(mocks.MockAutoevaluacionRepository)
	segRepo := new(mocks.MockSegmentoRepository)

	autoRepo.On("FindByID", mock.Anything, 1).Return(&domain.Autoevaluacion{ID: 1}, nil)
	segRepo.On("FindByID", mock.Anything, 999).Return(nil, nil)

	svc := newAutoevaluacionService(
		autoRepo, segRepo,
		new(mocks.MockCapituloRepository),
		new(mocks.MockIndicadorRepository),
		new(mocks.MockNivelRespuestaRepository),
		new(mocks.MockRespuestaRepository),
		new(mocks.MockEvidenciaRepository),
		new(mocks.MockResponsableRepository),
	)

	// Act
	err := svc.SeleccionarSegmento(context.Background(), 1, 999)

	// Assert
	assert.ErrorIs(t, err, domain.ErrNotFound)
	autoRepo.AssertExpectations(t)
	segRepo.AssertExpectations(t)
}

// TestSeleccionarSegmento_OK verifica el flujo exitoso: guarda el segmento
// y actualiza el estado de evidencia a SIN_EVIDENCIA.
func TestSeleccionarSegmento_OK(t *testing.T) {
	// Arrange
	autoRepo := new(mocks.MockAutoevaluacionRepository)
	segRepo := new(mocks.MockSegmentoRepository)

	autoRepo.On("FindByID", mock.Anything, 1).Return(&domain.Autoevaluacion{ID: 1}, nil)
	segRepo.On("FindByID", mock.Anything, 2).Return(&domain.Segmento{ID: 2, Nombre: "Pequeño"}, nil)
	autoRepo.On("UpdateSegmento", mock.Anything, 1, 2).Return(nil)
	autoRepo.On("UpdateEvidenciaStatus", mock.Anything, 1, domain.EstadoSinEvidencia).Return(nil)

	svc := newAutoevaluacionService(
		autoRepo, segRepo,
		new(mocks.MockCapituloRepository),
		new(mocks.MockIndicadorRepository),
		new(mocks.MockNivelRespuestaRepository),
		new(mocks.MockRespuestaRepository),
		new(mocks.MockEvidenciaRepository),
		new(mocks.MockResponsableRepository),
	)

	// Act
	err := svc.SeleccionarSegmento(context.Background(), 1, 2)

	// Assert
	assert.NoError(t, err)
	autoRepo.AssertExpectations(t)
	segRepo.AssertExpectations(t)
}

// ─────────────────────────────────────────────
// CancelarAutoevaluacion
// ─────────────────────────────────────────────

// TestCancelarAutoevaluacion_AutoNoExiste verifica que se retorna ErrNotFound
// cuando no existe la autoevaluación a cancelar.
func TestCancelarAutoevaluacion_AutoNoExiste(t *testing.T) {
	// Arrange
	autoRepo := new(mocks.MockAutoevaluacionRepository)
	autoRepo.On("FindByID", mock.Anything, 1).Return(nil, nil)

	svc := newAutoevaluacionService(
		autoRepo,
		new(mocks.MockSegmentoRepository),
		new(mocks.MockCapituloRepository),
		new(mocks.MockIndicadorRepository),
		new(mocks.MockNivelRespuestaRepository),
		new(mocks.MockRespuestaRepository),
		new(mocks.MockEvidenciaRepository),
		new(mocks.MockResponsableRepository),
	)

	// Act
	err := svc.CancelarAutoevaluacion(context.Background(), 1)

	// Assert
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// TestCancelarAutoevaluacion_EstadoInvalido verifica que no se puede cancelar
// una autoevaluación que no está en estado PENDIENTE.
func TestCancelarAutoevaluacion_EstadoInvalido(t *testing.T) {
	// Arrange
	autoRepo := new(mocks.MockAutoevaluacionRepository)
	autoCompletada := &domain.Autoevaluacion{ID: 1, Estado: domain.EstadoCompletada}
	autoRepo.On("FindByID", mock.Anything, 1).Return(autoCompletada, nil)

	svc := newAutoevaluacionService(
		autoRepo,
		new(mocks.MockSegmentoRepository),
		new(mocks.MockCapituloRepository),
		new(mocks.MockIndicadorRepository),
		new(mocks.MockNivelRespuestaRepository),
		new(mocks.MockRespuestaRepository),
		new(mocks.MockEvidenciaRepository),
		new(mocks.MockResponsableRepository),
	)

	// Act
	err := svc.CancelarAutoevaluacion(context.Background(), 1)

	// Assert
	assert.Error(t, err)
	// No debe haberse llamado a Cancel
	autoRepo.AssertNotCalled(t, "Cancel")
}

// TestCancelarAutoevaluacion_OK verifica el flujo exitoso de cancelación.
func TestCancelarAutoevaluacion_OK(t *testing.T) {
	// Arrange
	autoRepo := new(mocks.MockAutoevaluacionRepository)
	autoPendiente := &domain.Autoevaluacion{ID: 1, Estado: domain.EstadoPendiente}
	autoRepo.On("FindByID", mock.Anything, 1).Return(autoPendiente, nil)
	autoRepo.On("Cancel", mock.Anything, 1).Return(nil)

	svc := newAutoevaluacionService(
		autoRepo,
		new(mocks.MockSegmentoRepository),
		new(mocks.MockCapituloRepository),
		new(mocks.MockIndicadorRepository),
		new(mocks.MockNivelRespuestaRepository),
		new(mocks.MockRespuestaRepository),
		new(mocks.MockEvidenciaRepository),
		new(mocks.MockResponsableRepository),
	)

	// Act
	err := svc.CancelarAutoevaluacion(context.Background(), 1)

	// Assert
	assert.NoError(t, err)
	autoRepo.AssertExpectations(t)
}

// ─────────────────────────────────────────────
// CompletarAutoevaluacion
// ─────────────────────────────────────────────

// TestCompletarAutoevaluacion_AutoNoExiste verifica que retorna ErrNotFound
// si la autoevaluación no existe.
func TestCompletarAutoevaluacion_AutoNoExiste(t *testing.T) {
	// Arrange
	autoRepo := new(mocks.MockAutoevaluacionRepository)
	autoRepo.On("FindByID", mock.Anything, 1).Return(nil, nil)

	svc := newAutoevaluacionService(
		autoRepo,
		new(mocks.MockSegmentoRepository),
		new(mocks.MockCapituloRepository),
		new(mocks.MockIndicadorRepository),
		new(mocks.MockNivelRespuestaRepository),
		new(mocks.MockRespuestaRepository),
		new(mocks.MockEvidenciaRepository),
		new(mocks.MockResponsableRepository),
	)

	// Act
	err := svc.CompletarAutoevaluacion(context.Background(), 1)

	// Assert
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// TestCompletarAutoevaluacion_SinSegmento verifica que no se puede completar
// una autoevaluación que no tiene segmento seleccionado.
func TestCompletarAutoevaluacion_SinSegmento(t *testing.T) {
	// Arrange
	autoRepo := new(mocks.MockAutoevaluacionRepository)
	// IDSegmento es nil: no se seleccionó segmento
	auto := &domain.Autoevaluacion{ID: 1, Estado: domain.EstadoPendiente, IDSegmento: nil}
	autoRepo.On("FindByID", mock.Anything, 1).Return(auto, nil)

	svc := newAutoevaluacionService(
		autoRepo,
		new(mocks.MockSegmentoRepository),
		new(mocks.MockCapituloRepository),
		new(mocks.MockIndicadorRepository),
		new(mocks.MockNivelRespuestaRepository),
		new(mocks.MockRespuestaRepository),
		new(mocks.MockEvidenciaRepository),
		new(mocks.MockResponsableRepository),
	)

	// Act
	err := svc.CompletarAutoevaluacion(context.Background(), 1)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "segmento")
}

// TestCompletarAutoevaluacion_RespuestasIncompletas verifica que no se puede completar
// si la cantidad de respuestas no coincide con los indicadores requeridos.
func TestCompletarAutoevaluacion_RespuestasIncompletas(t *testing.T) {
	// Arrange
	autoRepo := new(mocks.MockAutoevaluacionRepository)
	indRepo := new(mocks.MockIndicadorRepository)
	respRepo := new(mocks.MockRespuestaRepository)

	idSegmento := 1
	auto := &domain.Autoevaluacion{ID: 1, Estado: domain.EstadoPendiente, IDSegmento: &idSegmento}
	// El segmento requiere 3 indicadores pero solo hay 1 respuesta
	autoRepo.On("FindByID", mock.Anything, 1).Return(auto, nil)
	respRepo.On("FindByAutoevaluacion", mock.Anything, 1).Return([]*domain.Respuesta{
		{ID: 1, IDIndicador: 10, IDNivelRespuesta: 2, IDAutoevaluacion: 1},
	}, nil)
	indRepo.On("FindBySegmento", mock.Anything, 1).Return([]int{10, 11, 12}, nil) // 3 requeridos

	svc := newAutoevaluacionService(
		autoRepo,
		new(mocks.MockSegmentoRepository),
		new(mocks.MockCapituloRepository),
		indRepo,
		new(mocks.MockNivelRespuestaRepository),
		respRepo,
		new(mocks.MockEvidenciaRepository),
		new(mocks.MockResponsableRepository),
	)

	// Act
	err := svc.CompletarAutoevaluacion(context.Background(), 1)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incomplete")
}

// TestCompletarAutoevaluacion_OK verifica el flujo exitoso: calcula puntaje,
// asigna nivel de sostenibilidad y marca como completada.
func TestCompletarAutoevaluacion_OK(t *testing.T) {
	// Arrange
	autoRepo := new(mocks.MockAutoevaluacionRepository)
	indRepo := new(mocks.MockIndicadorRepository)
	respRepo := new(mocks.MockRespuestaRepository)
	segRepo := new(mocks.MockSegmentoRepository)
	nivelRepo := new(mocks.MockNivelRespuestaRepository)

	idSegmento := 1
	auto := &domain.Autoevaluacion{ID: 1, Estado: domain.EstadoPendiente, IDSegmento: &idSegmento}

	autoRepo.On("FindByID", mock.Anything, 1).Return(auto, nil)
	respRepo.On("FindByAutoevaluacion", mock.Anything, 1).Return([]*domain.Respuesta{
		{ID: 1, IDIndicador: 10, IDNivelRespuesta: 2, IDAutoevaluacion: 1},
	}, nil)
	indRepo.On("FindBySegmento", mock.Anything, 1).Return([]int{10}, nil) // 1 requerido, 1 respondido
	respRepo.On("CalculateTotalScore", mock.Anything, 1).Return(50, nil)
	segRepo.On("FindNivelesSostenibilidadBySegmento", mock.Anything, 1).Return([]*domain.NivelSostenibilidad{
		{ID: 3, Nombre: "Medio", MinPuntaje: 30, MaxPuntaje: 70},
	}, nil)
	autoRepo.On("CompleteWithScore", mock.Anything, 1, 50, 3).Return(nil)

	svc := newAutoevaluacionService(
		autoRepo, segRepo,
		new(mocks.MockCapituloRepository),
		indRepo,
		nivelRepo,
		respRepo,
		new(mocks.MockEvidenciaRepository),
		new(mocks.MockResponsableRepository),
	)

	// Act
	err := svc.CompletarAutoevaluacion(context.Background(), 1)

	// Assert
	assert.NoError(t, err)
	autoRepo.AssertExpectations(t)
	respRepo.AssertExpectations(t)
	indRepo.AssertExpectations(t)
}

// ─────────────────────────────────────────────
// GetSegmentos
// ─────────────────────────────────────────────

// TestGetSegmentos_ErrorDelRepo verifica que el error del repositorio
// se propaga correctamente al llamador.
func TestGetSegmentos_ErrorDelRepo(t *testing.T) {
	// Arrange
	segRepo := new(mocks.MockSegmentoRepository)
	errRepo := errors.New("conexión fallida")
	segRepo.On("FindAll", mock.Anything).Return(nil, errRepo)

	svc := newAutoevaluacionService(
		new(mocks.MockAutoevaluacionRepository),
		segRepo,
		new(mocks.MockCapituloRepository),
		new(mocks.MockIndicadorRepository),
		new(mocks.MockNivelRespuestaRepository),
		new(mocks.MockRespuestaRepository),
		new(mocks.MockEvidenciaRepository),
		new(mocks.MockResponsableRepository),
	)

	// Act
	result, err := svc.GetSegmentos(context.Background())

	// Assert
	assert.Nil(t, result)
	assert.Error(t, err)
}
