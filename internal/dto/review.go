package dto

import (
	"strings"

	"dotfiles-web/pkg/errors"
)

type CreateReviewRequest struct {
	TemplateID string `json:"template_id" binding:"required"`
	Rating     int    `json:"rating" binding:"required"`
	Comment    string `json:"comment"`
}

func (r *CreateReviewRequest) Validate() *errors.AppError {
	if err := validateRating(r.Rating); err != nil {
		return err
	}

	if r.Comment != "" {
		if err := validateReviewComment(r.Comment); err != nil {
			return err
		}
	}

	return nil
}

type UpdateReviewRequest struct {
	Rating  *int    `json:"rating"`
	Comment *string `json:"comment"`
}

func (r *UpdateReviewRequest) Validate() *errors.AppError {
	if r.Rating != nil {
		if err := validateRating(*r.Rating); err != nil {
			return err
		}
	}

	if r.Comment != nil && *r.Comment != "" {
		if err := validateReviewComment(*r.Comment); err != nil {
			return err
		}
	}

	return nil
}

type ReviewResponse struct {
	ID         string `json:"id"`
	TemplateID string `json:"template_id"`
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
	AvatarURL  string `json:"avatar_url"`
	Rating     int    `json:"rating"`
	Comment    string `json:"comment"`
	Helpful    int    `json:"helpful"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func validateRating(rating int) *errors.AppError {
	if rating < 1 || rating > 5 {
		return errors.NewValidationError("rating must be between 1 and 5")
	}

	return nil
}

func validateReviewComment(comment string) *errors.AppError {
	comment = strings.TrimSpace(comment)
	if len(comment) > 1000 {
		return errors.NewValidationError("review comment cannot be longer than 1000 characters")
	}

	return nil
}