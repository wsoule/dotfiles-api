package router

import (
	"dotfiles-api/internal/handlers"
	"dotfiles-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Router holds all the handlers
type Router struct {
	configHandler       *handlers.ConfigHandler
	templateHandler     *handlers.TemplateHandler
	userHandler         *handlers.UserHandler
	authHandler         *handlers.AuthHandler
	reviewHandler       *handlers.ReviewHandler
	organizationHandler *handlers.OrganizationHandler
	authMiddleware      *middleware.AuthMiddleware
}

// NewRouter creates a new router with all handlers
func NewRouter(
	configHandler *handlers.ConfigHandler,
	templateHandler *handlers.TemplateHandler,
	userHandler *handlers.UserHandler,
	authHandler *handlers.AuthHandler,
	reviewHandler *handlers.ReviewHandler,
	organizationHandler *handlers.OrganizationHandler,
	authMiddleware *middleware.AuthMiddleware,
) *Router {
	return &Router{
		configHandler:       configHandler,
		templateHandler:     templateHandler,
		userHandler:         userHandler,
		authHandler:         authHandler,
		reviewHandler:       reviewHandler,
		organizationHandler: organizationHandler,
		authMiddleware:      authMiddleware,
	}
}

// SetupRoutes configures all the routes
func (router *Router) SetupRoutes(r *gin.Engine) {
	// Add CORS middleware
	r.Use(middleware.CORS([]string{"*"}))

	// API root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Dotfiles API",
			"version": "1.0",
			"endpoints": gin.H{
				"auth":          "/auth",
				"api":           "/api",
				"health":        "/health",
				"documentation": "/docs",
			},
		})
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"service": "dotfiles-api",
		})
	})

	// Authentication routes
	auth := r.Group("/auth")
	{
		auth.GET("/github", router.authHandler.GitHubLogin)
		auth.GET("/github/callback", router.authHandler.GitHubCallback)
		auth.GET("/logout", router.authHandler.Logout)
		auth.GET("/user", router.authHandler.GetCurrentUser)
	}

	// API routes
	api := r.Group("/api")
	{
		// Config endpoints
		api.POST("/configs/upload", router.configHandler.UploadConfig)
		api.GET("/configs/:id", router.configHandler.GetConfig)
		api.GET("/configs/:id/download", router.configHandler.DownloadConfig)
		api.GET("/configs/search", router.configHandler.SearchConfigs)
		api.GET("/configs/featured", router.configHandler.GetFeaturedConfigs)
		api.GET("/configs/stats", router.configHandler.GetStats)

		// Template endpoints
		api.POST("/templates", router.templateHandler.CreateTemplate)
		api.GET("/templates", router.templateHandler.ListTemplates)
		api.GET("/templates/:id", router.templateHandler.GetTemplate)
		api.GET("/templates/:id/download", router.templateHandler.DownloadTemplate)
		api.GET("/templates/:id/reviews", router.reviewHandler.GetTemplateReviews)
		api.POST("/templates/:id/reviews", router.authMiddleware.RequireAuth(), router.reviewHandler.CreateReview)
		api.GET("/templates/:id/rating", router.reviewHandler.GetTemplateRating)

		// User endpoints
		api.GET("/users/:username", router.userHandler.GetUserByUsername)
		api.POST("/users/favorites/:templateId", router.authMiddleware.RequireAuth(), router.userHandler.AddFavorite)
		api.DELETE("/users/favorites/:templateId", router.authMiddleware.RequireAuth(), router.userHandler.RemoveFavorite)

		// Review endpoints
		api.PUT("/reviews/:id", router.authMiddleware.RequireAuth(), router.reviewHandler.UpdateReview)
		api.DELETE("/reviews/:id", router.authMiddleware.RequireAuth(), router.reviewHandler.DeleteReview)
		api.POST("/reviews/:id/helpful", router.authMiddleware.RequireAuth(), router.reviewHandler.MarkReviewHelpful)

		// Organization endpoints
		api.POST("/organizations", router.authMiddleware.RequireAuth(), router.organizationHandler.CreateOrganization)
		api.GET("/organizations", router.organizationHandler.GetOrganizations)
		api.GET("/organizations/:slug", router.organizationHandler.GetOrganizationBySlug)
		api.PUT("/organizations/:slug", router.authMiddleware.RequireAuth(), router.organizationHandler.UpdateOrganization)
		api.DELETE("/organizations/:slug", router.authMiddleware.RequireAuth(), router.organizationHandler.DeleteOrganization)
		api.GET("/organizations/:slug/members", router.organizationHandler.GetOrganizationMembers)
		api.POST("/organizations/:slug/members", router.authMiddleware.RequireAuth(), router.organizationHandler.InviteMember)
		api.DELETE("/organizations/:slug/members/:username", router.authMiddleware.RequireAuth(), router.organizationHandler.RemoveMember)
		api.PUT("/organizations/:slug/members/:username", router.authMiddleware.RequireAuth(), router.organizationHandler.UpdateMemberRole)
		api.GET("/organizations/:slug/invites", router.authMiddleware.RequireAuth(), router.organizationHandler.GetOrganizationInvites)
		api.POST("/invites/:token/accept", router.authMiddleware.RequireAuth(), router.organizationHandler.AcceptInvite)
		api.GET("/users/:username/organizations", router.userHandler.GetUserOrganizations)
	}

	// API documentation endpoint
	r.GET("/docs", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Dotfiles API Documentation",
			"version": "1.0",
			"endpoints": gin.H{
				"auth": gin.H{
					"GET /auth/github":          "GitHub OAuth login",
					"GET /auth/github/callback": "GitHub OAuth callback",
					"GET /auth/logout":          "Logout user",
					"GET /auth/user":            "Get current user",
				},
				"configs": gin.H{
					"POST /api/configs/upload":     "Upload config",
					"GET /api/configs/:id":         "Get config by ID",
					"GET /api/configs/:id/download": "Download config",
					"GET /api/configs/search":      "Search configs",
					"GET /api/configs/featured":    "Get featured configs",
					"GET /api/configs/stats":       "Get config statistics",
				},
				"templates": gin.H{
					"POST /api/templates":              "Create template",
					"GET /api/templates":               "List templates",
					"GET /api/templates/:id":           "Get template by ID",
					"GET /api/templates/:id/download":  "Download template",
					"GET /api/templates/:id/reviews":   "Get template reviews",
					"POST /api/templates/:id/reviews":  "Create review (auth required)",
					"GET /api/templates/:id/rating":    "Get template rating",
				},
				"users": gin.H{
					"GET /api/users/:username":                "Get user profile",
					"POST /api/users/favorites/:templateId":   "Add to favorites (auth required)",
					"DELETE /api/users/favorites/:templateId": "Remove from favorites (auth required)",
				},
				"reviews": gin.H{
					"PUT /api/reviews/:id":        "Update review (auth required)",
					"DELETE /api/reviews/:id":     "Delete review (auth required)",
					"POST /api/reviews/:id/helpful": "Mark review helpful (auth required)",
				},
				"organizations": gin.H{
					"POST /api/organizations":                            "Create organization (auth required)",
					"GET /api/organizations":                             "List organizations",
					"GET /api/organizations/:slug":                       "Get organization by slug",
					"PUT /api/organizations/:slug":                       "Update organization (auth required)",
					"DELETE /api/organizations/:slug":                    "Delete organization (auth required)",
					"GET /api/organizations/:slug/members":               "Get organization members",
					"POST /api/organizations/:slug/members":              "Invite member (auth required)",
					"DELETE /api/organizations/:slug/members/:username":  "Remove member (auth required)",
					"PUT /api/organizations/:slug/members/:username":     "Update member role (auth required)",
					"GET /api/organizations/:slug/invites":               "Get organization invites (auth required)",
					"POST /api/invites/:token/accept":                    "Accept invite (auth required)",
				},
			},
		})
	})
}