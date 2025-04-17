package merhongo_test

import (
	"testing"

	"github.com/isimtekin/merhongo"
	"github.com/isimtekin/merhongo/schema"
)

func TestVersionCheck(t *testing.T) {
	// Simple test to verify version string is returned
	version := merhongo.Version()
	if version == "" {
		t.Error("Version string is empty")
	}
}

func TestConnectionSingleton(t *testing.T) {
	// Ensure clean state for tests
	_ = merhongo.DisconnectAll()

	// Connect using default name
	client, err := merhongo.Connect("mongodb://localhost:27017", "merhongo_test")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Get default connection
	defaultClient := merhongo.GetConnection()
	if defaultClient != client {
		t.Error("GetConnection() should return the same client instance")
	}

	// Test multiple named connections
	namedClient, err := merhongo.ConnectWithName("secondary", "mongodb://localhost:27017", "merhongo_test2")
	if err != nil {
		t.Fatalf("Failed to connect with name: %v", err)
	}

	// Verify named connection can be retrieved
	retrievedClient := merhongo.GetConnectionByName("secondary")
	if retrievedClient != namedClient {
		t.Error("GetConnectionByName() should return the correct client instance")
	}

	// Verify disconnection of named connection
	err = merhongo.DisconnectByName("secondary")
	if err != nil {
		t.Errorf("Failed to disconnect named connection: %v", err)
	}

	// Verify named connection is removed after disconnection
	if merhongo.GetConnectionByName("secondary") != nil {
		t.Error("Connection should be nil after disconnection")
	}

	// Verify default connection still exists
	if merhongo.GetConnection() == nil {
		t.Error("Default connection should still exist")
	}

	// Disconnect default connection
	err = merhongo.Disconnect()
	if err != nil {
		t.Errorf("Failed to disconnect default connection: %v", err)
	}

	// Verify all connections are removed
	if merhongo.GetConnection() != nil {
		t.Error("Default connection should be nil after disconnection")
	}
}

func TestConnectInvalidURI(t *testing.T) {
	// Test with invalid MongoDB URI
	_, err := merhongo.Connect("mongodb://invalidhost:27017", "merhongo_test")
	if err == nil {
		t.Error("Connect should return an error with invalid URI")
	}
}

func TestConnectWithEmptyName(t *testing.T) {
	// Test with empty connection name
	_, err := merhongo.ConnectWithName("", "mongodb://localhost:27017", "merhongo_test")
	if err == nil {
		t.Error("ConnectWithName should return an error with empty name")
	}
}

func TestConvenienceFunctions(t *testing.T) {
	// Test SchemaNew
	fields := map[string]schema.Field{
		"Email": {Required: true, Unique: true},
	}
	s := merhongo.SchemaNew(fields, schema.WithCollection("users"))

	if s.Fields["Email"].Required != true {
		t.Errorf("Expected Email field to be required")
	}

	if s.Collection != "users" {
		t.Errorf("Expected collection name to be 'users'")
	}

	// Test QueryNew
	q := merhongo.QueryNew()
	if q == nil {
		t.Error("QueryNew() should return a new query builder")
	}
}

func TestDisconnectNonExistentConnection(t *testing.T) {
	// Ensure clean state
	_ = merhongo.DisconnectAll()

	// Disconnecting a non-existent connection should not error
	err := merhongo.DisconnectByName("non_existent")
	if err != nil {
		t.Errorf("DisconnectByName should not return an error for non-existent connection: %v", err)
	}
}

func TestMultipleConnections(t *testing.T) {
	// Ensure clean state
	_ = merhongo.DisconnectAll()

	// Create multiple connections
	_, err := merhongo.Connect("mongodb://localhost:27017", "db1")
	if err != nil {
		t.Fatalf("Failed to connect to db1: %v", err)
	}

	_, err = merhongo.ConnectWithName("conn2", "mongodb://localhost:27017", "db2")
	if err != nil {
		t.Fatalf("Failed to connect to db2: %v", err)
	}

	_, err = merhongo.ConnectWithName("conn3", "mongodb://localhost:27017", "db3")
	if err != nil {
		t.Fatalf("Failed to connect to db3: %v", err)
	}

	// Test DisconnectAll
	err = merhongo.DisconnectAll()
	if err != nil {
		t.Errorf("DisconnectAll failed: %v", err)
	}

	// Verify all connections are removed
	if merhongo.GetConnection() != nil {
		t.Error("Default connection should be nil after DisconnectAll")
	}
	if merhongo.GetConnectionByName("conn2") != nil {
		t.Error("conn2 should be nil after DisconnectAll")
	}
	if merhongo.GetConnectionByName("conn3") != nil {
		t.Error("conn3 should be nil after DisconnectAll")
	}
}

// Additional test cases for error scenarios

func TestDisconnectWithInvalidConnection(t *testing.T) {
	// Ensure clean state
	_ = merhongo.DisconnectAll()

	// Create a client with an intentionally broken connection
	_, err := merhongo.Connect("mongodb://localhost:99999", "merhongo_test")
	if err == nil {
		t.Fatal("Expected connection error with invalid port")
	}
}

func TestMultipleConnectionsWithSameName(t *testing.T) {
	// Ensure clean state
	_ = merhongo.DisconnectAll()

	// First connection
	_, err := merhongo.ConnectWithName("duplicate", "mongodb://localhost:27017", "db1")
	if err != nil {
		t.Fatalf("First connection failed: %v", err)
	}

	// Second connection with same name should override
	_, err = merhongo.ConnectWithName("duplicate", "mongodb://localhost:27017", "db2")
	if err != nil {
		t.Fatalf("Second connection failed: %v", err)
	}

	// Verify the latest connection is retrieved
	client := merhongo.GetConnectionByName("duplicate")
	if client == nil {
		t.Error("Connection with duplicate name should exist")
	}
}

func TestDisconnectAllWithError(t *testing.T) {
	// This is a bit tricky to test precisely, but we can at least
	// ensure DisconnectAll handles multiple connections
	_ = merhongo.DisconnectAll()

	// Create multiple connections
	_, err1 := merhongo.Connect("mongodb://localhost:27017", "db1")
	_, err2 := merhongo.ConnectWithName("conn2", "mongodb://localhost:27017", "db2")

	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to create connections: %v, %v", err1, err2)
	}

	// Disconnect all
	err := merhongo.DisconnectAll()
	if err != nil {
		t.Errorf("DisconnectAll should not return an error: %v", err)
	}
}

func TestModelNew(t *testing.T) {
	// Create a connection first
	client, err := merhongo.Connect("mongodb://localhost:27017", "merhongo_test")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer merhongo.Disconnect()

	// Define a schema
	fields := map[string]schema.Field{
		"Username": {
			Required: true,
			Unique:   true,
		},
		"Email": {
			Required: true,
		},
	}

	// Create a schema
	s := merhongo.SchemaNew(fields, schema.WithCollection("test_users"))

	// Create a model using ModelNew
	model := merhongo.ModelNew("TestUser", s, client.Database)

	// Verify the model is created correctly
	if model == nil {
		t.Error("ModelNew should return a non-nil model")
	}

	// Check if the model's name is set correctly
	if model.Name != "TestUser" {
		t.Errorf("Expected model name 'TestUser', got '%s'", model.Name)
	}

	// Check if the schema is set correctly
	if model.Schema != s {
		t.Error("ModelNew should set the provided schema")
	}

	// Verify the collection is set
	if model.Collection == nil {
		t.Error("Model should have a non-nil collection")
	}

	// Check the collection name
	if model.Collection.Name() != "test_users" {
		t.Errorf("Expected collection name 'test_users', got '%s'", model.Collection.Name())
	}
}
