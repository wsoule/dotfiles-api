package models

import "time"

// User represents a system user
type User struct {
	ID         string    `json:"id" bson:"_id"`
	Username   string    `json:"username" bson:"username"`
	Name       string    `json:"name" bson:"name"`
	Email      string    `json:"email" bson:"email"`
	AvatarURL  string    `json:"avatar_url" bson:"avatar_url"`
	Bio        string    `json:"bio" bson:"bio"`
	Location   string    `json:"location" bson:"location"`
	Website    string    `json:"website" bson:"website"`
	Company    string    `json:"company" bson:"company"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
	Favorites  []string  `json:"favorites" bson:"favorites"`
}

// UserProfile represents a user's public profile information
type UserProfile struct {
	User               *User                    `json:"user"`
	Reviews            []*Review                `json:"reviews"`
	FavoriteTemplates  []*Template              `json:"favorite_templates"`
	Organizations      []*Organization          `json:"organizations"`
	Stats              *UserStats               `json:"stats"`
}

// UserStats contains user statistics
type UserStats struct {
	TemplateCount     int `json:"template_count"`
	ReviewCount       int `json:"review_count"`
	FavoriteCount     int `json:"favorite_count"`
	OrganizationCount int `json:"organization_count"`
}