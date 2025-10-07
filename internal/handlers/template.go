package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"dotfiles-api/internal/dto"
	"dotfiles-api/internal/repository"
	"dotfiles-api/pkg/errors"
)

type TemplateHandler struct {
	templateRepo repository.TemplateRepository
}

func NewTemplateHandler(templateRepo repository.TemplateRepository) *TemplateHandler {
	return &TemplateHandler{
		templateRepo: templateRepo,
	}
}

func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var req dto.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewValidationError("invalid request body"),
		})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Template created successfully",
	})
}

func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("template ID is required"),
		})
		return
	}

	template, err := h.templateRepo.GetByID(c.Request.Context(), templateID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, gin.H{"error": appErr})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("failed to get template", err),
		})
		return
	}

	response := &dto.TemplateResponse{
		ID:             template.ID,
		Taps:           template.Template.Taps,
		Brews:          template.Template.Brews,
		Casks:          template.Template.Casks,
		Stow:           template.Template.Stow,
		Extends:        template.Template.Extends,
		Overrides:      template.Template.Overrides,
		AddOnly:        template.Template.AddOnly,
		Public:         template.Template.Public,
		Featured:       template.Template.Featured,
		OrganizationID: template.Template.OrganizationID,
		Downloads:      template.Downloads,
		CreatedAt:      template.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      template.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Metadata: dto.TemplateMetadataResponse{
			Name:        template.Template.Metadata.Name,
			Description: template.Template.Metadata.Description,
			Author:      template.Template.Metadata.Author,
			Version:     template.Template.Metadata.Version,
			Tags:        template.Template.Metadata.Tags,
			CreatedAt:   template.Template.Metadata.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   template.Template.Metadata.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("template ID is required"),
		})
		return
	}

	var req dto.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewValidationError("invalid request body"),
		})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Template updated successfully",
	})
}

func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("template ID is required"),
		})
		return
	}

	err := h.templateRepo.Delete(c.Request.Context(), templateID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, gin.H{"error": appErr})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("failed to delete template", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Template deleted successfully",
	})
}

func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	filters := repository.TemplateFilters{
		Author:         c.Query("author"),
		OrganizationID: c.Query("organization_id"),
		SortBy:         c.DefaultQuery("sort_by", "created_at"),
		SortOrder:      c.DefaultQuery("sort_order", "desc"),
	}

	if tags := c.QueryArray("tags"); len(tags) > 0 {
		filters.Tags = tags
	}

	if featuredStr := c.Query("featured"); featuredStr != "" {
		if featured, err := strconv.ParseBool(featuredStr); err == nil {
			filters.Featured = &featured
		}
	}

	if publicStr := c.Query("public"); publicStr != "" {
		if public, err := strconv.ParseBool(publicStr); err == nil {
			filters.Public = &public
		}
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

	filters.Limit = limit
	filters.Offset = offset

	templates, err := h.templateRepo.List(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("failed to list templates", err),
		})
		return
	}

	response := make([]dto.TemplateResponse, len(templates))
	for i, template := range templates {
		response[i] = dto.TemplateResponse{
			ID:             template.ID,
			Taps:           template.Template.Taps,
			Brews:          template.Template.Brews,
			Casks:          template.Template.Casks,
			Stow:           template.Template.Stow,
			Extends:        template.Template.Extends,
			Overrides:      template.Template.Overrides,
			AddOnly:        template.Template.AddOnly,
			Public:         template.Template.Public,
			Featured:       template.Template.Featured,
			OrganizationID: template.Template.OrganizationID,
			Downloads:      template.Downloads,
			CreatedAt:      template.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:      template.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Metadata: dto.TemplateMetadataResponse{
				Name:        template.Template.Metadata.Name,
				Description: template.Template.Metadata.Description,
				Author:      template.Template.Metadata.Author,
				Version:     template.Template.Metadata.Version,
				Tags:        template.Template.Metadata.Tags,
				CreatedAt:   template.Template.Metadata.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:   template.Template.Metadata.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			},
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": response,
		"limit":     limit,
		"offset":    offset,
		"total":     len(response),
	})
}

func (h *TemplateHandler) SearchTemplates(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("search query is required"),
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

	templates, err := h.templateRepo.Search(c.Request.Context(), query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("failed to search templates", err),
		})
		return
	}

	response := make([]dto.TemplateResponse, len(templates))
	for i, template := range templates {
		response[i] = dto.TemplateResponse{
			ID:             template.ID,
			Taps:           template.Template.Taps,
			Brews:          template.Template.Brews,
			Casks:          template.Template.Casks,
			Stow:           template.Template.Stow,
			Extends:        template.Template.Extends,
			Overrides:      template.Template.Overrides,
			AddOnly:        template.Template.AddOnly,
			Public:         template.Template.Public,
			Featured:       template.Template.Featured,
			OrganizationID: template.Template.OrganizationID,
			Downloads:      template.Downloads,
			CreatedAt:      template.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:      template.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Metadata: dto.TemplateMetadataResponse{
				Name:        template.Template.Metadata.Name,
				Description: template.Template.Metadata.Description,
				Author:      template.Template.Metadata.Author,
				Version:     template.Template.Metadata.Version,
				Tags:        template.Template.Metadata.Tags,
				CreatedAt:   template.Template.Metadata.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:   template.Template.Metadata.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			},
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": response,
		"query":     query,
		"limit":     limit,
		"offset":    offset,
		"total":     len(response),
	})
}

func (h *TemplateHandler) DownloadTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("template ID is required"),
		})
		return
	}

	template, err := h.templateRepo.GetByID(c.Request.Context(), templateID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, gin.H{"error": appErr})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("failed to get template", err),
		})
		return
	}

	err = h.templateRepo.IncrementDownloads(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("failed to increment download count", err),
		})
		return
	}

	c.JSON(http.StatusOK, template.Template)
}

func (h *TemplateHandler) GetTemplateStats(c *gin.Context) {
	stats, err := h.templateRepo.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("failed to get template stats", err),
		})
		return
	}

	response := &dto.TemplateStatsResponse{
		TotalTemplates:    stats.TotalTemplates,
		FeaturedTemplates: stats.FeaturedTemplates,
		TotalDownloads:    stats.TotalDownloads,
		Categories:        stats.Categories,
	}

	c.JSON(http.StatusOK, response)
}

func (h *TemplateHandler) GetTemplateRating(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("template ID is required"),
		})
		return
	}

	rating, err := h.templateRepo.GetRating(c.Request.Context(), templateID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, gin.H{"error": appErr})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("failed to get template rating", err),
		})
		return
	}

	response := &dto.TemplateRatingResponse{
		TemplateID:    rating.TemplateID,
		AverageRating: rating.AverageRating,
		TotalRatings:  rating.TotalRatings,
		Distribution:  rating.Distribution,
	}

	c.JSON(http.StatusOK, response)
}