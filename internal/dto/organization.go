package dto

import (
	"regexp"
	"strings"

	"dotfiles-web/pkg/errors"
)

type CreateOrganizationRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
	Website     string `json:"website"`
	Public      bool   `json:"public"`
}

func (r *CreateOrganizationRequest) Validate() *errors.AppError {
	if err := validateOrganizationName(r.Name); err != nil {
		return err
	}

	if err := validateOrganizationSlug(r.Slug); err != nil {
		return err
	}

	if r.Description != "" {
		if err := validateOrganizationDescription(r.Description); err != nil {
			return err
		}
	}

	if r.Website != "" {
		if err := validateURL(r.Website); err != nil {
			return err
		}
	}

	return nil
}

type UpdateOrganizationRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Website     *string `json:"website"`
	Public      *bool   `json:"public"`
}

func (r *UpdateOrganizationRequest) Validate() *errors.AppError {
	if r.Name != nil {
		if err := validateOrganizationName(*r.Name); err != nil {
			return err
		}
	}

	if r.Description != nil && *r.Description != "" {
		if err := validateOrganizationDescription(*r.Description); err != nil {
			return err
		}
	}

	if r.Website != nil && *r.Website != "" {
		if err := validateURL(*r.Website); err != nil {
			return err
		}
	}

	return nil
}

type OrganizationResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Website     string `json:"website"`
	OwnerID     string `json:"owner_id"`
	Public      bool   `json:"public"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	MemberCount int    `json:"member_count"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

func (r *AddMemberRequest) Validate() *errors.AppError {
	if err := validateOrganizationRole(r.Role); err != nil {
		return err
	}

	return nil
}

type UpdateMemberRequest struct {
	Role string `json:"role" binding:"required"`
}

func (r *UpdateMemberRequest) Validate() *errors.AppError {
	if err := validateOrganizationRole(r.Role); err != nil {
		return err
	}

	return nil
}

type OrganizationMemberResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	AvatarURL      string `json:"avatar_url"`
	Role           string `json:"role"`
	JoinedAt       string `json:"joined_at"`
}

type InviteUserRequest struct {
	Email string `json:"email" binding:"required"`
	Role  string `json:"role" binding:"required"`
}

func (r *InviteUserRequest) Validate() *errors.AppError {
	if err := validateEmail(r.Email); err != nil {
		return err
	}

	if err := validateOrganizationRole(r.Role); err != nil {
		return err
	}

	return nil
}

type OrganizationInviteResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	Token          string `json:"token"`
	InvitedBy      string `json:"invited_by"`
	CreatedAt      string `json:"created_at"`
	ExpiresAt      string `json:"expires_at"`
	AcceptedAt     string `json:"accepted_at,omitempty"`
}

type AcceptInviteRequest struct {
	Token string `json:"token" binding:"required"`
}

func validateOrganizationName(name string) *errors.AppError {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.NewValidationError("organization name is required")
	}

	if len(name) < 3 || len(name) > 50 {
		return errors.NewValidationError("organization name must be between 3 and 50 characters")
	}

	return nil
}

func validateOrganizationSlug(slug string) *errors.AppError {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return errors.NewValidationError("organization slug is required")
	}

	if len(slug) < 3 || len(slug) > 30 {
		return errors.NewValidationError("organization slug must be between 3 and 30 characters")
	}

	matched, _ := regexp.MatchString("^[a-z0-9-]+$", slug)
	if !matched {
		return errors.NewValidationError("organization slug can only contain lowercase letters, numbers, and hyphens")
	}

	if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return errors.NewValidationError("organization slug cannot start or end with a hyphen")
	}

	return nil
}

func validateOrganizationDescription(description string) *errors.AppError {
	description = strings.TrimSpace(description)
	if len(description) > 200 {
		return errors.NewValidationError("organization description cannot be longer than 200 characters")
	}

	return nil
}

func validateOrganizationRole(role string) *errors.AppError {
	validRoles := []string{"owner", "admin", "member"}
	for _, validRole := range validRoles {
		if role == validRole {
			return nil
		}
	}

	return errors.NewValidationError("invalid role: must be one of owner, admin, member")
}