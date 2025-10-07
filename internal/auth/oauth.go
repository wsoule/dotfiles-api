package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// OAuthState represents a time-limited OAuth state token
type OAuthState struct {
	Token     string
	ExpiresAt time.Time
}

// OAuthService handles OAuth configuration and operations
type OAuthService struct {
	config *oauth2.Config
	states map[string]*OAuthState
	mutex  sync.RWMutex
}

// NewOAuthService creates a new OAuth service
func NewOAuthService() *OAuthService {
	config := &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	service := &OAuthService{
		config: config,
		states: make(map[string]*OAuthState),
	}

	// Start cleanup goroutine
	go service.cleanupExpiredStates()

	return service
}

// generateState generates a cryptographically secure state token
func (s *OAuthService) generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetAuthURL returns the OAuth authorization URL with a unique state token
func (s *OAuthService) GetAuthURL() (string, error) {
	stateToken, err := s.generateState()
	if err != nil {
		return "", err
	}

	// Store state with 10 minute expiration
	s.mutex.Lock()
	s.states[stateToken] = &OAuthState{
		Token:     stateToken,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	s.mutex.Unlock()

	return s.config.AuthCodeURL(stateToken, oauth2.AccessTypeOffline), nil
}

// ExchangeCode exchanges an OAuth code for a token
func (s *OAuthService) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return s.config.Exchange(ctx, code)
}

// GetClient returns an HTTP client for the given token
func (s *OAuthService) GetClient(ctx context.Context, token *oauth2.Token) *http.Client {
	return s.config.Client(ctx, token)
}

// ValidateState validates the OAuth state parameter and removes it
func (s *OAuthService) ValidateState(state string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	oauthState, exists := s.states[state]
	if !exists {
		return false
	}

	// Check if expired
	if time.Now().After(oauthState.ExpiresAt) {
		delete(s.states, state)
		return false
	}

	// Remove state after use (one-time use)
	delete(s.states, state)
	return true
}

// IsConfigured returns true if OAuth is properly configured
func (s *OAuthService) IsConfigured() bool {
	return s.config.ClientID != ""
}

// GetConfig returns the OAuth config (for backward compatibility)
func (s *OAuthService) GetConfig() *oauth2.Config {
	return s.config
}

// cleanupExpiredStates removes expired state tokens periodically
func (s *OAuthService) cleanupExpiredStates() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		s.mutex.Lock()
		for token, state := range s.states {
			if now.After(state.ExpiresAt) {
				delete(s.states, token)
			}
		}
		s.mutex.Unlock()
	}
}