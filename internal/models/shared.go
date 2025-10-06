package models

import "time"

// ShareMetadata contains metadata for shareable configs and templates
type ShareMetadata struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Version     string    `json:"version"`
}

// BasicConfig represents a simple dotfiles configuration
type BasicConfig struct {
	Brews []string `json:"brews"`
	Casks []string `json:"casks"`
	Taps  []string `json:"taps"`
	Stow  []string `json:"stow"`
}

// ShareableConfig represents a shareable configuration with metadata
type ShareableConfig struct {
	BasicConfig `json:",inline"`
	Metadata    ShareMetadata `json:"metadata"`
}