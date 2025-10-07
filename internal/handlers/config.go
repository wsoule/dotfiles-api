package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"dotfiles-api/internal/models"
	"dotfiles-api/internal/repository"
	"dotfiles-api/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ConfigHandler handles config-related HTTP requests
type ConfigHandler struct {
	configRepo repository.ConfigRepository
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(configRepo repository.ConfigRepository) *ConfigHandler {
	return &ConfigHandler{
		configRepo: configRepo,
	}
}

// isAvailable checks if the handler is available (has required dependencies)
func (h *ConfigHandler) isAvailable() bool {
	return h.configRepo != nil
}

// handleUnavailable returns an error response when the feature is not available
func (h *ConfigHandler) handleUnavailable(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"error": errors.NewBadRequestError("Config feature requires MongoDB. Please configure MONGODB_URI environment variable."),
	})
}

// UploadConfig handles config upload
func (h *ConfigHandler) UploadConfig(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	var shareableConfig models.ShareableConfig
	if err := c.ShouldBindJSON(&shareableConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewValidationError("Invalid JSON format"),
		})
		return
	}

	// Validate required fields
	if shareableConfig.Metadata.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewValidationError("Name is required"),
		})
		return
	}

	if shareableConfig.Metadata.Author == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewValidationError("Author is required"),
		})
		return
	}

	// Get user ID from context (if authenticated)
	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	// Set metadata
	shareableConfig.Metadata.CreatedAt = time.Now()
	if shareableConfig.Metadata.Version == "" {
		shareableConfig.Metadata.Version = "1.0.0"
	}

	// Create stored config
	storedConfig := &models.StoredConfig{
		ID:            uuid.New().String(),
		Config:        shareableConfig,
		Public:        true, // Default to public, could be made configurable
		CreatedAt:     time.Now(),
		DownloadCount: 0,
		OwnerID:       userID,
	}

	if err := h.configRepo.Create(c.Request.Context(), storedConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to save config", err),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      storedConfig.ID,
		"message": "Config uploaded successfully",
		"config":  storedConfig,
	})
}

// GetConfig handles getting a config by ID
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("Config ID is required"),
		})
		return
	}

	config, err := h.configRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to retrieve config", err),
		})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": errors.NewNotFoundError("Config"),
		})
		return
	}

	c.JSON(http.StatusOK, config)
}

// DownloadConfig handles config download
func (h *ConfigHandler) DownloadConfig(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("Config ID is required"),
		})
		return
	}

	config, err := h.configRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to retrieve config", err),
		})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": errors.NewNotFoundError("Config"),
		})
		return
	}

	// Increment download count
	if err := h.configRepo.IncrementDownloads(c.Request.Context(), id); err != nil {
		// Log error but don't fail the request
		// In production, you'd use proper logging
	}

	// Return the config content
	c.Header("Content-Disposition", "attachment; filename=dotfiles-config.json")
	c.JSON(http.StatusOK, config.Config)
}

// SearchConfigs handles config search
func (h *ConfigHandler) SearchConfigs(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("Search query is required"),
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// For now, implement a simple search by listing and filtering
	// In production, you'd want proper text search
	configs, err := h.configRepo.List(c.Request.Context(), limit*2, offset) // Get more to filter
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to search configs", err),
		})
		return
	}

	// Simple text search in name and description
	var filtered []*models.StoredConfig
	searchTerm := strings.ToLower(query)
	for _, config := range configs {
		if strings.Contains(strings.ToLower(config.Config.Metadata.Name), searchTerm) ||
			strings.Contains(strings.ToLower(config.Config.Metadata.Description), searchTerm) {
			filtered = append(filtered, config)
			if len(filtered) >= limit {
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": filtered,
		"query":   query,
		"limit":   limit,
		"offset":  offset,
		"total":   len(filtered),
	})
}

// GetFeaturedConfigs handles getting featured configs
func (h *ConfigHandler) GetFeaturedConfigs(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// For now, return most downloaded configs as "featured"
	configs, err := h.configRepo.List(c.Request.Context(), limit, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to get featured configs", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": configs,
		"limit":   limit,
		"total":   len(configs),
	})
}

// GetStats handles getting config statistics
func (h *ConfigHandler) GetStats(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	stats, err := h.configRepo.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to get statistics", err),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
