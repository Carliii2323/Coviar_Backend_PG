package domain

import "errors"

var (
	ErrNotFound                   = errors.New("recurso no encontrado")
	ErrEmailYaRegistrado          = errors.New("el email ya está registrado")
	ErrCUITYaRegistrado           = errors.New("el CUIT ya está registrado")
	ErrNoAutorizado               = errors.New("no autorizado")
	ErrCredencialesInvalidas      = errors.New("credenciales inválidas")
	ErrInvalidCredentials         = errors.New("credenciales inválidas")
	ErrValidation                 = errors.New("error de validación")
	ErrAutoevaluacionesPendientes = errors.New("no se puede dar de baja: existen autoevaluaciones pendientes")
	ErrResponsableYaDadoDeBaja    = errors.New("el responsable ya está dado de baja")
	ErrSinResponsable             = errors.New("no hay un responsable activo asignado a esta bodega")
)
