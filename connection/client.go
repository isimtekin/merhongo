// Package connection handles MongoDB client connection management
package connection

import (
	"context"
	"log"
	"time"

	"github.com/isimtekin/merhongo/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client manages the database connection and models
type Client struct {
	// MongoClient is the underlying MongoDB client
	MongoClient *mongo.Client
	// Database is the default database for this connection
	Database *mongo.Database
	// Models stores model instances associated with this connection
	Models map[string]interface{}
	// Name of this connection instance
	Name string
}

// Connect creates a new MongoDB client instance and connects to the database
func Connect(uri, dbName string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create new client and connect
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Printf("⚠️ Failed to connect to MongoDB at %s: %v", uri, err)
		return nil, errors.WithDetails(errors.ErrConnection, "failed to connect")
	}

	// Verify connection with ping
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Printf("⚠️ Failed to ping MongoDB at %s: %v", uri, err)
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
	if c.MongoClient == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.MongoClient.Disconnect(ctx); err != nil {
		log.Printf("⚠️ Failed to disconnect from MongoDB: %v", err)
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
			log.Printf("⚠️ Failed to start transaction: %v", err)
			return errors.Wrap(errors.ErrDatabase, "failed to start transaction")
		}

		if err = fn(sessionContext); err != nil {
			abortErr := sessionContext.AbortTransaction(sessionContext)
			if abortErr != nil {
				log.Printf("⚠️ Failed to abort transaction: %v", abortErr)
			} else {
				log.Printf("✅ Transaction aborted due to error: %v", err)
			}
			return err
		}

		commitErr := sessionContext.CommitTransaction(sessionContext)
		if commitErr != nil {
			log.Printf("⚠️ Failed to commit transaction: %v", commitErr)
			return errors.Wrap(errors.ErrDatabase, "failed to commit transaction")
		}

		log.Println("✅ Transaction committed successfully")
		return nil
	})
}

// GetDatabase returns the database with the specified name
func (c *Client) GetDatabase(name string) *mongo.Database {
	if name == "" {
		return c.Database
	}
	return c.MongoClient.Database(name)
}

// RegisterModel registers a model with this connection
func (c *Client) RegisterModel(name string, model interface{}) {
	c.Models[name] = model
	log.Printf("✅ Registered model '%s'", name)
}

// GetModel retrieves a registered model by name
func (c *Client) GetModel(name string) interface{} {
	model := c.Models[name]
	if model == nil {
		log.Printf("⚠️ Model '%s' not found", name)
	}
	return model
}
