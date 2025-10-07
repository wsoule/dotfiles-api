package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Session represents a user session
type Session struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Username  string                 `json:"username"`
	Email     string                 `json:"email"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
	Data      map[string]interface{} `json:"data"`
}

// SessionManager manages user sessions
type SessionManager struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
	timeout  time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(timeout time.Duration) *SessionManager {
	manager := &SessionManager{
		sessions: make(map[string]*Session),
		timeout:  timeout,
	}

	// Start cleanup goroutine
	go manager.cleanupExpiredSessions()

	return manager
}

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(userID, username, email string) (*Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		Username:  username,
		Email:     email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(sm.timeout),
		Data:      make(map[string]interface{}),
	}

	sm.mutex.Lock()
	sm.sessions[sessionID] = session
	sm.mutex.Unlock()

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*Session, bool) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, false
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		delete(sm.sessions, sessionID)
		return nil, false
	}

	// Extend session expiry
	session.ExpiresAt = time.Now().Add(sm.timeout)

	return session, true
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mutex.Lock()
	delete(sm.sessions, sessionID)
	sm.mutex.Unlock()
}

// UpdateSession updates session data
func (sm *SessionManager) UpdateSession(sessionID string, data map[string]interface{}) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		for key, value := range data {
			session.Data[key] = value
		}
	}
}

// GetSessionFromContext extracts session from gin context
func (sm *SessionManager) GetSessionFromContext(c *gin.Context) (*Session, bool) {
	// Try to get session ID from cookie
	sessionCookie, err := c.Request.Cookie("session_id")
	if err != nil {
		return nil, false
	}

	return sm.GetSession(sessionCookie.Value)
}

// SetSessionCookie sets the session cookie
func (sm *SessionManager) SetSessionCookie(c *gin.Context, session *Session) {
	// Determine if we're in production (HTTPS) or development
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"

	c.SetCookie(
		"session_id",
		session.ID,
		int(sm.timeout.Seconds()),
		"/",
		"",
		secure, // secure flag - true in production with HTTPS
		true,   // httpOnly
	)
}

// ClearSessionCookie clears the session cookie
func (sm *SessionManager) ClearSessionCookie(c *gin.Context) {
	// Determine if we're in production (HTTPS) or development
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"

	c.SetCookie(
		"session_id",
		"",
		-1,
		"/",
		"",
		secure,
		true,
	)
}

// cleanupExpiredSessions removes expired sessions periodically
func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			sm.mutex.Lock()
			for id, session := range sm.sessions {
				if now.After(session.ExpiresAt) {
					delete(sm.sessions, id)
				}
			}
			sm.mutex.Unlock()
		}
	}
}

// generateSessionID generates a cryptographically secure session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}