package database

import (
	"context"
	"fmt"
	"log"

	"ecommerce-backend/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client wraps the MongoDB client
type Client struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// NewClient creates a new MongoDB client
func NewClient(cfg *config.DatabaseConfig) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.URI)
	clientOptions.SetMaxPoolSize(cfg.MaxPoolSize)
	clientOptions.SetServerSelectionTimeout(cfg.Timeout)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(cfg.Database)

	log.Println("Connected to MongoDB successfully")
	return &Client{
		Client:   client,
		Database: db,
	}, nil
}

// Close closes the MongoDB connection
func (c *Client) Close(ctx context.Context) error {
	if err := c.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}
	return nil
}

// GetCollection returns a MongoDB collection
func (c *Client) GetCollection(name string) *mongo.Collection {
	return c.Database.Collection(name)
}
