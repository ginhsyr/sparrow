package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucketState struct {
	tokens   float64
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu             sync.Mutex
	refillPerSec   float64
	burst          float64
	clients        map[string]*bucketState
	requestCounter int
}

func NewIPRateLimiter(perMinute, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		refillPerSec: float64(perMinute) / 60.0,
		burst:        float64(burst),
		clients:      make(map[string]*bucketState),
	}
}

func (l *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if l.allow(c.ClientIP()) {
			c.Next()
			return
		}

		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "rate limit exceeded",
		})
		c.Abort()
	}
}

func (l *IPRateLimiter) allow(clientIP string) bool {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	l.requestCounter++
	if l.requestCounter%200 == 0 {
		l.cleanup(now)
	}

	state, exists := l.clients[clientIP]
	if !exists {
		l.clients[clientIP] = &bucketState{
			tokens:   l.burst - 1,
			lastSeen: now,
		}
		return true
	}

	elapsed := now.Sub(state.lastSeen).Seconds()
	state.tokens += elapsed * l.refillPerSec
	if state.tokens > l.burst {
		state.tokens = l.burst
	}

	state.lastSeen = now
	if state.tokens < 1 {
		return false
	}

	state.tokens--
	return true
}

func (l *IPRateLimiter) cleanup(now time.Time) {
	expireBefore := now.Add(-5 * time.Minute)
	for ip, state := range l.clients {
		if state.lastSeen.Before(expireBefore) {
			delete(l.clients, ip)
		}
	}
}
