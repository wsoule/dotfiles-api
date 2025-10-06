package main

import (
	"log"
	"os"
	"time"

	"dotfiles-web/internal/auth"
	"dotfiles-web/internal/handlers"
	"dotfiles-web/internal/middleware"
	"dotfiles-web/internal/repository"
	"dotfiles-web/internal/repository/memory"
	"dotfiles-web/internal/repository/mongo"
	"dotfiles-web/internal/router"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize OAuth service
	oauthService := auth.NewOAuthService()

	// Initialize session manager
	sessionTimeout := 24 * time.Hour // 24 hours
	sessionManager := auth.NewSessionManager(sessionTimeout)

	// Initialize storage
	var mongoClient *mongo.Client
	var err error

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI != "" {
		dbName := os.Getenv("MONGODB_DATABASE")
		if dbName == "" {
			dbName = "dotfiles"
		}

		mongoClient, err = mongo.NewClient(mongoURI, dbName)
		if err != nil {
			log.Printf("Failed to connect to MongoDB: %v", err)
			log.Println("Falling back to memory storage")
		} else {
			log.Println("Connected to MongoDB")
		}
	}

	// Initialize repositories with fallback to in-memory storage
	var configRepo repository.ConfigRepository
	var templateRepo repository.TemplateRepository
	var userRepo repository.UserRepository
	var reviewRepo repository.ReviewRepository
	var orgRepo repository.OrganizationRepository

	if mongoClient != nil {
		configRepo = mongo.NewConfigRepository(mongoClient)
		templateRepo = mongo.NewTemplateRepository(mongoClient)
		userRepo = mongo.NewUserRepository(mongoClient)
		reviewRepo = mongo.NewReviewRepository(mongoClient)
		orgRepo = mongo.NewOrganizationRepository(mongoClient)
		log.Println("Using MongoDB repositories")
	} else {
		// Use in-memory repositories as fallback
		templateRepo = memory.NewTemplateRepository()
		userRepo = memory.NewUserRepository()
		log.Println("Using in-memory repositories (MongoDB not configured)")
		log.Println("Note: Some features (config, reviews, organizations) are not available without MongoDB")
	}

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(sessionManager)

	// Initialize handlers
	configHandler := handlers.NewConfigHandler(configRepo)
	authHandler := handlers.NewAuthHandler(oauthService, sessionManager, userRepo)
	templateHandler := handlers.NewTemplateHandler(templateRepo)
	userHandler := handlers.NewUserHandler(userRepo)
	reviewHandler := handlers.NewReviewHandler(reviewRepo)
	organizationHandler := handlers.NewOrganizationHandler(orgRepo)

	// Initialize router
	appRouter := router.NewRouter(
		configHandler,
		templateHandler,
		userHandler,
		authHandler,
		reviewHandler,
		organizationHandler,
		authMiddleware,
	)

	// Initialize Gin
	r := gin.Default()

	// Add logging middleware
	r.Use(middleware.Logger())

	// Setup routes
	appRouter.SetupRoutes(r)

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}