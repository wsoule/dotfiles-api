package models

import "time"

// PackageConfig represents configuration for a specific package
type PackageConfig struct {
	PostInstall []string `json:"post_install,omitempty" bson:"post_install,omitempty"`
	PreInstall  []string `json:"pre_install,omitempty" bson:"pre_install,omitempty"`
}

// Template represents a dotfiles template
type Template struct {
	Taps           []string                 `json:"taps" bson:"taps"`
	Brews          []string                 `json:"brews" bson:"brews"`
	Casks          []string                 `json:"casks" bson:"casks"`
	Stow           []string                 `json:"stow" bson:"stow"`
	Metadata       ShareMetadata            `json:"metadata" bson:"metadata"`
	Extends        string                   `json:"extends,omitempty" bson:"extends"`
	Overrides      []string                 `json:"overrides,omitempty" bson:"overrides"`
	AddOnly        bool                     `json:"addOnly" bson:"add_only"`
	Public         bool                     `json:"public" bson:"public"`
	Featured       bool                     `json:"featured" bson:"featured"`
	OrganizationID string                   `json:"organization_id,omitempty" bson:"organization_id,omitempty"`
	PackageConfigs map[string]PackageConfig `json:"package_configs,omitempty" bson:"package_configs,omitempty"`
}

// TemplateMetadata contains template metadata
type TemplateMetadata struct {
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
	Author      string    `json:"author" bson:"author"`
	Version     string    `json:"version" bson:"version"`
	Tags        []string  `json:"tags" bson:"tags"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}

// StoredTemplate represents a template stored in the database
type StoredTemplate struct {
	ID        string    `json:"id" bson:"_id"`
	Template  Template  `json:"template" bson:"template"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	Downloads int       `json:"downloads" bson:"downloads"`
}

// TemplateStats contains template statistics
type TemplateStats struct {
	TotalTemplates    int `json:"total_templates"`
	FeaturedTemplates int `json:"featured_templates"`
	TotalDownloads    int `json:"total_downloads"`
	Categories        int `json:"categories"`
}

// TemplateRating represents template rating information
type TemplateRating struct {
	TemplateID     string             `json:"template_id"`
	AverageRating  float64            `json:"average_rating"`
	TotalRatings   int                `json:"total_ratings"`
	Distribution   map[string]int     `json:"distribution"` // rating -> count
}