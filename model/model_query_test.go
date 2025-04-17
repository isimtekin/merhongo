package model

import (
	"context"
	"github.com/isimtekin/merhongo/connection"
	"github.com/isimtekin/merhongo/errors"
	"github.com/isimtekin/merhongo/query"
	"github.com/isimtekin/merhongo/schema"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"
)

// TestQueryUser represents a user for query testing
type TestQueryUser struct {
	ID        interface{} `bson:"_id,omitempty"`
	Username  string      `bson:"username"`
	Email     string      `bson:"email"`
	Age       int         `bson:"age"`
	Active    bool        `bson:"active"`
	Role      string      `bson:"role"`
	CreatedAt time.Time   `bson:"createdAt"`
	UpdatedAt time.Time   `bson:"updatedAt"`
}

// setupQueryTestCollection creates a test collection with sample data
func setupQueryTestCollection(t *testing.T) (*Model, func()) {
	ctx := context.Background()
	client, err := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	userSchema := schema.New(
		map[string]schema.Field{
			"Username": {Required: true, Unique: true},
			"Email":    {Required: true, Unique: true},
			"Age":      {Min: 18},
			"Active":   {Type: true},
			"Role":     {Type: ""},
		},
		schema.WithCollection("query_test_users"),
	)

	// Create the model
	userModel := New("TestQueryUser", userSchema, client.Database)

	// Clear the collection
	_ = userModel.Collection.Drop(ctx)

	// Insert test data
	testUsers := []interface{}{
		&TestQueryUser{
			Username:  "john_doe",
			Email:     "john@example.com",
			Age:       30,
			Active:    true,
			Role:      "user",
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		},
		&TestQueryUser{
			Username:  "jane_smith",
			Email:     "jane@example.com",
			Age:       25,
			Active:    true,
			Role:      "admin",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-12 * time.Hour),
		},
		&TestQueryUser{
			Username:  "bob_jones",
			Email:     "bob@example.com",
			Age:       45,
			Active:    false,
			Role:      "user",
			CreatedAt: time.Now().Add(-12 * time.Hour),
			UpdatedAt: time.Now().Add(-6 * time.Hour),
		},
		&TestQueryUser{
			Username:  "alice_wonder",
			Email:     "alice@example.com",
			Age:       22,
			Active:    true,
			Role:      "user",
			CreatedAt: time.Now().Add(-6 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	_, err = userModel.Collection.InsertMany(ctx, testUsers)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	cleanup := func() {
		_ = userModel.Collection.Drop(ctx)
		_ = client.Disconnect()
	}

	return userModel, cleanup
}

// TestModel_FindWithQuery_Success tests the FindWithQuery method with successful query
func TestModel_FindWithQuery_Success(t *testing.T) {
	model, cleanup := setupQueryTestCollection(t)
	defer cleanup()

	ctx := context.Background()

	// Create a query for active users
	q := query.New().
		Where("active", true).
		SortBy("age", true)

	var users []TestQueryUser
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

// TestModel_FindWithQuery_Empty tests the FindWithQuery method with no matching results
func TestModel_FindWithQuery_Empty(t *testing.T) {
	model, cleanup := setupQueryTestCollection(t)
	defer cleanup()

	ctx := context.Background()

	// Query for non-existent role
	q := query.New().
		Where("role", "guest")

	var users []TestQueryUser
	err := model.FindWithQuery(ctx, q, &users)
	if err != nil {
		t.Fatalf("FindWithQuery failed: %v", err)
	}

	// Should find 0 users
	if len(users) != 0 {
		t.Errorf("Expected 0 users, got %d", len(users))
	}
}

// TestModel_FindWithQuery_NilCollection tests FindWithQuery with a nil collection
func TestModel_FindWithQuery_NilCollection(t *testing.T) {
	nilModel := &Model{
		Name:   "Nil",
		Schema: &schema.Schema{},
	}

	ctx := context.Background()
	q := query.New().Where("active", true)

	var users []TestQueryUser
	err := nilModel.FindWithQuery(ctx, q, &users)

	if !errors.IsNilCollectionError(err) {
		t.Errorf("Expected NilCollection error, got %v", err)
	}
}

// TestModel_FindWithQuery_BuildError tests FindWithQuery with a build error
func TestModel_FindWithQuery_BuildError(t *testing.T) {
	model, cleanup := setupQueryTestCollection(t)
	defer cleanup()

	ctx := context.Background()

	// Create a query with an error (empty key)
	q := query.New().Where("", "invalid")

	var users []TestQueryUser
	err := model.FindWithQuery(ctx, q, &users)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.IsValidationError(err) {
		t.Errorf("Expected validation error, got %v", err)
	}
}

// TestModel_FindOneWithQuery_Success tests the FindOneWithQuery method with a successful query
func TestModel_FindOneWithQuery_Success(t *testing.T) {
	model, cleanup := setupQueryTestCollection(t)
	defer cleanup()

	ctx := context.Background()

	// Create a query for a specific user
	q := query.New().
		Where("username", "john_doe")

	var user TestQueryUser
	err := model.FindOneWithQuery(ctx, q, &user)
	if err != nil {
		t.Fatalf("FindOneWithQuery failed: %v", err)
	}

	// Check the user data
	if user.Username != "john_doe" || user.Email != "john@example.com" || user.Age != 30 {
		t.Errorf("Unexpected user data: %+v", user)
	}
}

// TestModel_FindOneWithQuery_NotFound tests FindOneWithQuery with no matching document
func TestModel_FindOneWithQuery_NotFound(t *testing.T) {
	model, cleanup := setupQueryTestCollection(t)
	defer cleanup()

	ctx := context.Background()

	// Create a query for a non-existent user
	q := query.New().
		Where("username", "not_exists")

	var user TestQueryUser
	err := model.FindOneWithQuery(ctx, q, &user)

	if !errors.IsNotFound(err) {
		t.Errorf("Expected NotFound error, got %v", err)
	}
}

// TestModel_FindOneWithQuery_WithOptions tests FindOneWithQuery with various options
func TestModel_FindOneWithQuery_WithOptions(t *testing.T) {
	model, cleanup := setupQueryTestCollection(t)
	defer cleanup()

	ctx := context.Background()

	// Create a query with various options
	q := query.New().
		Where("active", true).
		SortBy("age", false). // Sort by age descending
		Skip(1).              // Skip the first result
		Limit(1)              // Limit to 1 result

	var user TestQueryUser
	err := model.FindOneWithQuery(ctx, q, &user)
	if err != nil {
		t.Fatalf("FindOneWithQuery failed: %v", err)
	}

	// After skipping the oldest user and sorting by age descending,
	// we should get the second-oldest active user
	// Ages in our test data: [30(john), 25(jane), 45(bob-inactive), 22(alice)]
	// Active users by age descending: [30(john), 25(jane), 22(alice)]
	// Skip first, so expect jane (not alice, as we're sorting by age)
	if user.Username != "jane_smith" {
		t.Errorf("Expected user jane_smith, got %s", user.Username)
	}
}

// TestModel_FindOneWithQuery_NilCollection tests FindOneWithQuery with a nil collection
func TestModel_FindOneWithQuery_NilCollection(t *testing.T) {
	nilModel := &Model{
		Name:   "Nil",
		Schema: &schema.Schema{},
	}

	ctx := context.Background()
	q := query.New().Where("username", "john_doe")

	var user TestQueryUser
	err := nilModel.FindOneWithQuery(ctx, q, &user)

	if !errors.IsNilCollectionError(err) {
		t.Errorf("Expected NilCollection error, got %v", err)
	}
}

// TestModel_CountWithQuery_Success tests the CountWithQuery method with a successful query
func TestModel_CountWithQuery_Success(t *testing.T) {
	model, cleanup := setupQueryTestCollection(t)
	defer cleanup()

	ctx := context.Background()

	// Count active users
	q := query.New().
		Where("active", true)

	count, err := model.CountWithQuery(ctx, q)
	if err != nil {
		t.Fatalf("CountWithQuery failed: %v", err)
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
		t.Fatalf("CountWithQuery failed: %v", err)
	}

	// Should count 2 users with age >= 30
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

// TestModel_CountWithQuery_NilCollection tests CountWithQuery with a nil collection
func TestModel_CountWithQuery_NilCollection(t *testing.T) {
	nilModel := &Model{
		Name:   "Nil",
		Schema: &schema.Schema{},
	}

	ctx := context.Background()
	q := query.New().Where("active", true)

	_, err := nilModel.CountWithQuery(ctx, q)

	if !errors.IsNilCollectionError(err) {
		t.Errorf("Expected NilCollection error, got %v", err)
	}
}

// TestModel_UpdateWithQuery_Success tests the UpdateWithQuery method with a successful query
func TestModel_UpdateWithQuery_Success(t *testing.T) {
	model, cleanup := setupQueryTestCollection(t)
	defer cleanup()

	ctx := context.Background()

	// Update all inactive users to active
	q := query.New().
		Where("active", false)

	update := map[string]interface{}{
		"active": true,
		"role":   "activated",
	}

	modifiedCount, err := model.UpdateWithQuery(ctx, q, update)
	if err != nil {
		t.Fatalf("UpdateWithQuery failed: %v", err)
	}

	// Should update 1 inactive user
	if modifiedCount != 1 {
		t.Errorf("Expected 1 document updated, got %d", modifiedCount)
	}

	// Verify the update
	q = query.New().
		Where("role", "activated")

	var users []TestQueryUser
	err = model.FindWithQuery(ctx, q, &users)
	if err != nil {
		t.Fatalf("FindWithQuery failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 activated user, got %d", len(users))
	}

	if len(users) > 0 && !users[0].Active {
		t.Errorf("User should be active after update")
	}
}

// TestModel_UpdateWithQuery_BsonUpdate tests UpdateWithQuery with a BSON update document
func TestModel_UpdateWithQuery_BsonUpdate(t *testing.T) {
	model, cleanup := setupQueryTestCollection(t)
	defer cleanup()

	ctx := context.Background()

	// Update all users with age > 40 with a $inc operation
	q := query.New().
		GreaterThan("age", 40)

	// Use a BSON update with $inc operator directly
	update := bson.M{
		"$inc": bson.M{"age": 5}, // Increment age by 5
	}

	modifiedCount, err := model.UpdateWithQuery(ctx, q, update)
	if err != nil {
		t.Fatalf("UpdateWithQuery failed: %v", err)
	}

	// Should update 1 user with age > 40
	if modifiedCount != 1 {
		t.Errorf("Expected 1 document updated, got %d", modifiedCount)
	}

	// Verify the update
	var user TestQueryUser
	err = model.FindOneWithQuery(ctx, query.New().Where("username", "bob_jones"), &user)
	if err != nil {
		t.Fatalf("FindOneWithQuery failed: %v", err)
	}

	// Bob's age should now be 50 (45 + 5)
	if user.Age != 50 {
		t.Errorf("Expected age to be 50 after increment, got %d", user.Age)
	}
}

// TestModel_UpdateWithQuery_NilCollection tests UpdateWithQuery with a nil collection
func TestModel_UpdateWithQuery_NilCollection(t *testing.T) {
	nilModel := &Model{
		Name:   "Nil",
		Schema: &schema.Schema{},
	}

	ctx := context.Background()
	q := query.New().Where("active", false)
	update := map[string]interface{}{"active": true}

	_, err := nilModel.UpdateWithQuery(ctx, q, update)

	if !errors.IsNilCollectionError(err) {
		t.Errorf("Expected NilCollection error, got %v", err)
	}
}

// TestModel_DeleteWithQuery_Success tests the DeleteWithQuery method with a successful query
func TestModel_DeleteWithQuery_Success(t *testing.T) {
	model, cleanup := setupQueryTestCollection(t)
	defer cleanup()

	ctx := context.Background()

	// Count total users before delete
	totalBefore, _ := model.Count(ctx, bson.M{})

	// Delete users with role = "admin"
	q := query.New().
		Where("role", "admin")

	deletedCount, err := model.DeleteWithQuery(ctx, q)
	if err != nil {
		t.Fatalf("DeleteWithQuery failed: %v", err)
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

	// Try to find the deleted user
	var adminUser TestQueryUser
	err = model.FindOneWithQuery(ctx, query.New().Where("role", "admin"), &adminUser)

	// Should return NotFound error
	if !errors.IsNotFound(err) {
		t.Errorf("Expected NotFound error, got %v", err)
	}
}

// TestModel_DeleteWithQuery_NoMatch tests DeleteWithQuery with no matching documents
func TestModel_DeleteWithQuery_NoMatch(t *testing.T) {
	model, cleanup := setupQueryTestCollection(t)
	defer cleanup()

	ctx := context.Background()

	// Try to delete users with a non-existent role
	q := query.New().
		Where("role", "not_exists")

	deletedCount, err := model.DeleteWithQuery(ctx, q)
	if err != nil {
		t.Fatalf("DeleteWithQuery failed: %v", err)
	}

	// Should delete 0 users
	if deletedCount != 0 {
		t.Errorf("Expected 0 documents deleted, got %d", deletedCount)
	}
}

// TestModel_DeleteWithQuery_NilCollection tests DeleteWithQuery with a nil collection
func TestModel_DeleteWithQuery_NilCollection(t *testing.T) {
	nilModel := &Model{
		Name:   "Nil",
		Schema: &schema.Schema{},
	}

	ctx := context.Background()
	q := query.New().Where("role", "admin")

	_, err := nilModel.DeleteWithQuery(ctx, q)

	if !errors.IsNilCollectionError(err) {
		t.Errorf("Expected NilCollection error, got %v", err)
	}
}

// TestModel_QueryBuilder_ComplexQuery tests using the query builder for complex queries
func TestModel_QueryBuilder_ComplexQuery(t *testing.T) {
	model, cleanup := setupQueryTestCollection(t)
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

	var users []TestQueryUser
	err := model.FindWithQuery(ctx, q, &users)
	if err != nil {
		t.Fatalf("FindWithQuery failed: %v", err)
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

// TestModel_QueryBuilder_ErrorPropagation tests that builder errors propagate correctly
func TestModel_QueryBuilder_ErrorPropagation(t *testing.T) {
	model, cleanup := setupQueryTestCollection(t)
	defer cleanup()

	ctx := context.Background()

	// Create a query with an error (empty key)
	q := query.New().
		Where("active", true).
		Where("", "invalid") // This creates an error

	// Subsequent operations should not clear the error
	q.GreaterThan("age", 20).
		In("role", []string{"user", "admin"}).
		SortBy("age", false)

	// FindWithQuery
	var users []TestQueryUser
	err := model.FindWithQuery(ctx, q, &users)
	if err == nil {
		t.Error("Expected error from FindWithQuery, got nil")
	}

	// FindOneWithQuery
	var user TestQueryUser
	err = model.FindOneWithQuery(ctx, q, &user)
	if err == nil {
		t.Error("Expected error from FindOneWithQuery, got nil")
	}

	// CountWithQuery
	_, err = model.CountWithQuery(ctx, q)
	if err == nil {
		t.Error("Expected error from CountWithQuery, got nil")
	}

	// UpdateWithQuery
	_, err = model.UpdateWithQuery(ctx, q, bson.M{"active": false})
	if err == nil {
		t.Error("Expected error from UpdateWithQuery, got nil")
	}

	// DeleteWithQuery
	_, err = model.DeleteWithQuery(ctx, q)
	if err == nil {
		t.Error("Expected error from DeleteWithQuery, got nil")
	}
}
