package dto

import (
	"regexp"
	"strings"

	"dotfiles-web/pkg/errors"
)

type CreateUserRequest struct {
	Username  string `json:"username" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
	Location  string `json:"location"`
	Website   string `json:"website"`
	Company   string `json:"company"`
}

func (r *CreateUserRequest) Validate() *errors.AppError {
	if err := validateUsername(r.Username); err != nil {
		return err
	}

	if err := validateEmail(r.Email); err != nil {
		return err
	}

	if r.Website != "" {
		if err := validateURL(r.Website); err != nil {
			return err
		}
	}

	return nil
}

type UpdateUserRequest struct {
	Name     *string `json:"name"`
	Bio      *string `json:"bio"`
	Location *string `json:"location"`
	Website  *string `json:"website"`
	Company  *string `json:"company"`
}

func (r *UpdateUserRequest) Validate() *errors.AppError {
	if r.Website != nil && *r.Website != "" {
		if err := validateURL(*r.Website); err != nil {
			return err
		}
	}

	return nil
}

type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
	Location  string `json:"location"`
	Website   string `json:"website"`
	Company   string `json:"company"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UserProfileResponse struct {
	User              *UserResponse       `json:"user"`
	Reviews           []ReviewResponse    `json:"reviews"`
	FavoriteTemplates []TemplateResponse  `json:"favorite_templates"`
	Organizations     []OrganizationResponse `json:"organizations"`
	Stats             *UserStatsResponse  `json:"stats"`
}

type UserStatsResponse struct {
	TemplateCount     int `json:"template_count"`
	ReviewCount       int `json:"review_count"`
	FavoriteCount     int `json:"favorite_count"`
	OrganizationCount int `json:"organization_count"`
}

func validateUsername(username string) *errors.AppError {
	if len(username) < 3 || len(username) > 30 {
		return errors.NewValidationError("username must be between 3 and 30 characters")
	}

	matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", username)
	if !matched {
		return errors.NewValidationError("username can only contain letters, numbers, hyphens, and underscores")
	}

	return nil
}

func validateEmail(email string) *errors.AppError {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.NewValidationError("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.NewValidationError("invalid email format")
	}

	return nil
}

func validateURL(url string) *errors.AppError {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil
	}

	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(url) {
		return errors.NewValidationError("invalid URL format")
	}

	return nil
}