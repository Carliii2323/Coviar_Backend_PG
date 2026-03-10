package tokenblacklist

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// InactivityTimeout es el tiempo máximo de inactividad antes de cerrar sesión automáticamente.
const InactivityTimeout = 1 * time.Hour

// Blacklist mantiene en memoria los tokens revocados (logout) y registra
// la última actividad por token para implementar timeout de inactividad.
// Usa hash SHA-256 del token como clave para no almacenar el token en texto plano.
type Blacklist struct {
	mu           sync.RWMutex
	entries      map[string]time.Time // hash → expiresAt original del token (revocados)
	lastActivity map[string]time.Time // hash → última vez que el token hizo un request
}

func New() *Blacklist {
	bl := &Blacklist{
		entries:      make(map[string]time.Time),
		lastActivity: make(map[string]time.Time),
	}
	go bl.cleanup()
	return bl
}

func hash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// Revoke agrega el token a la blacklist hasta su expiración original.
func (bl *Blacklist) Revoke(token string, expiresAt time.Time) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.entries[hash(token)] = expiresAt
}

// IsRevoked retorna true si el token fue revocado y aún estaría vigente.
func (bl *Blacklist) IsRevoked(token string) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	exp, ok := bl.entries[hash(token)]
	if !ok {
		return false
	}
	// Si el token ya expiró naturalmente, no hace falta considerarlo revocado
	return time.Now().Before(exp)
}

// RecordActivity registra que el token acaba de hacer un request activo.
func (bl *Blacklist) RecordActivity(token string) {
	h := hash(token)
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.lastActivity[h] = time.Now()
}

// IsInactive retorna true si el token no ha tenido actividad durante más de InactivityTimeout.
// Un token sin actividad registrada (primer uso tras login) se considera activo.
func (bl *Blacklist) IsInactive(token string) bool {
	h := hash(token)
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	last, ok := bl.lastActivity[h]
	if !ok {
		// Primera request tras login — se registra como activo
		return false
	}
	return time.Since(last) > InactivityTimeout
}

// cleanup elimina entradas cuya expiración original ya pasó y
// registros de actividad de tokens inactivos hace más de 2h.
func (bl *Blacklist) cleanup() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		bl.mu.Lock()
		now := time.Now()
		for k, exp := range bl.entries {
			if now.After(exp) {
				delete(bl.entries, k)
			}
		}
		for k, last := range bl.lastActivity {
			if now.Sub(last) > 2*InactivityTimeout {
				delete(bl.lastActivity, k)
			}
		}
		bl.mu.Unlock()
	}
}
