package testutil

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

// CreateCollection creates a new empty collection
func CreateCollection(t *testing.T, db *mongo.Database, collectionName string) {
	ctx := context.Background()
	// First drop if exists
	DropCollection(t, db, collectionName)

	// Then create explicitly
	err := db.CreateCollection(ctx, collectionName)
	if err != nil {
		t.Logf("Note: Failed to explicitly create collection %s: %v", collectionName, err)
		// This is often not an error as the collection will be created on first insert
	}
}

// InsertDocuments is a generic helper to insert test documents
func InsertDocuments(t *testing.T, coll *mongo.Collection, documents []interface{}) {
	ctx := context.Background()
	if len(documents) == 0 {
		return
	}

	_, err := coll.InsertMany(ctx, documents)
	if err != nil {
		t.Fatalf("Failed to insert test documents: %v", err)
	}
}

// PrepareCollection sets up a clean collection for testing
func PrepareCollection(t *testing.T, db *mongo.Database, collectionName string) *mongo.Collection {
	// Drop existing collection if any
	DropCollection(t, db, collectionName)

	// Return the collection reference (MongoDB will create it on first use)
	return db.Collection(collectionName)
}

// Test Assertion Helpers

// AssertErrorType checks if an error is of the expected type
func AssertErrorType(t *testing.T, err error, errorCheck func(error) bool, errorTypeName string) {
	if err == nil {
		t.Errorf("Expected %s error, got nil", errorTypeName)
		return
	}

	if !errorCheck(err) {
		t.Errorf("Expected %s error, got: %v", errorTypeName, err)
	}
}

// AssertNoError fails the test if there is an error
func AssertNoError(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

// AssertEqual provides a simple equality assertion
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}
