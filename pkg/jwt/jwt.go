package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("token inválido")
	ErrExpiredToken = errors.New("token expirado")
)

// Claims representa los claims personalizados del JWT
type Claims struct {
	UserID     int    `json:"user_id"`
	Email      string `json:"email"`
	TipoCuenta string `json:"tipo_cuenta"`
	BodegaID   int    `json:"bodega_id"` // 0 para ADMINISTRADOR_APP
	jwt.RegisteredClaims
}

// GenerateToken genera un nuevo JWT token
func GenerateToken(userID int, email, tipoCuenta, secret string, duration time.Duration, bodegaID int) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:     userID,
		Email:      email,
		TipoCuenta: tipoCuenta,
		BodegaID:   bodegaID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "coviar-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken valida un JWT token y retorna los claims
func ValidateToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verificar que el método de firma sea HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateRefreshToken genera un refresh token con mayor duración
func GenerateRefreshToken(userID int, email, tipoCuenta, secret string, bodegaID int) (string, error) {
	// Refresh token válido por 7 días
	return GenerateToken(userID, email, tipoCuenta, secret, 7*24*time.Hour, bodegaID)
}
