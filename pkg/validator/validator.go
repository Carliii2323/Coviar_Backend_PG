package validator

import (
	"fmt"
	"regexp"
	"strings"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var msgs []string
	for _, err := range v {
		msgs = append(msgs, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(msgs, "; ")
}

var (
	cuitRegex     = regexp.MustCompile(`^[0-9]{11}$`)
	dniRegex      = regexp.MustCompile(`^[0-9]{7,8}$`)
	telefonoRegex = regexp.MustCompile(`^[0-9]+$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	invCodeRegex  = regexp.MustCompile(`^[a-zA-Z][0-9]{5}$`)
)

func ValidateCUIT(cuit string) error {
	if !cuitRegex.MatchString(cuit) {
		return fmt.Errorf("CUIT debe tener exactamente 11 dígitos numéricos")
	}
	return nil
}

func ValidateDNI(dni string) error {
	if dni == "" {
		return nil // DNI es opcional
	}
	if !dniRegex.MatchString(dni) {
		return fmt.Errorf("DNI debe tener 7 u 8 dígitos numéricos")
	}
	return nil
}

func ValidateTelefono(telefono string) error {
	if !telefonoRegex.MatchString(telefono) {
		return fmt.Errorf("teléfono solo debe contener dígitos")
	}
	return nil
}

func ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("formato de email inválido")
	}
	return nil
}

func ValidateNotEmpty(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s no puede estar vacío", fieldName)
	}
	return nil
}

func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("la contraseña debe tener al menos 8 caracteres")
	}
	return nil
}


func ValidateInvCode(code *string, fieldName string) error {
	if code == nil || *code == "" {
		return nil
	}

	if len(*code) != 6 {
		return fmt.Errorf("%s debe tener exactamente 6 caracteres", fieldName)
	}

	if !invCodeRegex.MatchString(*code) {
		return fmt.Errorf("%s debe comenzar con una letra seguida de 5 dígitos", fieldName)
	}

	return nil
}

// ============================================
// FUNCIONES DE NORMALIZACIÓN DE TEXTO
// ============================================

// NormalizarTexto convierte el texto a mayúsculas preservando la Ñ y caracteres acentuados.
// Esta función es segura para UTF-8 y respeta los caracteres del idioma español.
func NormalizarTexto(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

// NormalizarTextoSinTildes convierte el texto a mayúsculas y remueve los acentos
// de las vocales, pero preserva la Ñ.
// Usar esta función solo si se requiere normalización para búsquedas.
func NormalizarTextoSinTildes(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	r := strings.NewReplacer(
		"Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U", "Ü", "U",
	)
	return r.Replace(s)
}

// NormalizarPuntero convierte el texto de un puntero a mayúsculas.
// Retorna nil si el puntero es nil o si el valor es vacío.
func NormalizarPuntero(s *string) *string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	normalizado := NormalizarTexto(*s)
	return &normalizado
}
