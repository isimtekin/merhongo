package model_test

import (
	"context"
	"github.com/isimtekin/merhongo/tests/testutil"
	"testing"
	"time"

	"github.com/isimtekin/merhongo/model"
	"github.com/isimtekin/merhongo/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGenericModel_TimestampBehavior(t *testing.T) {
	ctx := context.Background()
	client, cleanup := testutil.CreateTestClient(t)
	defer cleanup()

	collName := "timestamped_docs"
	testutil.DropCollection(t, client.Database, collName) // Clean up

	type TimestampedDoc struct {
		ID        primitive.ObjectID `bson:"_id,omitempty"`
		Name      string             `bson:"name"`
		CreatedAt time.Time          `bson:"createdAt,omitempty"`
		UpdatedAt time.Time          `bson:"updatedAt,omitempty"`
	}

	// Test with timestamps enabled
	s := schema.New(map[string]schema.Field{
		"Name": {Required: true},
	}, schema.WithCollection(collName), schema.WithTimestamps(true))

	m := model.NewGeneric[TimestampedDoc]("TimestampedDoc", s, client.Database)

	// Create a document
	doc := &TimestampedDoc{Name: "timestamp-test"}
	err := m.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Store the original ID and timestamps
	originalID := doc.ID
	originalCreatedAtTime := doc.CreatedAt
	time.Sleep(time.Millisecond * 10) // Ensure time difference

	err = m.UpdateById(ctx, doc.ID.Hex(), bson.M{"name": "updated-name"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Retrieve the updated document
	updated, err := m.FindById(ctx, doc.ID.Hex())
	if err != nil {
		t.Fatalf("FindById failed: %v", err)
	}

	// UpdatedAt should change
	if updated.UpdatedAt.Equal(doc.UpdatedAt) {
		t.Error("expected UpdatedAt to change after update")
	}

	// Convert timestamps to Unix time for more reliable comparison
	createdAtOriginalUnix := originalCreatedAtTime.Unix()
	createdAtUpdatedUnix := updated.CreatedAt.Unix()

	// Check if the Unix timestamps are equal
	if createdAtOriginalUnix != createdAtUpdatedUnix {
		t.Errorf("CreatedAt value should be preserved. Original: %v, Updated: %v",
			originalCreatedAtTime, updated.CreatedAt)
	}

	// ID should not change
	if updated.ID != originalID {
		t.Error("ID value changed, original ID should be preserved")
	}

	// Name should be updated
	if updated.Name != "updated-name" {
		t.Errorf("Name was not updated correctly. Expected: 'updated-name', Got: '%s'", updated.Name)
	}
}

func TestTimestampAndValidation(t *testing.T) {
	ctx := context.Background()
	client, cleanup := testutil.CreateTestClient(t)
	defer cleanup()

	collName := "timestamp_validation_test_docs"
	testutil.DropCollection(t, client.Database, collName) // Clean up

	type TestDocument struct {
		ID        primitive.ObjectID `bson:"_id,omitempty"`
		Name      string             `bson:"name"`
		Value     int                `bson:"value"`
		CreatedAt time.Time          `bson:"createdAt,omitempty"`
		UpdatedAt time.Time          `bson:"updatedAt,omitempty"`
	}

	// Validation flag to track when validation is called
	validationCalled := false

	// Create schema with custom validator
	s := schema.New(
		map[string]schema.Field{
			"Name":  {Required: true},
			"Value": {Min: 1, Max: 100},
		},
		schema.WithCollection(collName),
		schema.WithTimestamps(true),
	)

	// Set model type and custom validator
	var modelType TestDocument
	s.ModelType = &modelType
	s.CustomValidator = func(doc interface{}) error {
		validationCalled = true

		// Validation passes
		return nil
	}

	// Create model
	m := model.NewGeneric[TestDocument]("TestDocument", s, client.Database)

	// Test Create operation with validation
	validationCalled = false
	doc := &TestDocument{
		Name:  "test-doc",
		Value: 50,
	}

	err := m.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if !validationCalled {
		t.Error("Create should call ValidateDocument")
	}

	// Verify timestamps were set
	if doc.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set during Create")
	}
	if doc.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set during Create")
	}

	// Store original ID and values for comparison
	originalID := doc.ID
	_ = doc.Value

	// Convert timestamps to UTC for consistent comparison
	originalCreatedAtUTC := doc.CreatedAt.UTC()
	originalUpdatedAtUTC := doc.UpdatedAt.UTC()

	// Ensure time difference by waiting a bit longer
	time.Sleep(time.Millisecond * 50)

	// Test UpdateById with validation
	validationCalled = false

	// Try to update including createdAt (should be ignored)
	err = m.UpdateById(ctx, doc.ID.Hex(), bson.M{
		"value":     75,
		"createdAt": time.Now().Add(-10 * time.Hour), // This should be ignored
	})

	if err != nil {
		t.Fatalf("UpdateById failed: %v", err)
	}

	if !validationCalled {
		t.Error("UpdateById should call ValidateDocument")
	}

	// Verify document after update
	updatedDoc, err := m.FindById(ctx, doc.ID.Hex())
	if err != nil {
		t.Fatalf("FindById failed: %v", err)
	}

	// Check if values were updated correctly
	if updatedDoc.Value != 75 {
		t.Errorf("Expected Value=75 after update, got %d", updatedDoc.Value)
	}

	// ID should not change
	if updatedDoc.ID != originalID {
		t.Error("ID should not change during update")
	}

	// Convert retrieved timestamps to UTC for consistent comparison
	updatedCreatedAtUTC := updatedDoc.CreatedAt.UTC()
	updatedUpdatedAtUTC := updatedDoc.UpdatedAt.UTC()

	// For TimestampAndValidation test, use string comparison of the formatted dates
	// This is to avoid issues with precision in time comparison
	createdAtOriginalStr := originalCreatedAtUTC.Format(time.RFC3339)
	createdAtUpdatedStr := updatedCreatedAtUTC.Format(time.RFC3339)

	if createdAtOriginalStr != createdAtUpdatedStr {
		t.Errorf("CreatedAt should not change. Original=%s, Updated=%s",
			createdAtOriginalStr, createdAtUpdatedStr)
	}

	// Compare UpdatedAt by checking if the updated time is after the original
	if !updatedUpdatedAtUTC.After(originalUpdatedAtUTC) {
		t.Errorf("UpdatedAt should be more recent than original. Original=%v, Updated=%v",
			originalUpdatedAtUTC, updatedUpdatedAtUTC)
	}
}
