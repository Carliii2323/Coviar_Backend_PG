package tokenblacklist

import (
	"testing"
	"time"
)

func TestIsInactive_FirstRequest(t *testing.T) {
	bl := New()
	// Sin actividad registrada → no inactivo (primer uso tras login)
	if bl.IsInactive("token-nuevo") {
		t.Error("token sin actividad previa no debería considerarse inactivo")
	}
}

func TestIsInactive_RecentActivity(t *testing.T) {
	bl := New()
	token := "token-activo"

	bl.RecordActivity(token)

	if bl.IsInactive(token) {
		t.Error("token con actividad reciente no debería considerarse inactivo")
	}
}

func TestIsInactive_ExpiredActivity(t *testing.T) {
	bl := New()
	token := "token-inactivo"
	h := hash(token)

	// Simular última actividad hace más de 1h
	bl.mu.Lock()
	bl.lastActivity[h] = time.Now().Add(-(InactivityTimeout + time.Minute))
	bl.mu.Unlock()

	if !bl.IsInactive(token) {
		t.Error("token sin actividad por más de 1h debería considerarse inactivo")
	}
}

func TestRecordActivity_ResetsInactivity(t *testing.T) {
	bl := New()
	token := "token-reseteable"
	h := hash(token)

	// Simular que estaba inactivo
	bl.mu.Lock()
	bl.lastActivity[h] = time.Now().Add(-(InactivityTimeout + time.Minute))
	bl.mu.Unlock()

	if !bl.IsInactive(token) {
		t.Fatal("precondición: token debería estar inactivo")
	}

	// Registrar actividad → ya no inactivo
	bl.RecordActivity(token)

	if bl.IsInactive(token) {
		t.Error("después de RecordActivity el token no debería estar inactivo")
	}
}
