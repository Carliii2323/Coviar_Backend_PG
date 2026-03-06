// Package mocks contiene implementaciones falsas (mocks) de todas las interfaces
// de repositorio definidas en coviar_backend/internal/repository.
//
// Cada mock usa github.com/stretchr/testify/mock para registrar llamadas
// y devolver valores configurados en cada test.
//
// Uso básico en un test:
//
//	repo := new(mocks.MockAutoevaluacionRepository)
//	repo.On("FindByID", mock.Anything, 1).Return(&domain.Autoevaluacion{ID: 1}, nil)
//	// ... llamar al servicio que usa repo ...
//	repo.AssertExpectations(t)
package mocks

import (
	"context"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository"

	"github.com/stretchr/testify/mock"
)

// ─────────────────────────────────────────────
// AutoevaluacionRepository
// ─────────────────────────────────────────────

type MockAutoevaluacionRepository struct {
	mock.Mock
}

func (m *MockAutoevaluacionRepository) Create(ctx context.Context, tx repository.Transaction, auto *domain.Autoevaluacion) (int, error) {
	args := m.Called(ctx, tx, auto)
	return args.Int(0), args.Error(1)
}

func (m *MockAutoevaluacionRepository) FindByID(ctx context.Context, id int) (*domain.Autoevaluacion, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Autoevaluacion), args.Error(1)
}

func (m *MockAutoevaluacionRepository) UpdateSegmento(ctx context.Context, id int, idSegmento int) error {
	args := m.Called(ctx, id, idSegmento)
	return args.Error(0)
}

func (m *MockAutoevaluacionRepository) Complete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAutoevaluacionRepository) FindPendienteByBodega(ctx context.Context, idBodega int) (*domain.Autoevaluacion, error) {
	args := m.Called(ctx, idBodega)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Autoevaluacion), args.Error(1)
}

func (m *MockAutoevaluacionRepository) CompleteWithScore(ctx context.Context, id int, puntajeFinal int, idNivelSostenibilidad int) error {
	args := m.Called(ctx, id, puntajeFinal, idNivelSostenibilidad)
	return args.Error(0)
}

func (m *MockAutoevaluacionRepository) Cancel(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAutoevaluacionRepository) HasPendingByBodega(ctx context.Context, idBodega int) (bool, error) {
	args := m.Called(ctx, idBodega)
	return args.Bool(0), args.Error(1)
}

func (m *MockAutoevaluacionRepository) UpdateEvidenciaStatus(ctx context.Context, id int, estado domain.EstadoEvidencia) error {
	args := m.Called(ctx, id, estado)
	return args.Error(0)
}

func (m *MockAutoevaluacionRepository) FindCompletadasByBodega(ctx context.Context, idBodega int) ([]*domain.Autoevaluacion, error) {
	args := m.Called(ctx, idBodega)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Autoevaluacion), args.Error(1)
}

func (m *MockAutoevaluacionRepository) FindLastCompletadaByBodega(ctx context.Context, idBodega int) (*domain.Autoevaluacion, error) {
	args := m.Called(ctx, idBodega)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Autoevaluacion), args.Error(1)
}

// ─────────────────────────────────────────────
// SegmentoRepository
// ─────────────────────────────────────────────

type MockSegmentoRepository struct {
	mock.Mock
}

func (m *MockSegmentoRepository) FindAll(ctx context.Context) ([]*domain.Segmento, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Segmento), args.Error(1)
}

func (m *MockSegmentoRepository) FindByID(ctx context.Context, id int) (*domain.Segmento, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Segmento), args.Error(1)
}

func (m *MockSegmentoRepository) FindNivelesSostenibilidadBySegmento(ctx context.Context, idSegmento int) ([]*domain.NivelSostenibilidad, error) {
	args := m.Called(ctx, idSegmento)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.NivelSostenibilidad), args.Error(1)
}

// ─────────────────────────────────────────────
// CapituloRepository
// ─────────────────────────────────────────────

type MockCapituloRepository struct {
	mock.Mock
}

func (m *MockCapituloRepository) FindAll(ctx context.Context) ([]*domain.Capitulo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Capitulo), args.Error(1)
}

// ─────────────────────────────────────────────
// IndicadorRepository
// ─────────────────────────────────────────────

type MockIndicadorRepository struct {
	mock.Mock
}

func (m *MockIndicadorRepository) FindByCapitulo(ctx context.Context, idCapitulo int) ([]*domain.Indicador, error) {
	args := m.Called(ctx, idCapitulo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Indicador), args.Error(1)
}

func (m *MockIndicadorRepository) FindBySegmento(ctx context.Context, idSegmento int) ([]int, error) {
	args := m.Called(ctx, idSegmento)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockIndicadorRepository) FindByID(ctx context.Context, id int) (*domain.Indicador, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Indicador), args.Error(1)
}

// ─────────────────────────────────────────────
// NivelRespuestaRepository
// ─────────────────────────────────────────────

type MockNivelRespuestaRepository struct {
	mock.Mock
}

func (m *MockNivelRespuestaRepository) FindByIndicador(ctx context.Context, idIndicador int) ([]*domain.NivelRespuesta, error) {
	args := m.Called(ctx, idIndicador)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.NivelRespuesta), args.Error(1)
}

func (m *MockNivelRespuestaRepository) FindByID(ctx context.Context, id int) (*domain.NivelRespuesta, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.NivelRespuesta), args.Error(1)
}

func (m *MockNivelRespuestaRepository) FindMaxPuntosBySegmento(ctx context.Context, idSegmento int) (map[int]int, error) {
	args := m.Called(ctx, idSegmento)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int]int), args.Error(1)
}

// ─────────────────────────────────────────────
// RespuestaRepository
// ─────────────────────────────────────────────

type MockRespuestaRepository struct {
	mock.Mock
}

func (m *MockRespuestaRepository) Create(ctx context.Context, tx repository.Transaction, respuesta *domain.Respuesta) (int, error) {
	args := m.Called(ctx, tx, respuesta)
	return args.Int(0), args.Error(1)
}

func (m *MockRespuestaRepository) Upsert(ctx context.Context, tx repository.Transaction, respuesta *domain.Respuesta) (int, error) {
	args := m.Called(ctx, tx, respuesta)
	return args.Int(0), args.Error(1)
}

func (m *MockRespuestaRepository) FindByAutoevaluacion(ctx context.Context, idAutoevaluacion int) ([]*domain.Respuesta, error) {
	args := m.Called(ctx, idAutoevaluacion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Respuesta), args.Error(1)
}

func (m *MockRespuestaRepository) DeleteByAutoevaluacion(ctx context.Context, idAutoevaluacion int) error {
	args := m.Called(ctx, idAutoevaluacion)
	return args.Error(0)
}

func (m *MockRespuestaRepository) CalculateTotalScore(ctx context.Context, idAutoevaluacion int) (int, error) {
	args := m.Called(ctx, idAutoevaluacion)
	return args.Int(0), args.Error(1)
}

// ─────────────────────────────────────────────
// EvidenciaRepository
// ─────────────────────────────────────────────

type MockEvidenciaRepository struct {
	mock.Mock
}

func (m *MockEvidenciaRepository) Create(ctx context.Context, tx repository.Transaction, evidencia *domain.Evidencia) (int, error) {
	args := m.Called(ctx, tx, evidencia)
	return args.Int(0), args.Error(1)
}

func (m *MockEvidenciaRepository) FindByRespuesta(ctx context.Context, idRespuesta int) (*domain.Evidencia, error) {
	args := m.Called(ctx, idRespuesta)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Evidencia), args.Error(1)
}

func (m *MockEvidenciaRepository) FindByAutoevaluacion(ctx context.Context, idAutoevaluacion int) ([]*domain.Evidencia, error) {
	args := m.Called(ctx, idAutoevaluacion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Evidencia), args.Error(1)
}

func (m *MockEvidenciaRepository) Delete(ctx context.Context, tx repository.Transaction, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockEvidenciaRepository) CountEvidenciasByAutoevaluacion(ctx context.Context, idAutoevaluacion int) (int, error) {
	args := m.Called(ctx, idAutoevaluacion)
	return args.Int(0), args.Error(1)
}

// ─────────────────────────────────────────────
// ResponsableRepository
// ─────────────────────────────────────────────

type MockResponsableRepository struct {
	mock.Mock
}

func (m *MockResponsableRepository) Create(ctx context.Context, tx repository.Transaction, responsable *domain.Responsable) (int, error) {
	args := m.Called(ctx, tx, responsable)
	return args.Int(0), args.Error(1)
}

func (m *MockResponsableRepository) FindByID(ctx context.Context, id int) (*domain.Responsable, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Responsable), args.Error(1)
}

func (m *MockResponsableRepository) FindByCuentaID(ctx context.Context, cuentaID int) ([]*domain.Responsable, error) {
	args := m.Called(ctx, cuentaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Responsable), args.Error(1)
}

func (m *MockResponsableRepository) FindActivoByBodega(ctx context.Context, idBodega int) (*domain.Responsable, error) {
	args := m.Called(ctx, idBodega)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Responsable), args.Error(1)
}

func (m *MockResponsableRepository) Update(ctx context.Context, tx repository.Transaction, responsable *domain.Responsable) error {
	args := m.Called(ctx, tx, responsable)
	return args.Error(0)
}

func (m *MockResponsableRepository) Delete(ctx context.Context, tx repository.Transaction, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

// ─────────────────────────────────────────────
// CuentaRepository
// ─────────────────────────────────────────────

type MockCuentaRepository struct {
	mock.Mock
}

func (m *MockCuentaRepository) Create(ctx context.Context, tx repository.Transaction, cuenta *domain.Cuenta) (int, error) {
	args := m.Called(ctx, tx, cuenta)
	return args.Int(0), args.Error(1)
}

func (m *MockCuentaRepository) FindByID(ctx context.Context, id int) (*domain.Cuenta, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Cuenta), args.Error(1)
}

func (m *MockCuentaRepository) FindByEmail(ctx context.Context, email string) (*domain.Cuenta, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Cuenta), args.Error(1)
}

func (m *MockCuentaRepository) Update(ctx context.Context, tx repository.Transaction, cuenta *domain.Cuenta) error {
	args := m.Called(ctx, tx, cuenta)
	return args.Error(0)
}

func (m *MockCuentaRepository) Delete(ctx context.Context, tx repository.Transaction, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

// ─────────────────────────────────────────────
// BodegaRepository
// ─────────────────────────────────────────────

type MockBodegaRepository struct {
	mock.Mock
}

func (m *MockBodegaRepository) Create(ctx context.Context, tx repository.Transaction, bodega *domain.Bodega) (int, error) {
	args := m.Called(ctx, tx, bodega)
	return args.Int(0), args.Error(1)
}

func (m *MockBodegaRepository) FindByID(ctx context.Context, id int) (*domain.Bodega, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Bodega), args.Error(1)
}

func (m *MockBodegaRepository) FindByCUIT(ctx context.Context, cuit string) (*domain.Bodega, error) {
	args := m.Called(ctx, cuit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Bodega), args.Error(1)
}

func (m *MockBodegaRepository) Update(ctx context.Context, tx repository.Transaction, bodega *domain.Bodega) error {
	args := m.Called(ctx, tx, bodega)
	return args.Error(0)
}

func (m *MockBodegaRepository) Delete(ctx context.Context, tx repository.Transaction, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockBodegaRepository) GetAll(ctx context.Context) ([]*domain.Bodega, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Bodega), args.Error(1)
}

func (m *MockBodegaRepository) GetAllWithUltimaEval(ctx context.Context) ([]*domain.BodegaAdminItem, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.BodegaAdminItem), args.Error(1)
}
