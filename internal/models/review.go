package models

import "time"

// Review represents a user review of a template
type Review struct {
	ID         string    `json:"id" bson:"_id"`
	TemplateID string    `json:"template_id" bson:"template_id"`
	UserID     string    `json:"user_id" bson:"user_id"`
	Username   string    `json:"username" bson:"username"`
	AvatarURL  string    `json:"avatar_url" bson:"avatar_url"`
	Rating     int       `json:"rating" bson:"rating"` // 1-5 stars
	Comment    string    `json:"comment" bson:"comment"`
	Helpful    int       `json:"helpful" bson:"helpful"` // helpful votes count
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
}

// IsValidRating checks if the rating is within valid range (1-5)
func (r Review) IsValidRating() bool {
	return r.Rating >= 1 && r.Rating <= 5
}