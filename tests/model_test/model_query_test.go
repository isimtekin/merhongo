package model_test

import (
	"context"
	"github.com/isimtekin/merhongo/tests/testutil"
	"testing"

	"github.com/isimtekin/merhongo/errors"
	"github.com/isimtekin/merhongo/model"
	"github.com/isimtekin/merhongo/query"
	"github.com/isimtekin/merhongo/schema"
	"go.mongodb.org/mongo-driver/bson"
)

func TestModel_FindWithQuery_Success(t *testing.T) {
	model, cleanup := setupTestCollection(t, "query_test_users")
	defer cleanup()

	ctx := context.Background()

	// Create a query for active users
	q := query.New().
		Where("active", true).
		SortBy("age", true)

	var users []testutil.TestUser
	err := model.FindWithQuery(ctx, q, &users)
	if err != nil {
		t.Fatalf("FindWithQuery failed: %v", err)
	}

	// Should find 3 active users
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	// Check if sorted by age ascending
	if len(users) >= 2 && users[0].Age > users[1].Age {
		t.Errorf("Expected users to be sorted by age ascending")
	}
}

func TestGenericModel_FindWithQuery_Success(t *testing.T) {
	model, cleanup := setupGenericTestCollection(t, "generic_query_test_users")
	defer cleanup()

	ctx := context.Background()

	// Create a query for active users
	q := query.New().
		Where("active", true).
		SortBy("age", true)

	// Use type-safe FindWithQuery
	users, err := model.FindWithQuery(ctx, q)
	if err != nil {
		t.Fatalf("Type-safe FindWithQuery failed: %v", err)
	}

	// Should find 3 active users
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	// Check if sorted by age ascending
	if len(users) >= 2 && users[0].Age > users[1].Age {
		t.Errorf("Expected users to be sorted by age ascending")
	}
}

func TestGenericModel_FindOneWithQuery_Success(t *testing.T) {
	model, cleanup := setupGenericTestCollection(t, "generic_findone_query_test_users")
	defer cleanup()

	ctx := context.Background()

	// Create a query for a specific user
	q := query.New().
		Where("username", "john_doe")

	// Use type-safe FindOneWithQuery
	user, err := model.FindOneWithQuery(ctx, q)
	if err != nil {
		t.Fatalf("Type-safe FindOneWithQuery failed: %v", err)
	}

	// Check the user data
	if user.Username != "john_doe" || user.Email != "john@example.com" || user.Age != 30 {
		t.Errorf("Unexpected user data: %+v", user)
	}
}

func TestGenericModel_FindOneWithQuery_NotFound(t *testing.T) {
	model, cleanup := setupGenericTestCollection(t, "generic_notfound_query_test_users")
	defer cleanup()

	ctx := context.Background()

	// Create a query for a non-existent user
	q := query.New().
		Where("username", "not_exists")

	// Use type-safe FindOneWithQuery
	_, err := model.FindOneWithQuery(ctx, q)

	if !errors.IsNotFound(err) {
		t.Errorf("Expected NotFound error, got %v", err)
	}
}

func TestGenericModel_CountWithQuery_Success(t *testing.T) {
	model, cleanup := setupGenericTestCollection(t, "generic_count_query_test_users")
	defer cleanup()

	ctx := context.Background()

	// Count active users
	q := query.New().
		Where("active", true)

	// Use type-safe CountWithQuery
	count, err := model.CountWithQuery(ctx, q)
	if err != nil {
		t.Fatalf("Type-safe CountWithQuery failed: %v", err)
	}

	// Should count 3 active users
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	// Count users with age >= 30
	q = query.New().
		GreaterThanOrEqual("age", 30)

	count, err = model.CountWithQuery(ctx, q)
	if err != nil {
		t.Fatalf("Type-safe CountWithQuery failed: %v", err)
	}

	// Should count 2 users with age >= 30
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestGenericModel_UpdateWithQuery_Success(t *testing.T) {
	model, cleanup := setupGenericTestCollection(t, "generic_update_query_test_users")
	defer cleanup()

	ctx := context.Background()

	// Update all inactive users to active
	q := query.New().
		Where("active", false)

	update := map[string]interface{}{
		"active": true,
		"role":   "activated",
	}

	// Use type-safe UpdateWithQuery
	modifiedCount, err := model.UpdateWithQuery(ctx, q, update)
	if err != nil {
		t.Fatalf("Type-safe UpdateWithQuery failed: %v", err)
	}

	// Should update 1 inactive user
	if modifiedCount != 1 {
		t.Errorf("Expected 1 document updated, got %d", modifiedCount)
	}

	// Verify the update with type-safe query
	q = query.New().
		Where("role", "activated")

	users, err := model.FindWithQuery(ctx, q)
	if err != nil {
		t.Fatalf("Type-safe FindWithQuery failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 activated user, got %d", len(users))
	}

	if len(users) > 0 && !users[0].Active {
		t.Errorf("User should be active after update")
	}
}

func TestGenericModel_DeleteWithQuery_Success(t *testing.T) {
	model, cleanup := setupGenericTestCollection(t, "generic_delete_query_test_users")
	defer cleanup()

	ctx := context.Background()

	// Count total users before delete
	totalBefore, _ := model.Count(ctx, bson.M{})

	// Delete users with role = "admin"
	q := query.New().
		Where("role", "admin")

	// Use type-safe DeleteWithQuery
	deletedCount, err := model.DeleteWithQuery(ctx, q)
	if err != nil {
		t.Fatalf("Type-safe DeleteWithQuery failed: %v", err)
	}

	// Should delete 1 admin user
	if deletedCount != 1 {
		t.Errorf("Expected 1 document deleted, got %d", deletedCount)
	}

	// Count total users after delete
	totalAfter, _ := model.Count(ctx, bson.M{})

	// Verify one user was deleted
	if totalAfter != totalBefore-1 {
		t.Errorf("Expected %d users after delete, got %d", totalBefore-1, totalAfter)
	}

	// Try to find the deleted user with type-safe query
	q = query.New().Where("role", "admin")
	_, err = model.FindOneWithQuery(ctx, q)

	// Should return NotFound error
	if !errors.IsNotFound(err) {
		t.Errorf("Expected NotFound error, got %v", err)
	}
}

func TestGenericModel_QueryBuilder_ComplexQuery(t *testing.T) {
	model, cleanup := setupGenericTestCollection(t, "generic_complex_query_test_users")
	defer cleanup()

	ctx := context.Background()

	// Create a complex query with multiple conditions
	q := query.New().
		Where("active", true).
		GreaterThan("age", 20).
		In("role", []string{"user", "admin"}).
		NotEquals("username", "bob_jones").
		SortBy("age", false).
		Limit(2)

	// Use type-safe FindWithQuery
	users, err := model.FindWithQuery(ctx, q)
	if err != nil {
		t.Fatalf("Type-safe FindWithQuery failed: %v", err)
	}

	// Check the number of results
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	// Check if sorted by age descending
	if len(users) >= 2 && users[0].Age < users[1].Age {
		t.Errorf("Expected users to be sorted by age descending")
	}

	// Check that all users match the criteria
	for _, user := range users {
		if !user.Active {
			t.Errorf("Expected user to be active: %+v", user)
		}
		if user.Age <= 20 {
			t.Errorf("Expected user age > 20, got %d", user.Age)
		}
		if user.Role != "user" && user.Role != "admin" {
			t.Errorf("Expected user role to be 'user' or 'admin', got %s", user.Role)
		}
		if user.Username == "bob_jones" {
			t.Errorf("Did not expect bob_jones in results")
		}
	}
}

func TestGenericModel_QueryWithNilCollection(t *testing.T) {
	ctx := context.Background()

	s := schema.New(map[string]schema.Field{}, schema.WithCollection("invalid"))
	m := model.NewGeneric[testutil.TestUser]("Dummy", s, nil) // nil client.Database

	// Test FindWithQuery
	q := query.New().Where("active", true)
	_, err := m.FindWithQuery(ctx, q)
	if !errors.IsNilCollectionError(err) {
		t.Errorf("Expected NilCollection error for FindWithQuery, got %v", err)
	}

	// Test FindOneWithQuery
	_, err = m.FindOneWithQuery(ctx, q)
	if !errors.IsNilCollectionError(err) {
		t.Errorf("Expected NilCollection error for FindOneWithQuery, got %v", err)
	}

	// Test CountWithQuery
	_, err = m.CountWithQuery(ctx, q)
	if !errors.IsNilCollectionError(err) {
		t.Errorf("Expected NilCollection error for CountWithQuery, got %v", err)
	}

	// Test UpdateWithQuery
	_, err = m.UpdateWithQuery(ctx, q, bson.M{"active": true})
	if !errors.IsNilCollectionError(err) {
		t.Errorf("Expected NilCollection error for UpdateWithQuery, got %v", err)
	}

	// Test DeleteWithQuery
	_, err = m.DeleteWithQuery(ctx, q)
	if !errors.IsNilCollectionError(err) {
		t.Errorf("Expected NilCollection error for DeleteWithQuery, got %v", err)
	}
}

func TestGenericModel_QueryBuilder_ErrorPropagation(t *testing.T) {
	model, cleanup := setupGenericTestCollection(t, "generic_error_query_test_users")
	defer cleanup()

	ctx := context.Background()

	// Create a query with an error (empty key)
	q := query.New().
		Where("active", true).
		Where("", "invalid") // This creates an error

	// Test all query methods with the invalid query
	_, err := model.FindWithQuery(ctx, q)
	if err == nil {
		t.Error("Expected error from FindWithQuery, got nil")
	}

	_, err = model.FindOneWithQuery(ctx, q)
	if err == nil {
		t.Error("Expected error from FindOneWithQuery, got nil")
	}

	_, err = model.CountWithQuery(ctx, q)
	if err == nil {
		t.Error("Expected error from CountWithQuery, got nil")
	}

	_, err = model.UpdateWithQuery(ctx, q, bson.M{"active": false})
	if err == nil {
		t.Error("Expected error from UpdateWithQuery, got nil")
	}

	_, err = model.DeleteWithQuery(ctx, q)
	if err == nil {
		t.Error("Expected error from DeleteWithQuery, got nil")
	}
}
