package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"dotfiles-web/internal/dto"
	"dotfiles-web/internal/repository"
	"dotfiles-web/pkg/errors"
)

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
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
		"message": "User created successfully",
	})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("user ID is required"),
		})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, gin.H{"error": appErr})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("failed to get user", err),
		})
		return
	}

	response := &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		Bio:       user.Bio,
		Location:  user.Location,
		Website:   user.Website,
		Company:   user.Company,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("username is required"),
		})
		return
	}

	user, err := h.userRepo.GetByUsername(c.Request.Context(), username)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, gin.H{"error": appErr})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("failed to get user", err),
		})
		return
	}

	response := &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		Bio:       user.Bio,
		Location:  user.Location,
		Website:   user.Website,
		Company:   user.Company,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("user ID is required"),
		})
		return
	}

	var req dto.UpdateUserRequest
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
		"message": "User updated successfully",
	})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("user ID is required"),
		})
		return
	}

	err := h.userRepo.Delete(c.Request.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, gin.H{"error": appErr})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("failed to delete user", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

func (h *UserHandler) ListUsers(c *gin.Context) {
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

	users, err := h.userRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("failed to list users", err),
		})
		return
	}

	response := make([]dto.UserResponse, len(users))
	for i, user := range users {
		response[i] = dto.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Name:      user.Name,
			Email:     user.Email,
			AvatarURL: user.AvatarURL,
			Bio:       user.Bio,
			Location:  user.Location,
			Website:   user.Website,
			Company:   user.Company,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  response,
		"limit":  limit,
		"offset": offset,
		"total":  len(response),
	})
}

func (h *UserHandler) AddFavorite(c *gin.Context) {
	userID := c.Param("id")
	templateID := c.Param("templateId")

	if userID == "" || templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("user ID and template ID are required"),
		})
		return
	}

	err := h.userRepo.AddFavorite(c.Request.Context(), userID, templateID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, gin.H{"error": appErr})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("failed to add favorite", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Template added to favorites",
	})
}

func (h *UserHandler) RemoveFavorite(c *gin.Context) {
	userID := c.Param("id")
	templateID := c.Param("templateId")

	if userID == "" || templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("user ID and template ID are required"),
		})
		return
	}

	err := h.userRepo.RemoveFavorite(c.Request.Context(), userID, templateID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, gin.H{"error": appErr})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("failed to remove favorite", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Template removed from favorites",
	})
}

func (h *UserHandler) GetFavorites(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.NewBadRequestError("user ID is required"),
		})
		return
	}

	favorites, err := h.userRepo.GetFavorites(c.Request.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, gin.H{"error": appErr})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.NewInternalError("failed to get favorites", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"favorites": favorites,
	})
}