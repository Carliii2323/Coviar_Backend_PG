package service

import (
	"context"
	"time"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository"
	"coviar_backend/pkg/validator"
)

type ResponsableService struct {
	responsableRepo    repository.ResponsableRepository
	cuentaRepo         repository.CuentaRepository
	autoevaluacionRepo repository.AutoevaluacionRepository
}

func NewResponsableService(
	responsableRepo repository.ResponsableRepository,
	cuentaRepo repository.CuentaRepository,
	autoevaluacionRepo repository.AutoevaluacionRepository,
) *ResponsableService {
	return &ResponsableService{
		responsableRepo:    responsableRepo,
		cuentaRepo:         cuentaRepo,
		autoevaluacionRepo: autoevaluacionRepo,
	}
}

func (s *ResponsableService) GetByID(ctx context.Context, id int) (*domain.Responsable, error) {
	return s.responsableRepo.FindByID(ctx, id)
}

func (s *ResponsableService) GetByCuentaID(ctx context.Context, cuentaID int) ([]*domain.Responsable, error) {
	return s.responsableRepo.FindByCuentaID(ctx, cuentaID)
}

func (s *ResponsableService) Create(ctx context.Context, cuentaID int, dto *domain.ResponsableUpdateDTO) (*domain.Responsable, error) {
	// Validar campos
	if err := validator.ValidateNotEmpty(dto.Nombre, "nombre"); err != nil {
		return nil, domain.ErrValidation
	}
	if err := validator.ValidateNotEmpty(dto.Apellido, "apellido"); err != nil {
		return nil, domain.ErrValidation
	}
	if err := validator.ValidateNotEmpty(dto.Cargo, "cargo"); err != nil {
		return nil, domain.ErrValidation
	}
	if err := validator.ValidateDNI(dto.DNI); err != nil {
		return nil, domain.ErrValidation
	}

	// Normalizar campos de texto a mayúsculas
	dto.Nombre = validator.NormalizarTexto(dto.Nombre)
	dto.Apellido = validator.NormalizarTexto(dto.Apellido)
	dto.Cargo = validator.NormalizarTexto(dto.Cargo)

	// Verificar que la cuenta existe
	_, err := s.cuentaRepo.FindByID(ctx, cuentaID)
	if err != nil {
		return nil, err
	}

	responsable := &domain.Responsable{
		IDCuenta: cuentaID,
		Nombre:   dto.Nombre,
		Apellido: dto.Apellido,
		Cargo:    dto.Cargo,
		DNI:      dto.DNI,
		Activo:   true,
	}

	id, err := s.responsableRepo.Create(ctx, nil, responsable)
	if err != nil {
		return nil, err
	}

	responsable.ID = id
	return responsable, nil
}

func (s *ResponsableService) Update(ctx context.Context, id int, dto *domain.ResponsableUpdateDTO) error {
	if err := validator.ValidateNotEmpty(dto.Nombre, "nombre"); err != nil {
		return domain.ErrValidation
	}
	if err := validator.ValidateNotEmpty(dto.Apellido, "apellido"); err != nil {
		return domain.ErrValidation
	}
	if err := validator.ValidateNotEmpty(dto.Cargo, "cargo"); err != nil {
		return domain.ErrValidation
	}
	if err := validator.ValidateDNI(dto.DNI); err != nil {
		return domain.ErrValidation
	}

	// Normalizar campos de texto a mayúsculas
	dto.Nombre = validator.NormalizarTexto(dto.Nombre)
	dto.Apellido = validator.NormalizarTexto(dto.Apellido)
	dto.Cargo = validator.NormalizarTexto(dto.Cargo)

	responsable, err := s.responsableRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	responsable.Nombre = dto.Nombre
	responsable.Apellido = dto.Apellido
	responsable.Cargo = dto.Cargo
	responsable.DNI = dto.DNI

	return s.responsableRepo.Update(ctx, nil, responsable)
}

func (s *ResponsableService) DarDeBaja(ctx context.Context, id int, forzar bool) error {
	responsable, err := s.responsableRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if !responsable.Activo {
		return domain.ErrResponsableYaDadoDeBaja
	}

	cuenta, err := s.cuentaRepo.FindByID(ctx, responsable.IDCuenta)
	if err != nil {
		return err
	}

	if cuenta.IDBodega != nil {
		hasPending, err := s.autoevaluacionRepo.HasPendingByBodega(ctx, *cuenta.IDBodega)
		if err != nil {
			return err
		}
		if hasPending {
			if !forzar {
				return domain.ErrAutoevaluacionesPendientes
			}
			// Cancelar la autoevaluación pendiente antes de dar de baja
			pending, err := s.autoevaluacionRepo.FindPendienteByBodega(ctx, *cuenta.IDBodega)
			if err != nil {
				return err
			}
			if pending != nil {
				if err := s.autoevaluacionRepo.Cancel(ctx, pending.ID); err != nil {
					return err
				}
			}
		}
	}

	now := time.Now()
	responsable.Activo = false
	responsable.FechaBaja = &now

	return s.responsableRepo.Update(ctx, nil, responsable)
}

// CanAccess verifica si el usuario autenticado puede acceder al responsable
func (s *ResponsableService) CanAccess(ctx context.Context, responsableID int, userCuentaID int, userTipo string) (bool, error) {
	if userTipo == string(domain.TipoCuentaAdministradorApp) {
		return true, nil
	}

	responsable, err := s.responsableRepo.FindByID(ctx, responsableID)
	if err != nil {
		return false, err
	}

	return responsable.IDCuenta == userCuentaID, nil
}
