// Package testutil provides common utilities for testing Merhongo packages
package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/isimtekin/merhongo/connection"
	"github.com/isimtekin/merhongo/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TestUser represents a user for testing both regular and query operations
type TestUser struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Username  string             `bson:"username"`
	Email     string             `bson:"email"`
	Age       int                `bson:"age"`
	Active    bool               `bson:"active"`
	Role      string             `bson:"role"`
	CreatedAt time.Time          `bson:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"`
}

// CreateTestSchema creates a standard test schema for TestUser
func CreateTestSchema(collectionName string) *schema.Schema {
	return schema.New(
		map[string]schema.Field{
			// Use lowercase field names to match bson tags
			"username": {Required: true, Unique: true},
			"email":    {Required: true, Unique: true},
			"age":      {Min: 18},
			"active":   {Type: true},
			"role":     {Type: ""},
		},
		schema.WithCollection(collectionName),
	)
}

// CreateTestClient creates a MongoDB client for testing
func CreateTestClient(t *testing.T) (*connection.Client, func()) {
	client, err := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	cleanup := func() {
		if err := client.Disconnect(); err != nil {
			t.Logf("Failed to disconnect: %v", err)
		}
	}

	return client, cleanup
}

// DropCollection drops the specified collection
func DropCollection(t *testing.T, db *mongo.Database, collectionName string) {
	err := db.Collection(collectionName).Drop(context.Background())
	if err != nil {
		t.Logf("Failed to drop collection %s: %v", collectionName, err)
	}
}

// CreateTestUsers creates standard test user data
func CreateTestUsers() []TestUser {
	return []TestUser{
		{
			Username:  "john_doe",
			Email:     "john@example.com",
			Age:       30,
			Active:    true,
			Role:      "user",
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			Username:  "jane_smith",
			Email:     "jane@example.com",
			Age:       25,
			Active:    true,
			Role:      "admin",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-12 * time.Hour),
		},
		{
			Username:  "bob_jones",
			Email:     "bob@example.com",
			Age:       45,
			Active:    false,
			Role:      "user",
			CreatedAt: time.Now().Add(-12 * time.Hour),
			UpdatedAt: time.Now().Add(-6 * time.Hour),
		},
		{
			Username:  "alice_wonder",
			Email:     "alice@example.com",
			Age:       22,
			Active:    true,
			Role:      "user",
			CreatedAt: time.Now().Add(-6 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
	}
}
