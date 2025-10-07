package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"dotfiles-api/pkg/errors"
)

type RateLimiter struct {
	clients map[string]*Client
	mutex   sync.RWMutex
	limit   int
	window  time.Duration
}

type Client struct {
	count     int
	resetTime time.Time
	mutex     sync.Mutex
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*Client),
		limit:   limit,
		window:  window,
	}

	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !rl.allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": errors.NewRateLimitError("rate limit exceeded"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) allow(clientIP string) bool {
	rl.mutex.RLock()
	client, exists := rl.clients[clientIP]
	rl.mutex.RUnlock()

	if !exists {
		client = &Client{
			count:     0,
			resetTime: time.Now().Add(rl.window),
		}
		rl.mutex.Lock()
		rl.clients[clientIP] = client
		rl.mutex.Unlock()
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	now := time.Now()
	if now.After(client.resetTime) {
		client.count = 0
		client.resetTime = now.Add(rl.window)
	}

	if client.count >= rl.limit {
		return false
	}

	client.count++
	return true
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			rl.mutex.Lock()
			for ip, client := range rl.clients {
				client.mutex.Lock()
				if now.After(client.resetTime.Add(rl.window)) {
					delete(rl.clients, ip)
				}
				client.mutex.Unlock()
			}
			rl.mutex.Unlock()
		}
	}
}