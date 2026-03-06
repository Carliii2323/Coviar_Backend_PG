package ratelimit

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	MaxAttempts = 5
	LockoutDur  = 15 * time.Minute
	cleanupTick = 10 * time.Minute
)

type entry struct {
	attempts    int
	lockedUntil time.Time
}

type Limiter struct {
	mu      sync.Mutex
	entries map[string]*entry
}

func New() *Limiter {
	l := &Limiter{entries: make(map[string]*entry)}
	go l.cleanup()
	return l
}

func (l *Limiter) cleanup() {
	for range time.Tick(cleanupTick) {
		l.mu.Lock()
		now := time.Now()
		for ip, e := range l.entries {
			if now.After(e.lockedUntil) && e.attempts == 0 {
				delete(l.entries, ip)
			}
		}
		l.mu.Unlock()
	}
}

// IsBlocked informa si la IP está bloqueada por exceso de intentos.
func (l *Limiter) IsBlocked(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.entries[ip]
	if !ok {
		return false
	}
	return time.Now().Before(e.lockedUntil)
}

// RecordFailure registra un intento fallido. Bloquea la IP al alcanzar MaxAttempts.
func (l *Limiter) RecordFailure(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.entries[ip]
	if !ok {
		e = &entry{}
		l.entries[ip] = e
	}
	e.attempts++
	if e.attempts >= MaxAttempts {
		e.lockedUntil = time.Now().Add(LockoutDur)
		e.attempts = 0
	}
}

// RecordSuccess limpia el contador de la IP tras un login exitoso.
func (l *Limiter) RecordSuccess(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, ip)
}

// WindowLimiter es un rate limiter de ventana fija: limita N requests por ventana de tiempo.
type windowEntry struct {
	count       int
	windowStart time.Time
}

type WindowLimiter struct {
	mu      sync.Mutex
	entries map[string]*windowEntry
	limit   int
	window  time.Duration
}

func NewWindowLimiter(limit int, window time.Duration) *WindowLimiter {
	l := &WindowLimiter{
		entries: make(map[string]*windowEntry),
		limit:   limit,
		window:  window,
	}
	go l.windowCleanup()
	return l
}

func (l *WindowLimiter) windowCleanup() {
	for range time.Tick(5 * time.Minute) {
		l.mu.Lock()
		cutoff := time.Now().Add(-l.window)
		for ip, e := range l.entries {
			if e.windowStart.Before(cutoff) {
				delete(l.entries, ip)
			}
		}
		l.mu.Unlock()
	}
}

// Allow retorna true si la request está dentro del límite, false si lo supera.
func (l *WindowLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	e, ok := l.entries[ip]
	if !ok || now.Sub(e.windowStart) >= l.window {
		l.entries[ip] = &windowEntry{count: 1, windowStart: now}
		return true
	}
	e.count++
	return e.count <= l.limit
}

// GetIP obtiene la IP real del request, respetando proxies.
func GetIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}
