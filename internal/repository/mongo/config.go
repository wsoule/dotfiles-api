package mongo

import (
	"context"

	"dotfiles-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConfigRepository implements the ConfigRepository interface using MongoDB
type ConfigRepository struct {
	collection *mongo.Collection
}

// NewConfigRepository creates a new config repository
func NewConfigRepository(client *Client) *ConfigRepository {
	return &ConfigRepository{
		collection: client.Collection("configs"),
	}
}

// Create stores a new config
func (r *ConfigRepository) Create(ctx context.Context, config *models.StoredConfig) error {
	_, err := r.collection.InsertOne(ctx, config)
	return err
}

// GetByID retrieves a config by ID
func (r *ConfigRepository) GetByID(ctx context.Context, id string) (*models.StoredConfig, error) {
	var config models.StoredConfig
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Update updates an existing config
func (r *ConfigRepository) Update(ctx context.Context, config *models.StoredConfig) error {
	// Note: UpdatedAt field doesn't exist in current StoredConfig model
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": config.ID}, config)
	return err
}

// Delete removes a config
func (r *ConfigRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// List retrieves configs with pagination
func (r *ConfigRepository) List(ctx context.Context, limit, offset int) ([]*models.StoredConfig, error) {
	cursor, err := r.collection.Find(ctx, bson.M{},
		&options.FindOptions{
			Limit: int64ptr(limit),
			Skip:  int64ptr(offset),
		},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var configs []*models.StoredConfig
	if err = cursor.All(ctx, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}

// GetStats returns config statistics
func (r *ConfigRepository) GetStats(ctx context.Context) (*models.ConfigStats, error) {
	total, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	public, err := r.collection.CountDocuments(ctx, bson.M{"public": true})
	if err != nil {
		return nil, err
	}

	// Calculate total downloads
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$download_count"},
		}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result struct {
		Total int `bson:"total"`
	}
	totalDownloads := 0
	if cursor.Next(ctx) {
		cursor.Decode(&result)
		totalDownloads = result.Total
	}

	return &models.ConfigStats{
		TotalConfigs:   int(total),
		PublicConfigs:  int(public),
		TotalDownloads: totalDownloads,
	}, nil
}

// IncrementDownloads increments the download count for a config
func (r *ConfigRepository) IncrementDownloads(ctx context.Context, id string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$inc": bson.M{"download_count": 1}},
	)
	return err
}

func int64ptr(i int) *int64 {
	val := int64(i)
	return &val
}