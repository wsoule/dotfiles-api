package mongo

import (
	"context"
	"time"

	"dotfiles-web/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ReviewRepository implements the ReviewRepository interface using MongoDB
type ReviewRepository struct {
	collection *mongo.Collection
}

// NewReviewRepository creates a new review repository
func NewReviewRepository(client *Client) *ReviewRepository {
	return &ReviewRepository{
		collection: client.Collection("reviews"),
	}
}

// Create stores a new review
func (r *ReviewRepository) Create(ctx context.Context, review *models.Review) error {
	if review.ID == "" {
		review.ID = primitive.NewObjectID().Hex()
	}
	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, review)
	return err
}

// GetByID retrieves a review by ID
func (r *ReviewRepository) GetByID(ctx context.Context, id string) (*models.Review, error) {
	var review models.Review
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&review)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &review, nil
}

// Update updates an existing review
func (r *ReviewRepository) Update(ctx context.Context, review *models.Review) error {
	review.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": review.ID}, review)
	return err
}

// Delete removes a review
func (r *ReviewRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// GetByTemplate retrieves reviews for a template
func (r *ReviewRepository) GetByTemplate(ctx context.Context, templateID string, limit, offset int) ([]*models.Review, error) {
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "created_at", Value: -1}},
		Limit: int64ptr(limit),
		Skip:  int64ptr(offset),
	}

	cursor, err := r.collection.Find(ctx, bson.M{"template_id": templateID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reviews []*models.Review
	if err = cursor.All(ctx, &reviews); err != nil {
		return nil, err
	}
	return reviews, nil
}

// GetByUser retrieves reviews by a user
func (r *ReviewRepository) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Review, error) {
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "created_at", Value: -1}},
		Limit: int64ptr(limit),
		Skip:  int64ptr(offset),
	}

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reviews []*models.Review
	if err = cursor.All(ctx, &reviews); err != nil {
		return nil, err
	}
	return reviews, nil
}

// GetUserReviewForTemplate retrieves a user's review for a specific template
func (r *ReviewRepository) GetUserReviewForTemplate(ctx context.Context, userID, templateID string) (*models.Review, error) {
	var review models.Review
	err := r.collection.FindOne(ctx, bson.M{
		"user_id":     userID,
		"template_id": templateID,
	}).Decode(&review)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &review, nil
}

// IncrementHelpful increments the helpful count for a review
func (r *ReviewRepository) IncrementHelpful(ctx context.Context, id string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$inc": bson.M{"helpful": 1}},
	)
	return err
}

// CalculateTemplateRating calculates the rating information for a template
func (r *ReviewRepository) CalculateTemplateRating(ctx context.Context, templateID string) (*models.TemplateRating, error) {
	// Aggregate pipeline to calculate rating statistics
	pipeline := []bson.M{
		{"$match": bson.M{"template_id": templateID}},
		{"$group": bson.M{
			"_id": nil,
			"avg_rating": bson.M{"$avg": "$rating"},
			"total_ratings": bson.M{"$sum": 1},
			"ratings": bson.M{"$push": "$rating"},
		}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result struct {
		AvgRating    float64 `bson:"avg_rating"`
		TotalRatings int     `bson:"total_ratings"`
		Ratings      []int   `bson:"ratings"`
	}

	if !cursor.Next(ctx) {
		// No reviews found
		return &models.TemplateRating{
			TemplateID:     templateID,
			AverageRating:  0.0,
			TotalRatings:   0,
			Distribution:   make(map[string]int),
		}, nil
	}

	if err := cursor.Decode(&result); err != nil {
		return nil, err
	}

	// Calculate distribution
	distribution := make(map[string]int)
	for _, rating := range result.Ratings {
		key := string(rune('0' + rating)) // Convert to string
		distribution[key]++
	}

	return &models.TemplateRating{
		TemplateID:     templateID,
		AverageRating:  result.AvgRating,
		TotalRatings:   result.TotalRatings,
		Distribution:   distribution,
	}, nil
}