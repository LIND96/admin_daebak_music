package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// Burst awal: 10x percobaan boleh rapat (manusia salah ketik dsb).
	loginBurst = 10
	// Isi ulang 1 token tiap 6 detik ≈ 10 percobaan/menit/IP sustained.
	loginRefillEvery = 6 * time.Second
	// Entri IP yang idle lebih lama dari ini dibuang dari memori.
	loginIdleTTL = 10 * time.Minute
)

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
	lastSeen   time.Time
}

type loginLimiter struct {
	mu        sync.Mutex
	buckets   map[string]*tokenBucket
	lastSweep time.Time
}

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{buckets: make(map[string]*tokenBucket), lastSweep: time.Now()}
}

func (l *loginLimiter) allow(ip string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastSweep) > time.Minute {
		for k, b := range l.buckets {
			if now.Sub(b.lastSeen) > loginIdleTTL {
				delete(l.buckets, k)
			}
		}
		l.lastSweep = now
	}

	b, ok := l.buckets[ip]
	if !ok {
		b = &tokenBucket{tokens: loginBurst, lastRefill: now, lastSeen: now}
		l.buckets[ip] = b
	}
	elapsed := now.Sub(b.lastRefill)
	if elapsed > 0 {
		b.tokens += float64(elapsed) / float64(loginRefillEvery)
		if b.tokens > loginBurst {
			b.tokens = loginBurst
		}
		b.lastRefill = now
	}
	b.lastSeen = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

var sharedLoginLimiter = newLoginLimiter()

// LoginRateLimit membatasi percobaan login per IP klien (anti brute-force/DDoS
// endpoint login): burst 10x, lalu ~10x/menit. Melebihi itu dijawab 429.
// PENTING: panggil r.SetTrustedProxies(private ranges) agar c.ClientIP()
// membaca X-Forwarded-For dari Caddy, bukan IP container Caddy-nya.
func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !sharedLoginLimiter.allow(c.ClientIP(), time.Now()) {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many login attempts, try again later"})
			return
		}
		c.Next()
	}
}
