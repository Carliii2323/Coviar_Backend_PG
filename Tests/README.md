# Guía de Tests Unitarios — Backend CoviarPG

## Estructura

```
Backend/Tests/
├── README.md                        ← esta guía
├── mocks/
│   └── mocks.go                     ← mocks de todos los repositorios
├── autoevaluacion_service_test.go   ← tests de AutoevaluacionService
├── responsable_service_test.go      ← tests de ResponsableService
└── cuenta_service_test.go           ← tests de CuentaService
```

---

## Configuración inicial (solo una vez)

### 1. Instalar la librería de testing

Desde la carpeta `Backend/`:

```bash
go get github.com/stretchr/testify
```

Esto agrega `testify` al `go.mod` y `go.sum`.

### 2. Verificar que compila

```bash
cd Backend
go build ./...
```

---

## Cómo correr los tests

### Todos los tests de la carpeta Tests

```bash
cd Backend
go test ./Tests/ -v
```

### Solo un servicio

```bash
go test ./Tests/ -run TestAutoevaluacion -v
go test ./Tests/ -run TestResponsable -v
go test ./Tests/ -run TestCuenta -v
```

### Un test específico

```bash
go test ./Tests/ -run TestLogin_PasswordIncorrecta -v
```

### Con reporte de cobertura

```bash
go test ./Tests/ -coverprofile=coverage.out -v
go tool cover -html=coverage.out
```

---

## Cómo funcionan los tests

### Patrón AAA (Arrange - Act - Assert)

Cada test sigue tres secciones bien marcadas:

```go
func TestCreateAutoevaluacion_SinResponsableActivo(t *testing.T) {
    // Arrange: preparar mocks y servicio
    respableRepo := new(mocks.MockResponsableRepository)
    respableRepo.On("FindActivoByBodega", mock.Anything, 1).Return(nil, nil)
    svc := service.NewAutoevaluacionService(...)

    // Act: llamar al método que se quiere probar
    result, err := svc.CreateAutoevaluacion(context.Background(), 1)

    // Assert: verificar el resultado
    assert.Nil(t, result)
    assert.ErrorIs(t, err, domain.ErrSinResponsable)
    respableRepo.AssertExpectations(t) // verifica que se llamaron los mocks esperados
}
```

### Qué son los mocks

Los mocks reemplazan la base de datos en los tests. En lugar de conectarse a PostgreSQL, los mocks devuelven valores predefinidos.

```go
// Configurar qué devuelve el mock cuando se llame FindByID con id=1
repo.On("FindByID", mock.Anything, 1).Return(&domain.Autoevaluacion{ID: 1}, nil)

// Verificar que FindByID fue llamado exactamente como se configuró
repo.AssertExpectations(t)

// Verificar que un método NO fue llamado
repo.AssertNotCalled(t, "Create")
```

---

## Tests por servicio

### AutoevaluacionService

| Test | Qué verifica |
|------|-------------|
| `TestCreateAutoevaluacion_SinResponsableActivo` | Retorna `ErrSinResponsable` si no hay responsable activo |
| `TestCreateAutoevaluacion_RetornaPendienteExistente` | Retorna la evaluación pendiente existente con sus respuestas |
| `TestCreateAutoevaluacion_CreaNueva` | Crea una nueva evaluación cuando no hay pendiente |
| `TestSeleccionarSegmento_AutoNoExiste` | Retorna `ErrNotFound` si la evaluación no existe |
| `TestSeleccionarSegmento_SegmentoNoExiste` | Retorna `ErrNotFound` si el segmento no existe |
| `TestSeleccionarSegmento_OK` | Guarda segmento y actualiza estado a `SIN_EVIDENCIA` |
| `TestCancelarAutoevaluacion_AutoNoExiste` | Retorna `ErrNotFound` |
| `TestCancelarAutoevaluacion_EstadoInvalido` | No cancela si no está en estado PENDIENTE |
| `TestCancelarAutoevaluacion_OK` | Cancela correctamente |
| `TestCompletarAutoevaluacion_AutoNoExiste` | Retorna `ErrNotFound` |
| `TestCompletarAutoevaluacion_SinSegmento` | Falla si no se seleccionó segmento |
| `TestCompletarAutoevaluacion_RespuestasIncompletas` | Falla si faltan respuestas |
| `TestCompletarAutoevaluacion_OK` | Calcula puntaje y asigna nivel de sostenibilidad |
| `TestGetSegmentos_ErrorDelRepo` | Propaga errores del repositorio |

### ResponsableService

| Test | Qué verifica |
|------|-------------|
| `TestCreate_NombreVacio` | Retorna `ErrValidation` si el nombre está vacío |
| `TestCreate_DNIInvalido` | Retorna `ErrValidation` con DNI de formato inválido |
| `TestCreate_CuentaNoExiste` | Propaga `ErrNotFound` si la cuenta no existe |
| `TestCreate_OK` | Crea responsable y normaliza campos a mayúsculas |
| `TestUpdate_ApellidoVacio` | Retorna `ErrValidation` si el apellido está vacío |
| `TestUpdate_ResponsableNoExiste` | Retorna `ErrNotFound` |
| `TestUpdate_OK` | Actualiza correctamente |
| `TestDarDeBaja_YaDadoDeBaja` | Retorna `ErrResponsableYaDadoDeBaja` si ya está inactivo |
| `TestDarDeBaja_ConPendienteSinForzar` | Retorna `ErrAutoevaluacionesPendientes` |
| `TestDarDeBaja_ConPendienteForzando` | Cancela la evaluación pendiente y da de baja |
| `TestDarDeBaja_OK` | Da de baja sin evaluaciones pendientes |

### CuentaService

| Test | Qué verifica |
|------|-------------|
| `TestLogin_EmailVacio` | Retorna `ErrValidation` si el email está vacío |
| `TestLogin_EmailInvalido` | Retorna `ErrValidation` con email sin formato válido |
| `TestLogin_CuentaNoExiste` | Retorna `ErrInvalidCredentials` (no revela si el email existe) |
| `TestLogin_PasswordIncorrecta` | Retorna `ErrInvalidCredentials` con contraseña incorrecta |
| `TestLogin_OK_SinBodega` | Login exitoso para admin (sin bodega) |
| `TestLogin_OK_ConBodega` | Login exitoso para bodega, carga los datos de la bodega |
| `TestUpdatePassword_ContrasenaDebil` | Retorna `ErrValidation` con contraseña menor a 8 caracteres |
| `TestUpdatePassword_CuentaNoExiste` | Propaga `ErrNotFound` |
| `TestUpdatePassword_OK` | Hashea la nueva contraseña y la guarda (nunca en texto plano) |

---

## Agregar un nuevo test

1. Identificar el método del servicio a probar
2. Crear la función con el nombre `Test<Servicio>_<Escenario>`
3. Configurar los mocks con `.On("Metodo", args...).Return(valores...)`
4. Llamar al método del servicio
5. Verificar con `assert.ErrorIs`, `assert.Equal`, `assert.Nil`, etc.
6. Llamar a `repo.AssertExpectations(t)` al final para confirmar que los mocks se usaron

---

## Agregar mocks para nuevos repositorios

Si se agrega una nueva interfaz en `repository/repository.go`, agregar el mock correspondiente en `Tests/mocks/mocks.go` implementando todos los métodos de la interfaz.
