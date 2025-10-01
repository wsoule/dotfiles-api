package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	OAuth    OAuthConfig    `json:"oauth"`
	Storage  StorageConfig  `json:"storage"`
	Security SecurityConfig `json:"security"`
	Features FeatureConfig  `json:"features"`
}

type ServerConfig struct {
	Port         int           `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	Environment  string        `json:"environment"`
}

type DatabaseConfig struct {
	Type     string `json:"type"`
	MongoDB  MongoDB `json:"mongodb"`
	InMemory bool   `json:"in_memory"`
}

type MongoDB struct {
	URI        string `json:"uri"`
	Database   string `json:"database"`
	Collection string `json:"collection"`
}

type OAuthConfig struct {
	GitHub GitHubOAuth `json:"github"`
}

type GitHubOAuth struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

type StorageConfig struct {
	StaticFiles string `json:"static_files"`
	Templates   string `json:"templates"`
}

type SecurityConfig struct {
	JWTSecret           string        `json:"jwt_secret"`
	SessionTimeout      time.Duration `json:"session_timeout"`
	RateLimitRequests   int           `json:"rate_limit_requests"`
	RateLimitWindow     time.Duration `json:"rate_limit_window"`
	AllowedOrigins      []string      `json:"allowed_origins"`
	InviteTokenExpiry   time.Duration `json:"invite_token_expiry"`
	MaxUploadSize       int64         `json:"max_upload_size"`
	RequireHTTPS        bool          `json:"require_https"`
	EnableCSRFProtection bool         `json:"enable_csrf_protection"`
}

type FeatureConfig struct {
	EnableRegistration    bool `json:"enable_registration"`
	EnableOrganizations   bool `json:"enable_organizations"`
	EnableReviews         bool `json:"enable_reviews"`
	EnableFeaturedContent bool `json:"enable_featured_content"`
	EnableAnalytics       bool `json:"enable_analytics"`
	MaxTemplatesPerUser   int  `json:"max_templates_per_user"`
	MaxOrgsPerUser        int  `json:"max_orgs_per_user"`
}

func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnvAsInt("PORT", 8080),
			Host:         getEnv("HOST", "localhost"),
			ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvAsDuration("IDLE_TIMEOUT", 60*time.Second),
			Environment:  getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			Type: getEnv("DATABASE_TYPE", "memory"),
			MongoDB: MongoDB{
				URI:        getEnv("MONGODB_URI", "mongodb://localhost:27017"),
				Database:   getEnv("MONGODB_DATABASE", "dotfiles"),
				Collection: getEnv("MONGODB_COLLECTION", "templates"),
			},
			InMemory: getEnvAsBool("DATABASE_IN_MEMORY", true),
		},
		OAuth: OAuthConfig{
			GitHub: GitHubOAuth{
				ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
				ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/auth/github/callback"),
			},
		},
		Storage: StorageConfig{
			StaticFiles: getEnv("STATIC_FILES_PATH", "./static"),
			Templates:   getEnv("TEMPLATES_PATH", "./templates"),
		},
		Security: SecurityConfig{
			JWTSecret:             getEnv("JWT_SECRET", "your-secret-key"),
			SessionTimeout:        getEnvAsDuration("SESSION_TIMEOUT", 24*time.Hour),
			RateLimitRequests:     getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitWindow:       getEnvAsDuration("RATE_LIMIT_WINDOW", time.Hour),
			AllowedOrigins:        []string{getEnv("ALLOWED_ORIGINS", "*")},
			InviteTokenExpiry:     getEnvAsDuration("INVITE_TOKEN_EXPIRY", 7*24*time.Hour),
			MaxUploadSize:         getEnvAsInt64("MAX_UPLOAD_SIZE", 10*1024*1024), // 10MB
			RequireHTTPS:          getEnvAsBool("REQUIRE_HTTPS", false),
			EnableCSRFProtection:  getEnvAsBool("ENABLE_CSRF_PROTECTION", true),
		},
		Features: FeatureConfig{
			EnableRegistration:    getEnvAsBool("ENABLE_REGISTRATION", true),
			EnableOrganizations:   getEnvAsBool("ENABLE_ORGANIZATIONS", true),
			EnableReviews:         getEnvAsBool("ENABLE_REVIEWS", true),
			EnableFeaturedContent: getEnvAsBool("ENABLE_FEATURED_CONTENT", true),
			EnableAnalytics:       getEnvAsBool("ENABLE_ANALYTICS", false),
			MaxTemplatesPerUser:   getEnvAsInt("MAX_TEMPLATES_PER_USER", 100),
			MaxOrgsPerUser:        getEnvAsInt("MAX_ORGS_PER_USER", 10),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.OAuth.GitHub.ClientID == "" {
		return fmt.Errorf("GitHub OAuth client ID is required")
	}

	if c.OAuth.GitHub.ClientSecret == "" {
		return fmt.Errorf("GitHub OAuth client secret is required")
	}

	if c.Security.JWTSecret == "" || c.Security.JWTSecret == "your-secret-key" {
		return fmt.Errorf("JWT secret must be set and not use default value")
	}

	if c.Database.Type == "mongodb" && c.Database.MongoDB.URI == "" {
		return fmt.Errorf("MongoDB URI is required when using MongoDB")
	}

	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}