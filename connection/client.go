// Package connection handles MongoDB client connection management
package connection

import (
	"context"
	"github.com/isimtekin/merhongo/errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client manages the database connection and models
type Client struct {
	MongoClient *mongo.Client
	Database    *mongo.Database
	Models      map[string]interface{}
}

// Connect creates a new MongoDB client instance and connects to the database
func Connect(uri, dbName string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create new client and connect
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, errors.WithDetails(errors.ErrConnection, "failed to connect")
	}

	// Verify connection with ping
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, errors.WithDetails(errors.ErrConnection, "failed to ping MongoDB")
	}

	log.Println("✅ Connected to MongoDB")

	return &Client{
		MongoClient: client,
		Database:    client.Database(dbName),
		Models:      make(map[string]interface{}),
	}, nil
}

// Disconnect closes the MongoDB connection
func (c *Client) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.MongoClient.Disconnect(ctx); err != nil {
		return errors.WithDetails(errors.ErrConnection, "failed to disconnect")
	}

	log.Println("✅ Disconnected from MongoDB")
	return nil
}

// ExecuteTransaction runs operations in a transaction
func (c *Client) ExecuteTransaction(ctx context.Context, fn func(mongo.SessionContext) error) error {
	return c.MongoClient.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			return errors.Wrap(errors.ErrDatabase, "failed to start transaction")
		}

		if err = fn(sessionContext); err != nil {
			abortErr := sessionContext.AbortTransaction(sessionContext)
			if abortErr != nil {
				log.Printf("Warning: Failed to abort transaction: %v", abortErr)
			}
			return err
		}

		commitErr := sessionContext.CommitTransaction(sessionContext)
		if commitErr != nil {
			return errors.Wrap(errors.ErrDatabase, "failed to commit transaction")
		}

		return nil
	})
}
