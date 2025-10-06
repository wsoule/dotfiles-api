package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"dotfiles-web/internal/auth"
	"dotfiles-web/internal/models"
	"dotfiles-web/internal/repository"
	"dotfiles-web/pkg/errors"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	oauthService   *auth.OAuthService
	sessionManager *auth.SessionManager
	userRepo       repository.UserRepository
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(oauthService *auth.OAuthService, sessionManager *auth.SessionManager, userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		oauthService:   oauthService,
		sessionManager: sessionManager,
		userRepo:       userRepo,
	}
}

// GitHubLogin handles GitHub OAuth login
func (h *AuthHandler) GitHubLogin(c *gin.Context) {
	// Check if OAuth is configured
	if !h.oauthService.IsConfigured() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "GitHub OAuth not configured",
			"message": "Please set GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET, and OAUTH_REDIRECT_URL environment variables to enable GitHub authentication.",
		})
		return
	}

	url := h.oauthService.GetAuthURL()
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitHubCallback handles GitHub OAuth callback
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	state := c.Query("state")
	if !h.oauthService.ValidateState(state) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("Invalid OAuth state"),
		})
		return
	}

	code := c.Query("code")
	token, err := h.oauthService.ExchangeCode(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("Failed to exchange OAuth code"),
		})
		return
	}

	// Get user info from GitHub
	client := h.oauthService.GetClient(token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to get user info from GitHub", err),
		})
		return
	}
	defer resp.Body.Close()

	var githubUser struct {
		ID        int    `json:"id"`
		Username  string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
		Bio       string `json:"bio"`
		Location  string `json:"location"`
		Website   string `json:"blog"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to decode GitHub user data", err),
		})
		return
	}

	// Check if user already exists
	user, err := h.userRepo.GetByGitHubID(c.Request.Context(), githubUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to check existing user", err),
		})
		return
	}

	// Create or update user
	if user == nil {
		user = &models.User{
			GitHubID:    githubUser.ID,
			Username:    githubUser.Username,
			Name:        githubUser.Name,
			Email:       githubUser.Email,
			AvatarURL:   githubUser.AvatarURL,
			Bio:         githubUser.Bio,
			Location:    githubUser.Location,
			Website:     githubUser.Website,
			Favorites:   []string{},
			Collections: []string{},
		}

		if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": errors.NewInternalError("Failed to create user", err),
			})
			return
		}
	} else {
		// Update existing user info
		user.Name = githubUser.Name
		user.Email = githubUser.Email
		user.AvatarURL = githubUser.AvatarURL
		user.Bio = githubUser.Bio
		user.Location = githubUser.Location
		user.Website = githubUser.Website

		if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": errors.NewInternalError("Failed to update user", err),
			})
			return
		}
	}

	// Create session
	session, err := h.sessionManager.CreateSession(user.ID, user.Username, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to create session", err),
		})
		return
	}

	// Set session cookie
	h.sessionManager.SetSessionCookie(c, session)

	// Redirect to frontend or return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Authentication successful",
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"name":       user.Name,
			"email":      user.Email,
			"avatar_url": user.AvatarURL,
		},
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	session, exists := h.sessionManager.GetSessionFromContext(c)
	if exists {
		h.sessionManager.DeleteSession(session.ID)
	}

	h.sessionManager.ClearSessionCookie(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// GetCurrentUser handles getting current user info
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Check if OAuth is configured
	if !h.oauthService.IsConfigured() {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":       "GitHub OAuth not configured",
			"configured":  false,
			"message":     "Authentication is not available. Please configure GitHub OAuth to enable user features.",
		})
		return
	}

	session, exists := h.sessionManager.GetSessionFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "Not authenticated",
			"configured": true,
		})
		return
	}

	// Get full user details
	user, err := h.userRepo.GetByID(c.Request.Context(), session.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to get user details", err),
		})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": errors.NewUnauthorizedError("User not found"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"name":       user.Name,
			"email":      user.Email,
			"avatar_url": user.AvatarURL,
			"bio":        user.Bio,
			"location":   user.Location,
			"website":    user.Website,
			"created_at": user.CreatedAt.Format(time.RFC3339),
		},
		"configured": true,
	})
}