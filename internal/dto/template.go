package dto

import (
	"strings"

	"dotfiles-web/pkg/errors"
)

type CreateTemplateRequest struct {
	Taps           []string                  `json:"taps"`
	Brews          []string                  `json:"brews"`
	Casks          []string                  `json:"casks"`
	Stow           []string                  `json:"stow"`
	Metadata       CreateTemplateMetadata    `json:"metadata" binding:"required"`
	Extends        string                    `json:"extends"`
	Overrides      []string                  `json:"overrides"`
	AddOnly        bool                      `json:"add_only"`
	Public         bool                      `json:"public"`
	Featured       bool                      `json:"featured"`
	OrganizationID string                    `json:"organization_id"`
}

type CreateTemplateMetadata struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description" binding:"required"`
	Author      string   `json:"author" binding:"required"`
	Version     string   `json:"version" binding:"required"`
	Tags        []string `json:"tags"`
}

func (r *CreateTemplateRequest) Validate() *errors.AppError {
	if err := validateTemplateName(r.Metadata.Name); err != nil {
		return err
	}

	if err := validateTemplateDescription(r.Metadata.Description); err != nil {
		return err
	}

	if err := validateTemplateVersion(r.Metadata.Version); err != nil {
		return err
	}

	if err := validateTemplateTags(r.Metadata.Tags); err != nil {
		return err
	}

	return nil
}

type UpdateTemplateRequest struct {
	Taps        *[]string                 `json:"taps"`
	Brews       *[]string                 `json:"brews"`
	Casks       *[]string                 `json:"casks"`
	Stow        *[]string                 `json:"stow"`
	Metadata    *UpdateTemplateMetadata   `json:"metadata"`
	Extends     *string                   `json:"extends"`
	Overrides   *[]string                 `json:"overrides"`
	AddOnly     *bool                     `json:"add_only"`
	Public      *bool                     `json:"public"`
	Featured    *bool                     `json:"featured"`
}

type UpdateTemplateMetadata struct {
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	Version     *string   `json:"version"`
	Tags        *[]string `json:"tags"`
}

func (r *UpdateTemplateRequest) Validate() *errors.AppError {
	if r.Metadata != nil {
		if r.Metadata.Name != nil {
			if err := validateTemplateName(*r.Metadata.Name); err != nil {
				return err
			}
		}

		if r.Metadata.Description != nil {
			if err := validateTemplateDescription(*r.Metadata.Description); err != nil {
				return err
			}
		}

		if r.Metadata.Version != nil {
			if err := validateTemplateVersion(*r.Metadata.Version); err != nil {
				return err
			}
		}

		if r.Metadata.Tags != nil {
			if err := validateTemplateTags(*r.Metadata.Tags); err != nil {
				return err
			}
		}
	}

	return nil
}

type TemplateResponse struct {
	ID             string                    `json:"id"`
	Taps           []string                  `json:"taps"`
	Brews          []string                  `json:"brews"`
	Casks          []string                  `json:"casks"`
	Stow           []string                  `json:"stow"`
	Metadata       TemplateMetadataResponse  `json:"metadata"`
	Extends        string                    `json:"extends"`
	Overrides      []string                  `json:"overrides"`
	AddOnly        bool                      `json:"add_only"`
	Public         bool                      `json:"public"`
	Featured       bool                      `json:"featured"`
	OrganizationID string                    `json:"organization_id"`
	Downloads      int                       `json:"downloads"`
	CreatedAt      string                    `json:"created_at"`
	UpdatedAt      string                    `json:"updated_at"`
}

type TemplateMetadataResponse struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Version     string   `json:"version"`
	Tags        []string `json:"tags"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type TemplateStatsResponse struct {
	TotalTemplates    int `json:"total_templates"`
	FeaturedTemplates int `json:"featured_templates"`
	TotalDownloads    int `json:"total_downloads"`
	Categories        int `json:"categories"`
}

type TemplateRatingResponse struct {
	TemplateID    string         `json:"template_id"`
	AverageRating float64        `json:"average_rating"`
	TotalRatings  int            `json:"total_ratings"`
	Distribution  map[string]int `json:"distribution"`
}

func validateTemplateName(name string) *errors.AppError {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.NewValidationError("template name is required")
	}

	if len(name) < 3 || len(name) > 100 {
		return errors.NewValidationError("template name must be between 3 and 100 characters")
	}

	return nil
}

func validateTemplateDescription(description string) *errors.AppError {
	description = strings.TrimSpace(description)
	if description == "" {
		return errors.NewValidationError("template description is required")
	}

	if len(description) < 10 || len(description) > 500 {
		return errors.NewValidationError("template description must be between 10 and 500 characters")
	}

	return nil
}

func validateTemplateVersion(version string) *errors.AppError {
	version = strings.TrimSpace(version)
	if version == "" {
		return errors.NewValidationError("template version is required")
	}

	return nil
}

func validateTemplateTags(tags []string) *errors.AppError {
	if len(tags) > 10 {
		return errors.NewValidationError("template cannot have more than 10 tags")
	}

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return errors.NewValidationError("empty tags are not allowed")
		}

		if len(tag) > 30 {
			return errors.NewValidationError("tag cannot be longer than 30 characters")
		}
	}

	return nil
}