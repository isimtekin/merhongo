package model_test

import (
	"context"
	testutil2 "github.com/isimtekin/merhongo/tests/testutil"
	"testing"

	"github.com/isimtekin/merhongo/model"
	"go.mongodb.org/mongo-driver/mongo"
)

// setupTestCollection creates a test collection with sample data for regular model
func setupTestCollection(t *testing.T, collectionName string) (*model.Model, func()) {
	client, cleanup := testutil2.CreateTestClient(t)

	userSchema := testutil2.CreateTestSchema(collectionName)
	userModel := model.New("TestUser", userSchema, client.Database)

	// Clear the collection
	testutil2.DropCollection(t, client.Database, collectionName)

	// Insert test data
	testUsers := []interface{}{}
	for _, user := range testutil2.CreateTestUsers() {
		testUsers = append(testUsers, &user)
	}

	testutil2.InsertDocuments(t, userModel.Collection, testUsers)

	modelCleanup := func() {
		testutil2.DropCollection(t, client.Database, collectionName)
		cleanup()
	}

	return userModel, modelCleanup
}

// setupGenericTestCollection creates a test collection with sample data for generic model
func setupGenericTestCollection(t *testing.T, collectionName string) (*model.GenericModel[testutil2.TestUser], func()) {
	ctx := context.Background()
	client, cleanup := testutil2.CreateTestClient(t)

	userSchema := testutil2.CreateTestSchema(collectionName)
	userModel := model.NewGeneric[testutil2.TestUser]("TestUser", userSchema, client.Database)

	// Clear the collection
	testutil2.DropCollection(t, client.Database, collectionName)

	// Insert test data for generic model
	for _, user := range testutil2.CreateTestUsers() {
		userCopy := user // Create a copy to avoid issues with references
		err := userModel.Create(ctx, &userCopy)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	modelCleanup := func() {
		testutil2.DropCollection(t, client.Database, collectionName)
		cleanup()
	}

	return userModel, modelCleanup
}

// prepareCleanDatabase creates a clean database for a test
func prepareCleanDatabase(t *testing.T, collectionName string) (*mongo.Database, func()) {
	client, cleanup := testutil2.CreateTestClient(t)
	db := client.Database

	// Ensure the collection is clean
	testutil2.DropCollection(t, db, collectionName)

	return db, cleanup
}
