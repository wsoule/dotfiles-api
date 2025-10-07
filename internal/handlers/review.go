package handlers

import (
	"net/http"
	"strconv"
	"time"

	"dotfiles-api/internal/models"
	"dotfiles-api/internal/repository"
	"dotfiles-api/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ReviewHandler handles review-related HTTP requests
type ReviewHandler struct {
	reviewRepo repository.ReviewRepository
}

// NewReviewHandler creates a new review handler
func NewReviewHandler(reviewRepo repository.ReviewRepository) *ReviewHandler {
	return &ReviewHandler{
		reviewRepo: reviewRepo,
	}
}

// isAvailable checks if the handler is available (has required dependencies)
func (h *ReviewHandler) isAvailable() bool {
	return h.reviewRepo != nil
}

// handleUnavailable returns an error response when the feature is not available
func (h *ReviewHandler) handleUnavailable(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"error": errors.NewBadRequestError("Review feature requires MongoDB. Please configure MONGODB_URI environment variable."),
	})
}

// GetTemplateReviews handles getting reviews for a template
func (h *ReviewHandler) GetTemplateReviews(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("Template ID is required"),
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

	reviews, err := h.reviewRepo.GetByTemplate(c.Request.Context(), templateID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to get reviews", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reviews": reviews,
		"limit":   limit,
		"offset":  offset,
	})
}

// CreateReview handles creating a new review
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("Template ID is required"),
		})
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
		Rating  int    `json:"rating" binding:"required,min=1,max=5"`
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewValidationError("Invalid request format"),
		})
		return
	}

	// Check if user already reviewed this template
	existingReview, err := h.reviewRepo.GetUserReviewForTemplate(c.Request.Context(), userID.(string), templateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to check existing review", err),
		})
		return
	}

	if existingReview != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": errors.NewConflictError("User has already reviewed this template"),
		})
		return
	}

	review := &models.Review{
		ID:         uuid.New().String(),
		TemplateID: templateID,
		UserID:     userID.(string),
		Rating:     req.Rating,
		Comment:    req.Comment,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := h.reviewRepo.Create(c.Request.Context(), review); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to create review", err),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"review": review,
		"message": "Review created successfully",
	})
}

// GetTemplateRating handles getting template rating
func (h *ReviewHandler) GetTemplateRating(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("Template ID is required"),
		})
		return
	}

	rating, err := h.reviewRepo.CalculateTemplateRating(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to calculate rating", err),
		})
		return
	}

	c.JSON(http.StatusOK, rating)
}

// UpdateReview handles updating a review
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("Review ID is required"),
		})
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

	// Get existing review
	review, err := h.reviewRepo.GetByID(c.Request.Context(), reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to get review", err),
		})
		return
	}

	if review == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": errors.NewNotFoundError("Review"),
		})
		return
	}

	// Check if user owns the review
	if review.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": errors.NewForbiddenError("Cannot update review owned by another user"),
		})
		return
	}

	var req struct {
		Rating  int    `json:"rating" binding:"required,min=1,max=5"`
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewValidationError("Invalid request format"),
		})
		return
	}

	review.Rating = req.Rating
	review.Comment = req.Comment
	review.UpdatedAt = time.Now()

	if err := h.reviewRepo.Update(c.Request.Context(), review); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to update review", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"review": review,
		"message": "Review updated successfully",
	})
}

// DeleteReview handles deleting a review
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("Review ID is required"),
		})
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

	// Get existing review
	review, err := h.reviewRepo.GetByID(c.Request.Context(), reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to get review", err),
		})
		return
	}

	if review == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": errors.NewNotFoundError("Review"),
		})
		return
	}

	// Check if user owns the review
	if review.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": errors.NewForbiddenError("Cannot delete review owned by another user"),
		})
		return
	}

	if err := h.reviewRepo.Delete(c.Request.Context(), reviewID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to delete review", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Review deleted successfully",
	})
}

// MarkReviewHelpful handles marking a review as helpful
func (h *ReviewHandler) MarkReviewHelpful(c *gin.Context) {
	if !h.isAvailable() {
		h.handleUnavailable(c)
		return
	}

	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("Review ID is required"),
		})
		return
	}

	// Get user ID from context
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": errors.NewUnauthorizedError("Authentication required"),
		})
		return
	}

	// Check if review exists
	review, err := h.reviewRepo.GetByID(c.Request.Context(), reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to get review", err),
		})
		return
	}

	if review == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": errors.NewNotFoundError("Review"),
		})
		return
	}

	if err := h.reviewRepo.IncrementHelpful(c.Request.Context(), reviewID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("Failed to mark review as helpful", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Review marked as helpful",
	})
}
