package service

import (
	"context"
	"fmt"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository"
)

type UbicacionService struct {
	repo repository.UbicacionRepository
}

func NewUbicacionService(repo repository.UbicacionRepository) *UbicacionService {
	return &UbicacionService{repo: repo}
}

// ===== PROVINCIAS =====

func (s *UbicacionService) GetProvincias(ctx context.Context) ([]*domain.Provincia, error) {
	return s.repo.GetProvincias(ctx)
}

func (s *UbicacionService) GetProvinciaByID(ctx context.Context, id int) (*domain.Provincia, error) {
	provincia, err := s.repo.GetProvinciaByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if provincia == nil {
		return nil, fmt.Errorf("provincia no encontrada")
	}
	return provincia, nil
}

// ===== DEPARTAMENTOS =====

func (s *UbicacionService) GetDepartamentos(ctx context.Context) ([]*domain.Departamento, error) {
	return s.repo.GetDepartamentos(ctx)
}

func (s *UbicacionService) GetDepartamentoByID(ctx context.Context, id int) (*domain.Departamento, error) {
	departamento, err := s.repo.GetDepartamentoByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if departamento == nil {
		return nil, fmt.Errorf("departamento no encontrado")
	}
	return departamento, nil
}

func (s *UbicacionService) GetDepartamentosByProvinciaID(ctx context.Context, provinciaID int) ([]*domain.Departamento, error) {
	return s.repo.GetDepartamentosByProvinciaID(ctx, provinciaID)
}

// ===== LOCALIDADES =====

func (s *UbicacionService) GetLocalidades(ctx context.Context) ([]*domain.Localidad, error) {
	return s.repo.GetLocalidades(ctx)
}

func (s *UbicacionService) GetLocalidadByID(ctx context.Context, id int) (*domain.Localidad, error) {
	localidad, err := s.repo.GetLocalidadByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if localidad == nil {
		return nil, fmt.Errorf("localidad no encontrada")
	}
	return localidad, nil
}

func (s *UbicacionService) GetLocalidadesByDepartamentoID(ctx context.Context, departamentoID int) ([]*domain.Localidad, error) {
	return s.repo.GetLocalidadesByDepartamentoID(ctx, departamentoID)
}
