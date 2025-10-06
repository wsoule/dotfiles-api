package models

import "time"

// Config represents a dotfiles configuration
type Config struct {
	ID          string    `json:"id" bson:"_id"`
	Description string    `json:"description" bson:"description"`
	Brews       []string  `json:"brews" bson:"brews"`
	Casks       []string  `json:"casks" bson:"casks"`
	Taps        []string  `json:"taps" bson:"taps"`
	Stow        []string  `json:"stow" bson:"stow"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	Downloads   int       `json:"downloads" bson:"downloads"`
	Public      bool      `json:"public" bson:"public"`
	Featured    bool      `json:"featured" bson:"featured"`
}

// StoredConfig represents a config stored in the database
type StoredConfig struct {
	ID            string          `json:"id" bson:"_id"`
	Config        ShareableConfig `json:"config" bson:"config"`
	Public        bool            `json:"public" bson:"public"`
	CreatedAt     time.Time       `json:"created_at" bson:"created_at"`
	DownloadCount int             `json:"download_count" bson:"download_count"`
	OwnerID       string          `json:"owner_id" bson:"owner_id"`
}

// ConfigStats contains configuration statistics
type ConfigStats struct {
	TotalConfigs   int `json:"total_configs"`
	PublicConfigs  int `json:"public_configs"`
	TotalDownloads int `json:"total_downloads"`
}