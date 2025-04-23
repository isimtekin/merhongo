package merhongo_test_test

import (
	"context"
	"testing"
	"time"

	"github.com/isimtekin/merhongo"
	"github.com/isimtekin/merhongo/schema"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestUser is a test model for users
type TestUser struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Username  string             `bson:"username"`
	Email     string             `bson:"email"`
	CreatedAt time.Time          `bson:"createdAt,omitempty"`
	UpdatedAt time.Time          `bson:"updatedAt,omitempty"`
}

// TestPost is a test model for posts
type TestPost struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Title     string             `bson:"title"`
	Content   string             `bson:"content"`
	CreatedAt time.Time          `bson:"createdAt,omitempty"`
	UpdatedAt time.Time          `bson:"updatedAt,omitempty"`
}

func TestModelNew_WithGenerics(t *testing.T) {
	// Connect to MongoDB
	_, err := merhongo.Connect("mongodb://localhost:27017", "merhongo_test_generics")
	if err != nil {
		t.Skip("Skipping test; could not connect to MongoDB")
		return
	}
	defer merhongo.Disconnect()

	// Define a schema
	userSchema := merhongo.SchemaNew(
		map[string]schema.Field{
			"Username": {Required: true, Unique: true},
			"Email":    {Required: true},
		},
		schema.WithCollection("test_users_generics"),
	)

	// Create model with generics
	userModel := merhongo.ModelNew[TestUser]("TestUser", userSchema)
	assert.NotNil(t, userModel, "Model should not be nil")
	assert.Equal(t, "TestUser", userModel.Name, "Model name should match")
	assert.NotNil(t, userModel.Collection, "Collection should be initialized")

	// Try to create a document
	user := &TestUser{
		Username: "test_generics",
		Email:    "test_generics@example.com",
	}

	err = userModel.Create(context.Background(), user)
	assert.NoError(t, err, "Should be able to create a document")
	assert.NotEqual(t, primitive.NilObjectID, user.ID, "ID should be set")

	// Clean up
	_, _ = userModel.Collection.DeleteMany(context.Background(), map[string]interface{}{})
}

func TestModelNew_WithOptions(t *testing.T) {
	// Connect to MongoDB
	_, err := merhongo.Connect("mongodb://localhost:27017", "merhongo_test_options")
	if err != nil {
		t.Skip("Skipping test; could not connect to MongoDB")
		return
	}
	defer merhongo.Disconnect()

	// Define a schema
	postSchema := merhongo.SchemaNew(
		map[string]schema.Field{
			"Title":   {Required: true},
			"Content": {Required: true},
		},
		schema.WithCollection("test_posts_options"),
	)

	// Test with custom validator option
	customValidatorCalled := false
	postModel := merhongo.ModelNew[TestPost]("TestPost", postSchema, merhongo.ModelOptions{
		AutoCreateIndexes: true,
		CustomValidator: func(doc interface{}) error {
			customValidatorCalled = true
			return nil
		},
	})

	assert.NotNil(t, postModel, "Model should not be nil")

	// Create a post to trigger validation
	post := &TestPost{
		Title:   "Test Options",
		Content: "Testing model options",
	}

	err = postModel.Create(context.Background(), post)
	assert.NoError(t, err, "Should be able to create a document")
	assert.True(t, customValidatorCalled, "Custom validator should be called")

	// Clean up
	_, _ = postModel.Collection.DeleteMany(context.Background(), map[string]interface{}{})
}

func TestModelNew_WithConnectionName(t *testing.T) {
	// Create a named connection
	_, err := merhongo.ConnectWithName("test_conn", "mongodb://localhost:27017", "merhongo_test_named")
	if err != nil {
		t.Skip("Skipping test; could not connect to MongoDB")
		return
	}
	defer merhongo.DisconnectByName("test_conn")

	// Define a schema
	userSchema := merhongo.SchemaNew(
		map[string]schema.Field{
			"Username": {Required: true},
			"Email":    {Required: true},
		},
		schema.WithCollection("test_users_named"),
	)

	// Create model with connection name option
	userModel := merhongo.ModelNew[TestUser]("TestUserNamed", userSchema, merhongo.ModelOptions{
		ConnectionName: "test_conn",
	})

	assert.NotNil(t, userModel, "Model should not be nil")
	assert.NotNil(t, userModel.Collection, "Collection should be initialized")

	// Verify it's using the correct connection by creating a document
	user := &TestUser{
		Username: "test_named_conn",
		Email:    "test_named@example.com",
	}

	err = userModel.Create(context.Background(), user)
	assert.NoError(t, err, "Should be able to create a document")

	// Clean up
	_, _ = userModel.Collection.DeleteMany(context.Background(), map[string]interface{}{})
}

func TestDisconnectAll(t *testing.T) {
	// Create multiple connections
	conn1, err := merhongo.ConnectWithName("conn1", "mongodb://localhost:27017", "merhongo_test_disconnect_all1")
	assert.NotNil(t, conn1, "Connection object should not be nil")
	if err != nil {
		t.Skip("Skipping test; could not connect to MongoDB")
		return
	}

	conn2, err := merhongo.ConnectWithName("conn2", "mongodb://localhost:27017", "merhongo_test_disconnect_all2")
	assert.NotNil(t, conn2, "Connection object should not be nil")
	if err != nil {
		// Clean up first connection
		_ = merhongo.DisconnectByName("conn1")
		t.Skip("Skipping test; could not connect to MongoDB")
		return
	}

	conn3, err := merhongo.Connect("mongodb://localhost:27017", "merhongo_test_disconnect_all3")
	assert.NotNil(t, conn3, "Connection object should not be nil")
	if err != nil {
		// Clean up created connections
		_ = merhongo.DisconnectByName("conn1")
		_ = merhongo.DisconnectByName("conn2")
		t.Skip("Skipping test; could not connect to MongoDB")
		return
	}

	// Ensure all connections exist
	assert.NotNil(t, merhongo.GetConnectionByName("conn1"), "conn1 should exist")
	assert.NotNil(t, merhongo.GetConnectionByName("conn2"), "conn2 should exist")
	assert.NotNil(t, merhongo.GetConnection(), "default connection should exist")

	// Disconnect all connections
	err = merhongo.DisconnectAll()
	assert.NoError(t, err, "DisconnectAll should succeed")

	// Verify all connections were removed
	assert.Nil(t, merhongo.GetConnectionByName("conn1"), "conn1 should be removed after DisconnectAll")
	assert.Nil(t, merhongo.GetConnectionByName("conn2"), "conn2 should be removed after DisconnectAll")
	assert.Nil(t, merhongo.GetConnection(), "default connection should be removed after DisconnectAll")
}

func TestDisconnect(t *testing.T) {
	// Create a default connection
	conn, err := merhongo.Connect("mongodb://localhost:27017", "merhongo_test_disconnect")
	assert.NotNil(t, conn, "Connection object should not be nil")

	if err != nil {
		t.Skip("Skipping test; could not connect to MongoDB")
		return
	}

	// Ensure the connection exists
	assert.NotNil(t, merhongo.GetConnection(), "Default connection should exist")

	// Disconnect
	err = merhongo.Disconnect()
	assert.NoError(t, err, "Disconnect should succeed")

	// Verify the connection was removed
	assert.Nil(t, merhongo.GetConnection(), "Default connection should be removed after disconnect")
}

func TestDisconnectByName(t *testing.T) {
	// Create named connections
	conn1, err := merhongo.ConnectWithName("test_disconnect_1", "mongodb://localhost:27017", "merhongo_test_disconnect1")
	assert.NotNil(t, conn1, "Connection object should not be nil")

	if err != nil {
		t.Skip("Skipping test; could not connect to MongoDB")
		return
	}

	conn2, err := merhongo.ConnectWithName("test_disconnect_2", "mongodb://localhost:27017", "merhongo_test_disconnect2")
	assert.NotNil(t, conn2, "Connection object should not be nil")

	if err != nil {
		_ = merhongo.DisconnectByName("test_disconnect_1")
		t.Skip("Skipping test; could not connect to MongoDB")
		return
	}

	// Ensure both connections exist
	assert.NotNil(t, merhongo.GetConnectionByName("test_disconnect_1"), "test_disconnect_1 should exist")
	assert.NotNil(t, merhongo.GetConnectionByName("test_disconnect_2"), "test_disconnect_2 should exist")

	// Disconnect one connection
	err = merhongo.DisconnectByName("test_disconnect_1")
	assert.NoError(t, err, "DisconnectByName should succeed")

	// Verify one connection was removed but the other remains
	assert.Nil(t, merhongo.GetConnectionByName("test_disconnect_1"), "test_disconnect_1 should be removed after disconnect")
	assert.NotNil(t, merhongo.GetConnectionByName("test_disconnect_2"), "test_disconnect_2 should still exist")

	// Clean up the second connection
	_ = merhongo.DisconnectByName("test_disconnect_2")
}

func TestDisconnectByName_NonExistent(t *testing.T) {
	// Try to disconnect a non-existent connection
	err := merhongo.DisconnectByName("non_existent_connection")

	// Should not return an error
	assert.NoError(t, err, "DisconnectByName on non-existent connection should not return error")
}
