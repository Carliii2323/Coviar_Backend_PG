package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository"
)

type UsuarioService struct {
	repo repository.UsuarioRepository
}

func NewUsuarioService(repo repository.UsuarioRepository) *UsuarioService {
	return &UsuarioService{repo: repo}
}

func (s *UsuarioService) Create(ctx context.Context, dto *domain.UsuarioDTO) (*domain.Usuario, error) {
	// Validar email
	email := strings.TrimSpace(strings.ToLower(dto.Email))
	if !isValidEmail(email) {
		return nil, fmt.Errorf("email inválido")
	}

	// Validar que el email no exista
	existing, _ := s.repo.FindByEmail(ctx, email)
	if existing != nil {
		return nil, fmt.Errorf("el email ya está registrado")
	}

	// Validar password
	password := strings.TrimSpace(dto.Password)
	if len(password) < 8 {
		return nil, fmt.Errorf("la contraseña debe tener al menos 8 caracteres")
	}

	// Validar nombre y apellido
	if strings.TrimSpace(dto.Nombre) == "" {
		return nil, fmt.Errorf("el nombre es requerido")
	}
	if strings.TrimSpace(dto.Apellido) == "" {
		return nil, fmt.Errorf("el apellido es requerido")
	}

	// Validar rol
	validRoles := map[string]bool{"admin": true, "bodega": true, "auditor": true}
	rol := dto.Rol
	if !validRoles[rol] {
		rol = "bodega" // Rol por defecto
	}

	// Hash de la contraseña
	hashedPassword, err := hashPasswordService(password)
	if err != nil {
		return nil, fmt.Errorf("error al procesar contraseña")
	}

	// Crear usuario
	usuario := &domain.Usuario{
		Email:         email,
		PasswordHash:  hashedPassword,
		Nombre:        strings.TrimSpace(dto.Nombre),
		Apellido:      strings.TrimSpace(dto.Apellido),
		Rol:           rol,
		Activo:        true,
		FechaRegistro: time.Now(),
	}

	id, err := s.repo.Create(ctx, usuario)
	if err != nil {
		return nil, err
	}

	usuario.IdUsuario = id
	return usuario, nil
}

func (s *UsuarioService) Verify(ctx context.Context, login *domain.UsuarioLogin) (*domain.Usuario, error) {
	if login.Email == "" || login.Password == "" {
		return nil, fmt.Errorf("email y contraseña son requeridos")
	}

	email := strings.ToLower(strings.TrimSpace(login.Email))
	password := strings.TrimSpace(login.Password)

	usuario, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("credenciales inválidas")
	}
	if usuario == nil {
		return nil, fmt.Errorf("credenciales inválidas")
	}

	// Verificar que esté activo
	if !usuario.Activo {
		return nil, fmt.Errorf("usuario desactivado")
	}

	// Verificar contraseña
	if err := verifyPasswordService(usuario.PasswordHash, password); err != nil {
		return nil, fmt.Errorf("credenciales inválidas")
	}

	// Actualizar último acceso
	now := time.Now()
	usuario.UltimoAcceso = &now
	_ = s.repo.Update(ctx, usuario)

	return usuario, nil
}

func (s *UsuarioService) GetByID(ctx context.Context, id int) (*domain.Usuario, error) {
	usuario, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if usuario == nil {
		return nil, fmt.Errorf("usuario no encontrado")
	}
	return usuario, nil
}

func (s *UsuarioService) GetAll(ctx context.Context) ([]*domain.Usuario, error) {
	return s.repo.GetAll(ctx)
}

func (s *UsuarioService) Update(ctx context.Context, usuario *domain.Usuario) error {
	return s.repo.Update(ctx, usuario)
}

func (s *UsuarioService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func isValidEmail(email string) bool {
	// Simple email validation
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func hashPasswordService(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func verifyPasswordService(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
