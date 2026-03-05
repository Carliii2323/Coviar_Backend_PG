package service

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository"
	"coviar_backend/pkg/validator"
)

type CuentaService struct {
	cuentaRepo repository.CuentaRepository
	bodegaRepo repository.BodegaRepository
}

func NewCuentaService(cuentaRepo repository.CuentaRepository, bodegaRepo repository.BodegaRepository) *CuentaService {
	return &CuentaService{
		cuentaRepo: cuentaRepo,
		bodegaRepo: bodegaRepo,
	}
}

type CuentaConBodega struct {
	ID            int               `json:"id_cuenta"`
	Tipo          domain.TipoCuenta `json:"tipo"`
	EmailLogin    string            `json:"email_login"`
	FechaRegistro string            `json:"fecha_registro"`
	Bodega        *domain.Bodega    `json:"bodega,omitempty"`
}

func (s *CuentaService) Login(ctx context.Context, req *domain.CuentaRequest) (*CuentaConBodega, error) {
	// Validar campos requeridos
	if err := validator.ValidateNotEmpty(req.EmailLogin, "email_login"); err != nil {
		return nil, domain.ErrValidation
	}
	if err := validator.ValidateNotEmpty(req.Password, "password"); err != nil {
		return nil, domain.ErrValidation
	}
	if err := validator.ValidateEmail(req.EmailLogin); err != nil {
		return nil, domain.ErrValidation
	}

	cuenta, err := s.cuentaRepo.FindByEmail(ctx, req.EmailLogin)
	if err != nil {
		return nil, fmt.Errorf("error al buscar cuenta: %w", err)
	}
	if cuenta == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(cuenta.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	result := &CuentaConBodega{
		ID:            cuenta.ID,
		Tipo:          cuenta.Tipo,
		EmailLogin:    cuenta.EmailLogin,
		FechaRegistro: cuenta.FechaRegistro.UTC().Format("2006-01-02T15:04:05Z"),
	}

	if cuenta.IDBodega != nil {
		bodega, err := s.bodegaRepo.FindByID(ctx, *cuenta.IDBodega)
		if err == nil && bodega != nil {
			result.Bodega = bodega
		}
	}

	return result, nil
}

func (s *CuentaService) GetByIDWithBodega(ctx context.Context, id int) (*CuentaConBodega, error) {
	cuenta, err := s.cuentaRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &CuentaConBodega{
		ID:            cuenta.ID,
		Tipo:          cuenta.Tipo,
		EmailLogin:    cuenta.EmailLogin,
		FechaRegistro: cuenta.FechaRegistro.UTC().Format("2006-01-02T15:04:05Z"),
	}

	if cuenta.IDBodega != nil {
		bodega, err := s.bodegaRepo.FindByID(ctx, *cuenta.IDBodega)
		if err == nil && bodega != nil {
			result.Bodega = bodega
		}
	}

	return result, nil
}

func (s *CuentaService) UpdatePassword(ctx context.Context, id int, newPassword string) error {
	if err := validator.ValidatePasswordStrength(newPassword); err != nil {
		return domain.ErrValidation
	}

	cuenta, err := s.cuentaRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error al hashear contraseña: %w", err)
	}

	cuenta.PasswordHash = string(hash)
	return s.cuentaRepo.Update(ctx, nil, cuenta)
}
