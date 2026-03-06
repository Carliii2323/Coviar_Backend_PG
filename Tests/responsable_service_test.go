// Tests unitarios para ResponsableService.
//
// Cubre los flujos principales de:
//   - Create: validaciones de campos, verificación de cuenta, creación exitosa
//   - Update: validaciones, responsable no existe, actualización exitosa
//   - DarDeBaja: ya dado de baja, con evaluaciones pendientes sin forzar,
//     con evaluaciones pendientes forzando, flujo exitoso sin pendientes
//
// Para correr solo estos tests:
//
//	cd Backend && go test ./Tests/ -run TestResponsable -v
package tests

import (
	"context"
	"testing"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/service"
	"coviar_backend/Tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newResponsableService crea el servicio con sus mocks para los tests.
func newResponsableService(
	respableRepo *mocks.MockResponsableRepository,
	cuentaRepo *mocks.MockCuentaRepository,
	autoRepo *mocks.MockAutoevaluacionRepository,
) *service.ResponsableService {
	return service.NewResponsableService(respableRepo, cuentaRepo, autoRepo)
}

// ─────────────────────────────────────────────
// Create
// ─────────────────────────────────────────────

// TestCreate_NombreVacio verifica que se retorna ErrValidation si el nombre está vacío.
func TestCreate_NombreVacio(t *testing.T) {
	// Arrange
	svc := newResponsableService(
		new(mocks.MockResponsableRepository),
		new(mocks.MockCuentaRepository),
		new(mocks.MockAutoevaluacionRepository),
	)
	dto := &domain.ResponsableUpdateDTO{
		Nombre:   "", // vacío → debe fallar
		Apellido: "Lopez",
		Cargo:    "Gerente",
		DNI:      "12345678",
	}

	// Act
	result, err := svc.Create(context.Background(), 1, dto)

	// Assert
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

// TestCreate_DNIInvalido verifica que se retorna ErrValidation si el DNI
// no tiene el formato correcto (7-8 dígitos numéricos).
func TestCreate_DNIInvalido(t *testing.T) {
	// Arrange
	svc := newResponsableService(
		new(mocks.MockResponsableRepository),
		new(mocks.MockCuentaRepository),
		new(mocks.MockAutoevaluacionRepository),
	)
	dto := &domain.ResponsableUpdateDTO{
		Nombre:   "Juan",
		Apellido: "Lopez",
		Cargo:    "Gerente",
		DNI:      "abc", // formato inválido
	}

	// Act
	result, err := svc.Create(context.Background(), 1, dto)

	// Assert
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

// TestCreate_CuentaNoExiste verifica que el error del repositorio de cuenta
// se propaga si la cuenta no existe.
func TestCreate_CuentaNoExiste(t *testing.T) {
	// Arrange
	cuentaRepo := new(mocks.MockCuentaRepository)
	cuentaRepo.On("FindByID", mock.Anything, 99).Return(nil, domain.ErrNotFound)

	svc := newResponsableService(
		new(mocks.MockResponsableRepository),
		cuentaRepo,
		new(mocks.MockAutoevaluacionRepository),
	)
	dto := &domain.ResponsableUpdateDTO{
		Nombre:   "Juan",
		Apellido: "Lopez",
		Cargo:    "Gerente",
		DNI:      "12345678",
	}

	// Act
	result, err := svc.Create(context.Background(), 99, dto)

	// Assert
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	cuentaRepo.AssertExpectations(t)
}

// TestCreate_OK verifica el flujo exitoso: valida, normaliza a mayúsculas y crea el responsable.
func TestCreate_OK(t *testing.T) {
	// Arrange
	respableRepo := new(mocks.MockResponsableRepository)
	cuentaRepo := new(mocks.MockCuentaRepository)

	cuenta := &domain.Cuenta{ID: 1, Tipo: domain.TipoCuentaBodega}
	cuentaRepo.On("FindByID", mock.Anything, 1).Return(cuenta, nil)
	// El repo devuelve ID 7 para el nuevo responsable
	respableRepo.On("Create", mock.Anything, nil, mock.AnythingOfType("*domain.Responsable")).Return(7, nil)

	svc := newResponsableService(
		respableRepo,
		cuentaRepo,
		new(mocks.MockAutoevaluacionRepository),
	)
	dto := &domain.ResponsableUpdateDTO{
		Nombre:   "juan",    // será normalizado a "JUAN"
		Apellido: "lopez",   // será normalizado a "LOPEZ"
		Cargo:    "gerente", // será normalizado a "GERENTE"
		DNI:      "12345678",
	}

	// Act
	result, err := svc.Create(context.Background(), 1, dto)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 7, result.ID)
	assert.Equal(t, 1, result.IDCuenta)
	assert.True(t, result.Activo)
	respableRepo.AssertExpectations(t)
	cuentaRepo.AssertExpectations(t)
}

// ─────────────────────────────────────────────
// Update
// ─────────────────────────────────────────────

// TestUpdate_ApellidoVacio verifica que se retorna ErrValidation si el apellido está vacío.
func TestUpdate_ApellidoVacio(t *testing.T) {
	// Arrange
	svc := newResponsableService(
		new(mocks.MockResponsableRepository),
		new(mocks.MockCuentaRepository),
		new(mocks.MockAutoevaluacionRepository),
	)
	dto := &domain.ResponsableUpdateDTO{
		Nombre:   "Juan",
		Apellido: "", // vacío → debe fallar
		Cargo:    "Gerente",
		DNI:      "12345678",
	}

	// Act
	err := svc.Update(context.Background(), 1, dto)

	// Assert
	assert.ErrorIs(t, err, domain.ErrValidation)
}

// TestUpdate_ResponsableNoExiste verifica que si el responsable no existe en la DB
// se retorna el error correspondiente.
func TestUpdate_ResponsableNoExiste(t *testing.T) {
	// Arrange
	respableRepo := new(mocks.MockResponsableRepository)
	respableRepo.On("FindByID", mock.Anything, 99).Return(nil, domain.ErrNotFound)

	svc := newResponsableService(
		respableRepo,
		new(mocks.MockCuentaRepository),
		new(mocks.MockAutoevaluacionRepository),
	)
	dto := &domain.ResponsableUpdateDTO{
		Nombre:   "Juan",
		Apellido: "Lopez",
		Cargo:    "Gerente",
		DNI:      "12345678",
	}

	// Act
	err := svc.Update(context.Background(), 99, dto)

	// Assert
	assert.ErrorIs(t, err, domain.ErrNotFound)
	respableRepo.AssertExpectations(t)
}

// TestUpdate_OK verifica el flujo exitoso de actualización.
func TestUpdate_OK(t *testing.T) {
	// Arrange
	respableRepo := new(mocks.MockResponsableRepository)
	responsable := &domain.Responsable{ID: 1, IDCuenta: 5, Activo: true, DNI: "11111111"}
	respableRepo.On("FindByID", mock.Anything, 1).Return(responsable, nil)
	respableRepo.On("Update", mock.Anything, nil, mock.AnythingOfType("*domain.Responsable")).Return(nil)

	svc := newResponsableService(
		respableRepo,
		new(mocks.MockCuentaRepository),
		new(mocks.MockAutoevaluacionRepository),
	)
	dto := &domain.ResponsableUpdateDTO{
		Nombre:   "Pedro",
		Apellido: "Sanchez",
		Cargo:    "Director",
		DNI:      "22222222",
	}

	// Act
	err := svc.Update(context.Background(), 1, dto)

	// Assert
	assert.NoError(t, err)
	respableRepo.AssertExpectations(t)
}

// ─────────────────────────────────────────────
// DarDeBaja
// ─────────────────────────────────────────────

// TestDarDeBaja_YaDadoDeBaja verifica que retorna ErrResponsableYaDadoDeBaja
// si el responsable ya está inactivo.
func TestDarDeBaja_YaDadoDeBaja(t *testing.T) {
	// Arrange
	respableRepo := new(mocks.MockResponsableRepository)
	// Activo: false → ya fue dado de baja antes
	respableRepo.On("FindByID", mock.Anything, 1).Return(
		&domain.Responsable{ID: 1, Activo: false},
		nil,
	)

	svc := newResponsableService(
		respableRepo,
		new(mocks.MockCuentaRepository),
		new(mocks.MockAutoevaluacionRepository),
	)

	// Act
	err := svc.DarDeBaja(context.Background(), 1, false)

	// Assert
	assert.ErrorIs(t, err, domain.ErrResponsableYaDadoDeBaja)
}

// TestDarDeBaja_ConPendienteSinForzar verifica que si la bodega tiene una
// autoevaluación pendiente y no se fuerza, retorna ErrAutoevaluacionesPendientes.
func TestDarDeBaja_ConPendienteSinForzar(t *testing.T) {
	// Arrange
	respableRepo := new(mocks.MockResponsableRepository)
	cuentaRepo := new(mocks.MockCuentaRepository)
	autoRepo := new(mocks.MockAutoevaluacionRepository)

	idBodega := 10
	respableRepo.On("FindByID", mock.Anything, 1).Return(
		&domain.Responsable{ID: 1, IDCuenta: 5, Activo: true},
		nil,
	)
	cuentaRepo.On("FindByID", mock.Anything, 5).Return(
		&domain.Cuenta{ID: 5, IDBodega: &idBodega},
		nil,
	)
	// Hay una evaluación pendiente en la bodega
	autoRepo.On("HasPendingByBodega", mock.Anything, idBodega).Return(true, nil)

	svc := newResponsableService(respableRepo, cuentaRepo, autoRepo)

	// Act — forzar = false
	err := svc.DarDeBaja(context.Background(), 1, false)

	// Assert
	assert.ErrorIs(t, err, domain.ErrAutoevaluacionesPendientes)
	// No debe haberse llamado a Update porque no se debería dar de baja
	respableRepo.AssertNotCalled(t, "Update")
}

// TestDarDeBaja_ConPendienteForzando verifica que si se fuerza la baja con evaluación
// pendiente, primero la cancela y luego da de baja al responsable.
func TestDarDeBaja_ConPendienteForzando(t *testing.T) {
	// Arrange
	respableRepo := new(mocks.MockResponsableRepository)
	cuentaRepo := new(mocks.MockCuentaRepository)
	autoRepo := new(mocks.MockAutoevaluacionRepository)

	idBodega := 10
	respableRepo.On("FindByID", mock.Anything, 1).Return(
		&domain.Responsable{ID: 1, IDCuenta: 5, Activo: true},
		nil,
	)
	cuentaRepo.On("FindByID", mock.Anything, 5).Return(
		&domain.Cuenta{ID: 5, IDBodega: &idBodega},
		nil,
	)
	autoRepo.On("HasPendingByBodega", mock.Anything, idBodega).Return(true, nil)
	autoRepo.On("FindPendienteByBodega", mock.Anything, idBodega).Return(
		&domain.Autoevaluacion{ID: 55, Estado: domain.EstadoPendiente},
		nil,
	)
	// Primero cancela la evaluación pendiente
	autoRepo.On("Cancel", mock.Anything, 55).Return(nil)
	// Luego da de baja al responsable
	respableRepo.On("Update", mock.Anything, nil, mock.AnythingOfType("*domain.Responsable")).Return(nil)

	svc := newResponsableService(respableRepo, cuentaRepo, autoRepo)

	// Act — forzar = true
	err := svc.DarDeBaja(context.Background(), 1, true)

	// Assert
	assert.NoError(t, err)
	autoRepo.AssertExpectations(t)
	respableRepo.AssertExpectations(t)
}

// TestDarDeBaja_OK verifica el flujo exitoso sin evaluaciones pendientes.
func TestDarDeBaja_OK(t *testing.T) {
	// Arrange
	respableRepo := new(mocks.MockResponsableRepository)
	cuentaRepo := new(mocks.MockCuentaRepository)
	autoRepo := new(mocks.MockAutoevaluacionRepository)

	idBodega := 10
	respableRepo.On("FindByID", mock.Anything, 1).Return(
		&domain.Responsable{ID: 1, IDCuenta: 5, Activo: true},
		nil,
	)
	cuentaRepo.On("FindByID", mock.Anything, 5).Return(
		&domain.Cuenta{ID: 5, IDBodega: &idBodega},
		nil,
	)
	autoRepo.On("HasPendingByBodega", mock.Anything, idBodega).Return(false, nil)
	respableRepo.On("Update", mock.Anything, nil, mock.AnythingOfType("*domain.Responsable")).Return(nil)

	svc := newResponsableService(respableRepo, cuentaRepo, autoRepo)

	// Act
	err := svc.DarDeBaja(context.Background(), 1, false)

	// Assert
	assert.NoError(t, err)
	respableRepo.AssertExpectations(t)
	autoRepo.AssertExpectations(t)
}
