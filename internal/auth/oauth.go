package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// OAuthService handles OAuth configuration and operations
type OAuthService struct {
	config     *oauth2.Config
	stateString string
}

// NewOAuthService creates a new OAuth service
func NewOAuthService() *OAuthService {
	// Generate secure random state string
	b := make([]byte, 32)
	rand.Read(b)
	stateString := base64.URLEncoding.EncodeToString(b)

	config := &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	return &OAuthService{
		config:      config,
		stateString: stateString,
	}
}

// GetAuthURL returns the OAuth authorization URL
func (s *OAuthService) GetAuthURL() string {
	return s.config.AuthCodeURL(s.stateString, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges an OAuth code for a token
func (s *OAuthService) ExchangeCode(code string) (*oauth2.Token, error) {
	return s.config.Exchange(context.Background(), code)
}

// GetClient returns an HTTP client for the given token
func (s *OAuthService) GetClient(token *oauth2.Token) *http.Client {
	return s.config.Client(context.Background(), token)
}

// ValidateState validates the OAuth state parameter
func (s *OAuthService) ValidateState(state string) bool {
	return state == s.stateString
}

// IsConfigured returns true if OAuth is properly configured
func (s *OAuthService) IsConfigured() bool {
	return s.config.ClientID != ""
}

// GetConfig returns the OAuth config (for backward compatibility)
func (s *OAuthService) GetConfig() *oauth2.Config {
	return s.config
}