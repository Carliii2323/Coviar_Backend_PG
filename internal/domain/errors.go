package domain

import "errors"

var (
	ErrNotFound                   = errors.New("Recurso no encontrado")
	ErrEmailYaRegistrado          = errors.New("El email ya está registrado")
	ErrCUITYaRegistrado           = errors.New("El CUIT ya está registrado")
	ErrNoAutorizado               = errors.New("No autorizado")
	ErrCredencialesInvalidas      = errors.New("Credenciales inválidas")
	ErrInvalidCredentials         = errors.New("Credenciales inválidas")
	ErrValidation                 = errors.New("Error de validación")
	ErrAutoevaluacionesPendientes = errors.New("No se puede dar de baja: existen autoevaluaciones pendientes")
	ErrResponsableYaDadoDeBaja    = errors.New("El responsable ya está dado de baja")
	ErrSinResponsable             = errors.New("No hay un responsable activo asignado a esta bodega")
	ErrArchivoInvalido            = errors.New("el archivo no es un PDF válido")
	ErrArchivoDemasiadoGrande     = errors.New("el archivo supera el límite de 2MB")
)
