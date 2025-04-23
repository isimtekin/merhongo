package model_test

import (
	"context"
	"fmt"
	"github.com/isimtekin/merhongo/tests/testutil"
	"strings"
	"testing"
	"time"

	"github.com/isimtekin/merhongo/errors"
	"github.com/isimtekin/merhongo/model"
	"github.com/isimtekin/merhongo/query"
	"github.com/isimtekin/merhongo/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestValidateDocumentCalls(t *testing.T) {
	ctx := context.Background()
	client, cleanup := testutil.CreateTestClient(t)
	defer cleanup()

	collName := "validation_calls_test"
	testutil.DropCollection(t, client.Database, collName) // Clean up

	type ValidationTracker struct {
		Calls      int
		LastOp     string
		ShouldFail bool
	}

	tracker := ValidationTracker{0, "", false}

	type TestDocument struct {
		ID        primitive.ObjectID `bson:"_id,omitempty"`
		Name      string             `bson:"name"`
		Age       int                `bson:"age"`
		CreatedAt time.Time          `bson:"createdAt,omitempty"`
		UpdatedAt time.Time          `bson:"updatedAt,omitempty"`
	}

	// Create schema with validation requirements
	s := schema.New(
		map[string]schema.Field{
			"Name": {Required: true},
			"Age":  {Min: 18, Max: 100},
		},
		schema.WithCollection(collName),
		schema.WithTimestamps(true),
	)

	// Set model type and custom validator that tracks calls
	var modelType TestDocument
	s.ModelType = &modelType
	s.CustomValidator = func(doc interface{}) error {
		// Track that validation was called
		tracker.Calls++

		// Check document type
		testDoc, ok := doc.(*TestDocument)
		if !ok {
			return errors.WithDetails(errors.ErrValidation, "invalid document type")
		}

		// If we should fail validation, return error
		if tracker.ShouldFail {
			return errors.WithDetails(errors.ErrValidation, "validation failed as requested")
		}

		// Real validation - check Age within range
		if testDoc.Age < 18 {
			return errors.WithDetails(errors.ErrValidation,
				fmt.Sprintf("age must be at least 18, got %d", testDoc.Age))
		}
		if testDoc.Age > 100 {
			return errors.WithDetails(errors.ErrValidation,
				fmt.Sprintf("age must be at most 100, got %d", testDoc.Age))
		}

		// Real validation - check Name is not empty
		if testDoc.Name == "" {
			return errors.WithDetails(errors.ErrValidation, "name cannot be empty")
		}

		return nil
	}

	// Create model
	m := model.NewGeneric[TestDocument]("TestDocument", s, client.Database)

	// ---------------- CREATE TESTS ----------------

	// Test 1: Valid document create
	tracker = ValidationTracker{0, "create_valid", false}
	validDoc := &TestDocument{
		Name: "valid-person",
		Age:  25,
	}

	err := m.Create(ctx, validDoc)
	if err != nil {
		t.Fatalf("Valid create failed: %v", err)
	}

	if tracker.Calls != 1 {
		t.Errorf("ValidateDocument should be called exactly once for Create, got %d calls", tracker.Calls)
	}

	// Test 2: Invalid document create (missing required field)
	tracker = ValidationTracker{0, "create_invalid_name", false}
	invalidDoc1 := &TestDocument{
		Name: "", // Empty name - should fail validation
		Age:  30,
	}

	err = m.Create(ctx, invalidDoc1)
	if err == nil {
		t.Error("Create should fail validation for empty name")
	}

	if tracker.Calls != 1 {
		t.Errorf("ValidateDocument should be called exactly once for invalid Create, got %d calls", tracker.Calls)
	}

	// Check error type
	if !errors.IsValidationError(err) {
		t.Errorf("Expected validation error for empty name, got: %v", err)
	}

	// Test 3: Invalid document create (age out of range)
	tracker = ValidationTracker{0, "create_invalid_age", false}
	invalidDoc2 := &TestDocument{
		Name: "underage-person",
		Age:  16, // Under 18 - should fail validation
	}

	err = m.Create(ctx, invalidDoc2)
	if err == nil {
		t.Error("Create should fail validation for underage person")
	}

	if tracker.Calls != 1 {
		t.Errorf("ValidateDocument should be called exactly once for invalid age Create, got %d calls", tracker.Calls)
	}

	// Check error message
	if err != nil && !strings.Contains(err.Error(), "age must be at least 18") {
		t.Errorf("Expected specific age validation error, got: %v", err)
	}

	// ---------------- UPDATE TESTS ----------------

	// Test 4: Valid document update
	tracker = ValidationTracker{0, "update_valid", false}

	err = m.UpdateById(ctx, validDoc.ID.Hex(), bson.M{
		"age": 35,
	})

	if err != nil {
		t.Fatalf("Valid update failed: %v", err)
	}

	if tracker.Calls != 1 {
		t.Errorf("ValidateDocument should be called exactly once for Update, got %d calls", tracker.Calls)
	}

	// Verify update was applied
	updatedDoc, err := m.FindById(ctx, validDoc.ID.Hex())
	if err != nil {
		t.Fatalf("FindById failed: %v", err)
	}

	if updatedDoc.Age != 35 {
		t.Errorf("Update failed to change age, expected 35, got %d", updatedDoc.Age)
	}

	// Test 5: Invalid update (age out of range)
	tracker = ValidationTracker{0, "update_invalid_age", false}

	err = m.UpdateById(ctx, validDoc.ID.Hex(), bson.M{
		"age": 120, // Over 100 - should fail validation
	})

	if err == nil {
		t.Error("Update should fail validation for age over 100")
	}

	if tracker.Calls != 1 {
		t.Errorf("ValidateDocument should be called exactly once for invalid Update, got %d calls", tracker.Calls)
	}

	// Check error is validation error
	if !errors.IsValidationError(err) {
		t.Errorf("Expected validation error for age > 100, got: %v", err)
	}

	// Check error message
	if err != nil && !strings.Contains(err.Error(), "age must be at most 100") {
		t.Errorf("Expected specific age validation error, got: %v", err)
	}

	// Test 6: Invalid update (empty name)
	tracker = ValidationTracker{0, "update_invalid_name", false}

	err = m.UpdateById(ctx, validDoc.ID.Hex(), bson.M{
		"name": "", // Empty name - should fail validation
	})

	if err == nil {
		t.Error("Update should fail validation for empty name")
	}

	if tracker.Calls != 1 {
		t.Errorf("ValidateDocument should be called exactly once for invalid name Update, got %d calls", tracker.Calls)
	}

	// Test 7: Force validation failure
	tracker = ValidationTracker{0, "update_forced_failure", true}

	err = m.UpdateById(ctx, validDoc.ID.Hex(), bson.M{
		"age": 40, // Valid age but validator will force failure
	})

	if err == nil {
		t.Error("Update should fail when validator is forced to fail")
	}

	if tracker.Calls != 1 {
		t.Errorf("ValidateDocument should be called exactly once for forced failure, got %d calls", tracker.Calls)
	}

	// ---------------- QUERY UPDATE TESTS ----------------

	// Create a few more documents for query tests
	for i := 0; i < 3; i++ {
		doc := &TestDocument{
			Name: fmt.Sprintf("person-%d", i),
			Age:  20 + i,
		}
		tracker.ShouldFail = false
		_ = m.Create(ctx, doc) // Ignore validation calls for these
	}

	// Test 8: UpdateWithQuery valid update
	tracker = ValidationTracker{0, "query_update_valid", false}

	q := query.New().Where("age", 20) // Should match 1 document
	count, err := m.UpdateWithQuery(ctx, q, bson.M{
		"age": 50,
	})

	if err != nil {
		t.Fatalf("UpdateWithQuery failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected to update 1 document, got %d", count)
	}

	if tracker.Calls != 1 {
		t.Errorf("ValidateDocument should be called once for UpdateWithQuery, got %d calls", tracker.Calls)
	}

	// Test 9: UpdateWithQuery with invalid update
	tracker = ValidationTracker{0, "query_update_invalid", false}

	q = query.New().GreaterThan("age", 20) // Should match multiple documents
	_, err = m.UpdateWithQuery(ctx, q, bson.M{
		"age": 130, // Over 100 - should fail validation
	})

	if err == nil {
		t.Error("UpdateWithQuery should fail validation for age over 100")
	}

	// Validation should be called for the first matching document
	if tracker.Calls != 1 {
		t.Errorf("ValidateDocument should be called at least once for invalid UpdateWithQuery, got %d calls", tracker.Calls)
	}

	// Check error type
	if !errors.IsValidationError(err) {
		t.Errorf("Expected validation error for UpdateWithQuery, got: %v", err)
	}
}

func TestCustomValidatorBehavior(t *testing.T) {
	ctx := context.Background()
	client, cleanup := testutil.CreateTestClient(t)
	defer cleanup()

	collName := "custom_validator_test"
	testutil.DropCollection(t, client.Database, collName) // Clean up

	type DataDocument struct {
		ID    primitive.ObjectID `bson:"_id,omitempty"`
		Key   string             `bson:"key"`
		Value interface{}        `bson:"value"`
		Type  string             `bson:"type"`
	}

	// Create schema
	s := schema.New(
		map[string]schema.Field{
			"Key":   {Required: true},
			"Value": {Required: true},
			"Type":  {Required: true},
		},
		schema.WithCollection(collName),
	)

	// Add custom validator that checks type consistency
	validationErrors := []string{}
	s.CustomValidator = func(doc interface{}) error {
		validationErrors = []string{} // Reset errors

		dataDoc, ok := doc.(*DataDocument)
		if !ok {
			return errors.WithDetails(errors.ErrValidation, "invalid document type")
		}

		// Validate key is not empty
		if dataDoc.Key == "" {
			validationErrors = append(validationErrors, "key cannot be empty")
		}

		// Type consistency validation
		if dataDoc.Value != nil {
			actualType := ""
			switch dataDoc.Value.(type) {
			case int, int32, int64, float32, float64:
				actualType = "number"
			case string:
				actualType = "string"
			case bool:
				actualType = "boolean"
			case []interface{}, primitive.A:
				actualType = "array"
			case map[string]interface{}, primitive.D, primitive.M:
				actualType = "object"
			default:
				actualType = "unknown"
			}

			// Check if declared type matches actual type
			if dataDoc.Type != actualType && actualType != "unknown" {
				validationErrors = append(validationErrors,
					fmt.Sprintf("type mismatch: declared as '%s' but value is '%s'",
						dataDoc.Type, actualType))
			}
		}

		if len(validationErrors) > 0 {
			return errors.WithDetails(errors.ErrValidation,
				fmt.Sprintf("validation failed: %s", strings.Join(validationErrors, "; ")))
		}

		return nil
	}

	// Create model
	m := model.NewGeneric[DataDocument]("DataDocument", s, client.Database)

	// Test valid document
	validDoc := &DataDocument{
		Key:   "age",
		Value: 25,
		Type:  "number",
	}

	err := m.Create(ctx, validDoc)
	if err != nil {
		t.Fatalf("Valid create failed: %v", err)
	}

	// Test type mismatch
	invalidDoc := &DataDocument{
		Key:   "name",
		Value: "John",
		Type:  "number", // Mismatch - value is string but type says number
	}

	err = m.Create(ctx, invalidDoc)
	if err == nil {
		t.Error("Create should fail validation for type mismatch")
	}

	if !errors.IsValidationError(err) {
		t.Errorf("Expected validation error for type mismatch, got: %v", err)
	}

	if !strings.Contains(err.Error(), "type mismatch") {
		t.Errorf("Expected type mismatch error message, got: %v", err)
	}

	// Test with multiple validation errors
	multiErrorDoc := &DataDocument{
		Key:   "", // Empty key - error 1
		Value: true,
		Type:  "number", // Type mismatch - error 2
	}

	err = m.Create(ctx, multiErrorDoc)
	if err == nil {
		t.Error("Create should fail with multiple validation errors")
	}

	if !errors.IsValidationError(err) {
		t.Errorf("Expected validation error, got: %v", err)
	}

	// Should contain both error messages
	if !strings.Contains(err.Error(), "key cannot be empty") ||
		!strings.Contains(err.Error(), "type mismatch") {
		t.Errorf("Expected multiple validation error messages, got: %v", err)
	}
}
