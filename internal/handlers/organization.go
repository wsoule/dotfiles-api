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

// OrganizationHandler handles organization-related HTTP requests
type OrganizationHandler struct {
	orgRepo repository.OrganizationRepository
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(orgRepo repository.OrganizationRepository) *OrganizationHandler {
	return &OrganizationHandler{
		orgRepo: orgRepo,
	}
}

// isAvailable checks if the handler is available (has required dependencies)
func (h *OrganizationHandler) isAvailable() bool {
	return h.orgRepo != nil
}

// handleUnavailable returns an error response when the feature is not available
func (h *OrganizationHandler) handleUnavailable(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"error": errors.NewBadRequestError("Organization feature requires MongoDB. Please configure MONGODB_URI environment variable."),
	})
}

// CreateOrganization handles creating a new organization
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": errors.NewUnauthorizedError("Authentication required"),
		})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Slug        string `json:"slug" binding:"required"`
		Description string `json:"description"`
		Website     string `json:"website"`
		Public      bool   `json:"public"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewValidationError("Invalid request format"),
		})
		return
	}

	// Validate slug format
	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
	if len(req.Slug) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewValidationError("Slug must be at least 3 characters"),
		})
		return
	}

	// Check if slug already exists
	existing, err := h.orgRepo.GetBySlug(c.Request.Context(), req.Slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to check existing organization", err),
		})
		return
	}

	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": errors.NewConflictError("Organization slug already exists"),
		})
		return
	}

	org := &models.Organization{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Website:     req.Website,
		Public:      req.Public,
		OwnerID:     userID.(string),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.orgRepo.Create(c.Request.Context(), org); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to create organization", err),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"organization": org,
		"message":      "Organization created successfully",
	})
}

// GetOrganizations handles getting all organizations
func (h *OrganizationHandler) GetOrganizations(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
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

	orgs, err := h.orgRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to get organizations", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"organizations": orgs,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetOrganizationBySlug handles getting organization by slug
func (h *OrganizationHandler) GetOrganizationBySlug(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("Organization slug is required"),
		})
		return
	}

	org, err := h.orgRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to get organization", err),
		})
		return
	}

	if org == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": errors.NewNotFoundError("Organization"),
		})
		return
	}

	c.JSON(http.StatusOK, org)
}

// UpdateOrganization handles updating an organization
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

// DeleteOrganization handles deleting an organization
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

// GetOrganizationMembers handles getting organization members
func (h *OrganizationHandler) GetOrganizationMembers(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

// InviteMember handles inviting a member to organization
func (h *OrganizationHandler) InviteMember(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

// RemoveMember handles removing a member from organization
func (h *OrganizationHandler) RemoveMember(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

// UpdateMemberRole handles updating member role
func (h *OrganizationHandler) UpdateMemberRole(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

// GetOrganizationInvites handles getting organization invites
func (h *OrganizationHandler) GetOrganizationInvites(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}

// AcceptInvite handles accepting an organization invite
func (h *OrganizationHandler) AcceptInvite(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented yet"})
}
