// Tests unitarios para CuentaService.
//
// Cubre los flujos principales de:
//   - Login: campos vacíos, email inválido, cuenta no encontrada,
//     contraseña incorrecta, login exitoso con y sin bodega
//   - UpdatePassword: contraseña débil, cuenta no existe, actualización exitosa
//
// Para correr solo estos tests:
//
//	cd Backend && go test ./Tests/ -run TestCuenta -v
package tests

import (
	"context"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/service"
	"coviar_backend/Tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newCuentaService crea el servicio con sus mocks para los tests.
func newCuentaService(
	cuentaRepo *mocks.MockCuentaRepository,
	bodegaRepo *mocks.MockBodegaRepository,
) *service.CuentaService {
	return service.NewCuentaService(cuentaRepo, bodegaRepo)
}

// hashPassword genera un bcrypt hash con costo mínimo para que los tests sean rápidos.
func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("no se pudo generar hash para el test: %v", err)
	}
	return string(hash)
}

// ─────────────────────────────────────────────
// Login
// ─────────────────────────────────────────────

// TestLogin_EmailVacio verifica que retorna ErrValidation si el email está vacío.
func TestLogin_EmailVacio(t *testing.T) {
	// Arrange
	svc := newCuentaService(
		new(mocks.MockCuentaRepository),
		new(mocks.MockBodegaRepository),
	)
	req := &domain.CuentaRequest{EmailLogin: "", Password: "clave123"}

	// Act
	result, err := svc.Login(context.Background(), req)

	// Assert
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

// TestLogin_EmailInvalido verifica que retorna ErrValidation si el email
// no tiene formato válido (sin arroba).
func TestLogin_EmailInvalido(t *testing.T) {
	// Arrange
	svc := newCuentaService(
		new(mocks.MockCuentaRepository),
		new(mocks.MockBodegaRepository),
	)
	req := &domain.CuentaRequest{EmailLogin: "no-es-un-email", Password: "clave123"}

	// Act
	result, err := svc.Login(context.Background(), req)

	// Assert
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

// TestLogin_CuentaNoExiste verifica que retorna ErrInvalidCredentials cuando
// el email no corresponde a ninguna cuenta registrada.
// IMPORTANTE: no se debe revelar si el email existe o no (seguridad).
func TestLogin_CuentaNoExiste(t *testing.T) {
	// Arrange
	cuentaRepo := new(mocks.MockCuentaRepository)
	// El repositorio devuelve nil: email no encontrado
	cuentaRepo.On("FindByEmail", mock.Anything, "noexiste@test.com").Return(nil, nil)

	svc := newCuentaService(cuentaRepo, new(mocks.MockBodegaRepository))
	req := &domain.CuentaRequest{EmailLogin: "noexiste@test.com", Password: "clave123"}

	// Act
	result, err := svc.Login(context.Background(), req)

	// Assert
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	cuentaRepo.AssertExpectations(t)
}

// TestLogin_PasswordIncorrecta verifica que retorna ErrInvalidCredentials cuando
// la contraseña no coincide con el hash almacenado.
func TestLogin_PasswordIncorrecta(t *testing.T) {
	// Arrange
	cuentaRepo := new(mocks.MockCuentaRepository)
	cuenta := &domain.Cuenta{
		ID:            1,
		EmailLogin:    "bodega@test.com",
		PasswordHash:  hashPassword(t, "claveCorrecta"),
		FechaRegistro: time.Now(),
	}
	cuentaRepo.On("FindByEmail", mock.Anything, "bodega@test.com").Return(cuenta, nil)

	svc := newCuentaService(cuentaRepo, new(mocks.MockBodegaRepository))
	req := &domain.CuentaRequest{EmailLogin: "bodega@test.com", Password: "claveIncorrecta"}

	// Act
	result, err := svc.Login(context.Background(), req)

	// Assert
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

// TestLogin_OK_SinBodega verifica el flujo exitoso de login para una cuenta
// de tipo ADMINISTRADOR_APP (sin bodega asociada).
func TestLogin_OK_SinBodega(t *testing.T) {
	// Arrange
	cuentaRepo := new(mocks.MockCuentaRepository)
	cuenta := &domain.Cuenta{
		ID:            1,
		Tipo:          domain.TipoCuentaAdministradorApp,
		EmailLogin:    "admin@test.com",
		PasswordHash:  hashPassword(t, "clave123"),
		IDBodega:      nil, // admin no tiene bodega
		FechaRegistro: time.Now(),
	}
	cuentaRepo.On("FindByEmail", mock.Anything, "admin@test.com").Return(cuenta, nil)

	svc := newCuentaService(cuentaRepo, new(mocks.MockBodegaRepository))
	req := &domain.CuentaRequest{EmailLogin: "admin@test.com", Password: "clave123"}

	// Act
	result, err := svc.Login(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, domain.TipoCuentaAdministradorApp, result.Tipo)
	assert.Nil(t, result.Bodega)
	cuentaRepo.AssertExpectations(t)
}

// TestLogin_OK_ConBodega verifica el flujo exitoso de login para una cuenta
// de tipo BODEGA, incluyendo que la bodega se carga correctamente.
func TestLogin_OK_ConBodega(t *testing.T) {
	// Arrange
	cuentaRepo := new(mocks.MockCuentaRepository)
	bodegaRepo := new(mocks.MockBodegaRepository)

	idBodega := 10
	cuenta := &domain.Cuenta{
		ID:            2,
		Tipo:          domain.TipoCuentaBodega,
		EmailLogin:    "bodega@test.com",
		PasswordHash:  hashPassword(t, "clave123"),
		IDBodega:      &idBodega,
		FechaRegistro: time.Now(),
	}
	bodega := &domain.Bodega{ID: idBodega, RazonSocial: "Bodega El Sol"}

	cuentaRepo.On("FindByEmail", mock.Anything, "bodega@test.com").Return(cuenta, nil)
	bodegaRepo.On("FindByID", mock.Anything, idBodega).Return(bodega, nil)

	svc := newCuentaService(cuentaRepo, bodegaRepo)
	req := &domain.CuentaRequest{EmailLogin: "bodega@test.com", Password: "clave123"}

	// Act
	result, err := svc.Login(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.ID)
	assert.NotNil(t, result.Bodega)
	assert.Equal(t, "Bodega El Sol", result.Bodega.RazonSocial)
	cuentaRepo.AssertExpectations(t)
	bodegaRepo.AssertExpectations(t)
}

// ─────────────────────────────────────────────
// UpdatePassword
// ─────────────────────────────────────────────

// TestUpdatePassword_ContrasenaDebil verifica que retorna ErrValidation si la
// nueva contraseña no cumple los requisitos mínimos de seguridad (menos de 8 caracteres).
func TestUpdatePassword_ContrasenaDebil(t *testing.T) {
	// Arrange
	svc := newCuentaService(
		new(mocks.MockCuentaRepository),
		new(mocks.MockBodegaRepository),
	)

	// Act — contraseña corta
	err := svc.UpdatePassword(context.Background(), 1, "abc")

	// Assert
	assert.ErrorIs(t, err, domain.ErrValidation)
}

// TestUpdatePassword_CuentaNoExiste verifica que si la cuenta no existe
// se retorna el error correspondiente.
func TestUpdatePassword_CuentaNoExiste(t *testing.T) {
	// Arrange
	cuentaRepo := new(mocks.MockCuentaRepository)
	cuentaRepo.On("FindByID", mock.Anything, 99).Return(nil, domain.ErrNotFound)

	svc := newCuentaService(cuentaRepo, new(mocks.MockBodegaRepository))

	// Act
	err := svc.UpdatePassword(context.Background(), 99, "NuevaPassword123")

	// Assert
	assert.ErrorIs(t, err, domain.ErrNotFound)
	cuentaRepo.AssertExpectations(t)
}

// TestUpdatePassword_OK verifica el flujo exitoso: la nueva contraseña se hashea
// y se guarda, nunca en texto plano.
func TestUpdatePassword_OK(t *testing.T) {
	// Arrange
	cuentaRepo := new(mocks.MockCuentaRepository)
	cuenta := &domain.Cuenta{
		ID:           1,
		EmailLogin:   "bodega@test.com",
		PasswordHash: hashPassword(t, "claveAntigua"),
	}
	cuentaRepo.On("FindByID", mock.Anything, 1).Return(cuenta, nil)
	cuentaRepo.On("Update", mock.Anything, nil, mock.AnythingOfType("*domain.Cuenta")).Return(nil)

	svc := newCuentaService(cuentaRepo, new(mocks.MockBodegaRepository))

	// Act
	err := svc.UpdatePassword(context.Background(), 1, "NuevaPassword123")

	// Assert
	assert.NoError(t, err)
	// Verificar que el hash guardado NO es la contraseña en texto plano
	assert.NotEqual(t, "NuevaPassword123", cuenta.PasswordHash)
	// Verificar que el nuevo hash es válido para la nueva contraseña
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(cuenta.PasswordHash), []byte("NuevaPassword123")))
	cuentaRepo.AssertExpectations(t)
}
