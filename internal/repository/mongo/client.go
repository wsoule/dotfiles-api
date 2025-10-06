package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client wraps MongoDB client and database
type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewClient creates a new MongoDB client
func NewClient(mongoURI, dbName string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	// Test connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	database := client.Database(dbName)
	return &Client{
		client:   client,
		database: database,
	}, nil
}

// Close closes the MongoDB connection
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// Collection returns a MongoDB collection
func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// Database returns the MongoDB database
func (c *Client) Database() *mongo.Database {
	return c.database
}