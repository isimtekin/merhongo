package model_test

import (
	"context"
	"fmt"
	"github.com/isimtekin/merhongo"
	"github.com/isimtekin/merhongo/tests/testutil"
	"strings"
	"testing"
	_ "time"

	"github.com/isimtekin/merhongo/errors"
	"github.com/isimtekin/merhongo/model"
	"github.com/isimtekin/merhongo/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserModel_CreateAndFindOne(t *testing.T) {
	ctx := context.Background()

	client, cleanup := testutil.CreateTestClient(t)
	defer cleanup()

	// Clean up collections before test
	testutil.DropCollection(t, client.Database, "users_basic")
	testutil.DropCollection(t, client.Database, "users_generic")

	// Test with regular model
	userSchema := schema.New(
		map[string]schema.Field{
			"Username": {Required: true, Unique: true},
			"Email":    {Required: true, Unique: true},
			"Age":      {Min: 18, Max: 100},
		},
		schema.WithCollection("users_basic"),
	)

	regularModel := model.New("User", userSchema, client.Database)

	user := &testutil.TestUser{
		Username: "testuser",
		Email:    "test@example.com",
		Age:      27,
		Active:   true,
	}

	err := regularModel.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var result testutil.TestUser
	err = regularModel.FindOne(ctx, map[string]interface{}{"username": "testuser"}, &result)
	if err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}

	if result.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, result.Email)
	}

	// Test with generic model - use a different collection to avoid conflicts
	genericSchema := schema.New(
		map[string]schema.Field{
			"Username": {Required: true, Unique: true},
			"Email":    {Required: true, Unique: true},
			"Age":      {Min: 18, Max: 100},
		},
		schema.WithCollection("users_generic"),
	)

	genericModel := model.NewGeneric[testutil.TestUser]("GenericUser", genericSchema, client.Database)

	user2 := &testutil.TestUser{
		Username: "genericuser",
		Email:    "generic@example.com",
		Age:      30,
		Active:   true,
	}

	err = genericModel.Create(ctx, user2)
	if err != nil {
		t.Fatalf("Generic Create failed: %v", err)
	}

	// Test type-safe FindOne
	foundUser, err := genericModel.FindOne(ctx, map[string]interface{}{"username": "genericuser"})
	if err != nil {
		t.Fatalf("Generic FindOne failed: %v", err)
	}

	if foundUser.Email != user2.Email {
		t.Errorf("expected email %s, got %s", user2.Email, foundUser.Email)
	}

	// Test type-safe FindById
	foundById, err := genericModel.FindById(ctx, user2.ID.Hex())
	if err != nil {
		t.Fatalf("Generic FindById failed: %v", err)
	}

	if foundById.Username != user2.Username {
		t.Errorf("expected username %s, got %s", user2.Username, foundById.Username)
	}

	// Test type-safe Find
	allUsers, err := genericModel.Find(ctx, bson.M{})
	if err != nil {
		t.Fatalf("Generic Find failed: %v", err)
	}

	if len(allUsers) != 1 { // Only one user in the generic_users collection
		t.Errorf("expected 1 user, got %d", len(allUsers))
	}
}

func TestGenericModel_CRUDOperations(t *testing.T) {
	ctx := context.Background()
	client, cleanup := testutil.CreateTestClient(t)
	defer cleanup()

	collName := "users_crud_test"
	testutil.DropCollection(t, client.Database, collName) // Clean up

	userSchema := schema.New(map[string]schema.Field{
		"Username": {Required: true, Unique: true},
		"Age":      {Min: 0},
	}, schema.WithCollection(collName))

	userModel := model.NewGeneric[testutil.TestUser]("User", userSchema, client.Database)

	user := &testutil.TestUser{Username: "john_doe", Age: 28, Active: true}
	err := userModel.Create(ctx, user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Test type-safe Count
	count, err := userModel.Count(ctx, bson.M{"username": "john_doe"})
	if err != nil || count != 1 {
		t.Errorf("expected 1 user, got %d", count)
	}

	// Test type-safe FindById
	id := user.ID.Hex()
	found, err := userModel.FindById(ctx, id)
	if err != nil || found.Username != "john_doe" {
		t.Errorf("expected to find user by ID")
	}

	// Test UpdateById
	err = userModel.UpdateById(ctx, id, bson.M{"age": 30})
	if err != nil {
		t.Errorf("failed to update user: %v", err)
	}

	// Verify update with type-safe FindById
	updated, err := userModel.FindById(ctx, id)
	if err != nil || updated.Age != 30 {
		t.Errorf("expected updated age to be 30, got %d", updated.Age)
	}

	// Test DeleteById
	err = userModel.DeleteById(ctx, id)
	if err != nil {
		t.Errorf("failed to delete user: %v", err)
	}

	// Verify deletion
	_, err = userModel.FindById(ctx, id)
	if !errors.IsNotFound(err) {
		t.Errorf("expected NotFound error, got %v", err)
	}
}

func TestModel_Create_WithMiddlewareError(t *testing.T) {
	ctx := context.Background()
	client, cleanup := testutil.CreateTestClient(t)
	defer cleanup()

	collName := "dummies_middleware"
	testutil.DropCollection(t, client.Database, collName) // Clean up

	type Dummy struct {
		ID   primitive.ObjectID `bson:"_id,omitempty"`
		Name string             `bson:"name"`
	}

	s := schema.New(map[string]schema.Field{
		"Name": {Required: true},
	}, schema.WithCollection(collName))

	// Add middleware that always fails
	s.Pre("save", func(doc interface{}) error {
		return fmt.Errorf("middleware failed")
	})

	// Test with generic model
	m := model.NewGeneric[Dummy]("Dummy", s, client.Database)

	doc := &Dummy{Name: "test"}
	err := m.Create(ctx, doc)

	// Use the helper function instead of checking the exact error message
	if !errors.IsMiddlewareError(err) {
		t.Errorf("expected middleware failure, got %v", err)
	}

	// Check that the details contain the original error message
	if err != nil && !strings.Contains(err.Error(), "middleware failed") {
		t.Errorf("expected error to contain 'middleware failed', got: %v", err)
	}
}

func TestGenericModel_Create_WithValidationError(t *testing.T) {
	ctx := context.Background()
	client, cleanup := testutil.CreateTestClient(t)
	defer cleanup()

	collName := "dummies_validation"
	testutil.DropCollection(t, client.Database, collName) // Clean up

	type Dummy struct {
		ID   primitive.ObjectID `bson:"_id,omitempty"`
		Name string             `bson:"name"`
	}

	s := schema.New(map[string]schema.Field{
		"Name": {Required: true},
	}, schema.WithCollection(collName))

	// Override validation temporarily
	s.CustomValidator = func(doc interface{}) error {
		return errors.ErrValidation
	}

	m := model.NewGeneric[Dummy]("Dummy", s, client.Database)

	doc := &Dummy{Name: "fail"}
	err := m.Create(ctx, doc)

	// Use the helper function instead of errors.Is
	if !errors.IsValidationError(err) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

func TestGenericModel_Find_InvalidCollection(t *testing.T) {
	ctx := context.Background()

	type Dummy struct {
		ID primitive.ObjectID `bson:"_id,omitempty"`
	}

	s := schema.New(map[string]schema.Field{}, schema.WithCollection("invalid"))
	m := model.NewGeneric[Dummy]("Dummy", s, nil) // nil client.Database

	_, err := m.Find(ctx, bson.M{})

	// Use the helper function instead of errors.Is
	if !errors.IsNilCollectionError(err) {
		t.Errorf("expected ErrNilCollection, got %v", err)
	}
}

func TestGenericModel_FindOne_NotFound(t *testing.T) {
	ctx := context.Background()
	client, cleanup := testutil.CreateTestClient(t)
	defer cleanup()

	collName := "dummies_notfound"
	testutil.DropCollection(t, client.Database, collName) // Clean up

	type Dummy struct {
		ID primitive.ObjectID `bson:"_id,omitempty"`
	}

	s := schema.New(map[string]schema.Field{}, schema.WithCollection(collName))
	m := model.NewGeneric[Dummy]("Dummy", s, client.Database)

	_, err := m.FindOne(ctx, bson.M{"username": "not_exists"})

	// Use the helper function instead of errors.Is
	if !errors.IsNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGenericModel_FindById_InvalidObjectID(t *testing.T) {
	ctx := context.Background()
	client, cleanup := testutil.CreateTestClient(t)
	defer cleanup()

	collName := "dummies_invalid_id"
	testutil.DropCollection(t, client.Database, collName) // Clean up

	type Dummy struct {
		ID primitive.ObjectID `bson:"_id,omitempty"`
	}

	s := schema.New(map[string]schema.Field{}, schema.WithCollection(collName))
	m := model.NewGeneric[Dummy]("Dummy", s, client.Database)

	_, err := m.FindById(ctx, "not-an-objectid")

	// Use the helper function instead of errors.Is
	if !errors.IsInvalidObjectID(err) {
		t.Errorf("expected ErrInvalidObjectID, got: %v", err)
	}
}

// Test partial update scenarios with a simpler approach

func TestBasicPartialUpdate(t *testing.T) {
	ctx := context.Background()
	client, cleanup := testutil.CreateTestClient(t)
	defer cleanup()

	// Use a single collection for this test and delete it first
	collName := "basic_partial_update"
	testutil.DropCollection(t, client.Database, collName)

	// Create schema and model
	userSchema := schema.New(map[string]schema.Field{
		"Username": {Required: true, Type: ""},
		"Email":    {Required: true, Type: ""},
		"Age":      {Min: 0},
		"Active":   {Type: true},
		"Role":     {Type: ""},
	}, schema.WithCollection(collName), schema.WithTimestamps(true))

	userModel := model.NewGeneric[testutil.TestUser]("BasicPartialUser", userSchema, client.Database)

	// Create a test user
	user := &testutil.TestUser{
		Username: "testuser",
		Email:    "test@example.com",
		Age:      28,
		Active:   true,
		Role:     "user",
	}

	err := userModel.Create(ctx, user)
	testutil.AssertNoError(t, err, "Failed to create user")

	// Update only the email field
	err = userModel.UpdateById(ctx, user.ID.Hex(), bson.M{"email": "updated@example.com"})
	testutil.AssertNoError(t, err, "Failed to update email")

	// Find the user and verify fields
	updated, err := userModel.FindById(ctx, user.ID.Hex())
	testutil.AssertNoError(t, err, "Failed to find user")

	// Email should be updated
	testutil.AssertEqual(t, "updated@example.com", updated.Email, "Email should be updated")

	// Other fields should be preserved
	testutil.AssertEqual(t, "testuser", updated.Username, "Username should be preserved")
	testutil.AssertEqual(t, 28, updated.Age, "Age should be preserved")
	testutil.AssertEqual(t, true, updated.Active, "Active should be preserved")
	testutil.AssertEqual(t, "user", updated.Role, "Role should be preserved")
}

func TestMultiFieldUpdate(t *testing.T) {
	ctx := context.Background()
	client, cleanup := testutil.CreateTestClient(t)
	defer cleanup()

	// Use a separate collection for this test and clean it first
	collName := "multi_field_update"
	testutil.DropCollection(t, client.Database, collName)

	// Create schema and model
	userSchema := schema.New(map[string]schema.Field{
		"Username": {Required: true, Type: ""},
		"Email":    {Required: true, Type: ""},
		"Age":      {Min: 0},
		"Active":   {Type: true},
		"Role":     {Type: ""},
	}, schema.WithCollection(collName), schema.WithTimestamps(true))

	userModel := model.NewGeneric[testutil.TestUser]("MultiFieldUser", userSchema, client.Database)

	// Create a test user
	user := &testutil.TestUser{
		Username: "multiuser",
		Email:    "multi@example.com",
		Age:      28,
		Active:   true,
		Role:     "user",
	}

	err := userModel.Create(ctx, user)
	testutil.AssertNoError(t, err, "Failed to create user")

	// Update multiple fields (email and age)
	err = userModel.UpdateById(ctx, user.ID.Hex(), bson.M{
		"email": "multi_updated@example.com",
		"age":   35,
	})
	testutil.AssertNoError(t, err, "Failed to update multiple fields")

	// Find user and verify fields
	updated, err := userModel.FindById(ctx, user.ID.Hex())
	testutil.AssertNoError(t, err, "Failed to find user")

	// Verify updated fields
	testutil.AssertEqual(t, "multi_updated@example.com", updated.Email, "Email should be updated")
	testutil.AssertEqual(t, 35, updated.Age, "Age should be updated")

	// Other fields should be preserved
	testutil.AssertEqual(t, "multiuser", updated.Username, "Username should be preserved")
	testutil.AssertEqual(t, true, updated.Active, "Active should be preserved")
	testutil.AssertEqual(t, "user", updated.Role, "Role should be preserved")
}

func TestQueryUpdate(t *testing.T) {
	ctx := context.Background()
	client, cleanup := testutil.CreateTestClient(t)
	defer cleanup()

	// Use a separate collection for this test and clean it first
	collName := "query_update"
	testutil.DropCollection(t, client.Database, collName)

	// Create schema and model
	userSchema := schema.New(map[string]schema.Field{
		"Username": {Required: true, Type: ""},
		"Email":    {Required: true, Type: ""},
		"Age":      {Min: 0},
		"Active":   {Type: true},
		"Role":     {Type: ""},
	}, schema.WithCollection(collName), schema.WithTimestamps(true))

	userModel := model.NewGeneric[testutil.TestUser]("QueryUser", userSchema, client.Database)

	// Create two users - one admin and one regular user
	admin := &testutil.TestUser{
		Username: "admin",
		Email:    "admin@example.com",
		Age:      40,
		Active:   true,
		Role:     "admin",
	}

	regular := &testutil.TestUser{
		Username: "regular",
		Email:    "regular@example.com",
		Age:      25,
		Active:   true,
		Role:     "user",
	}

	err := userModel.Create(ctx, admin)
	testutil.AssertNoError(t, err, "Failed to create admin")

	err = userModel.Create(ctx, regular)
	testutil.AssertNoError(t, err, "Failed to create regular user")

	// Use QueryBuilder to update users with admin role
	queryBuilder := merhongo.QueryNew().Where("role", "admin")
	updateCount, err := userModel.UpdateWithQuery(
		ctx,
		queryBuilder,
		bson.M{"email": "admin_new@example.com"},
	)
	testutil.AssertNoError(t, err, "Failed to update with query")
	testutil.AssertEqual(t, int64(1), updateCount, "Should update exactly 1 document")

	// Find admin user and verify fields
	updatedAdmin, err := userModel.FindById(ctx, admin.ID.Hex())
	testutil.AssertNoError(t, err, "Failed to find admin")

	// Admin's email should be updated, other fields preserved
	testutil.AssertEqual(t, "admin_new@example.com", updatedAdmin.Email, "Admin email should be updated")
	testutil.AssertEqual(t, "admin", updatedAdmin.Username, "Admin username should be preserved")
	testutil.AssertEqual(t, 40, updatedAdmin.Age, "Admin age should be preserved")
	testutil.AssertEqual(t, "admin", updatedAdmin.Role, "Admin role should be preserved")

	// Regular user's fields should remain unchanged
	regularAfter, err := userModel.FindById(ctx, regular.ID.Hex())
	testutil.AssertNoError(t, err, "Failed to find regular user")
	testutil.AssertEqual(t, "regular@example.com", regularAfter.Email, "Regular user email should not change")
}

func TestEmptyValueUpdate(t *testing.T) {
	ctx := context.Background()
	client, cleanup := testutil.CreateTestClient(t)
	defer cleanup()

	// Use a separate collection for this test and clean it first
	collName := "empty_value_update"
	testutil.DropCollection(t, client.Database, collName)

	// Create schema and model
	userSchema := schema.New(map[string]schema.Field{
		"Username": {Required: true, Type: ""},
		"Email":    {Required: true, Type: ""},
		"Role":     {Type: ""}, // Not required
	}, schema.WithCollection(collName))

	userModel := model.NewGeneric[testutil.TestUser]("EmptyValueUser", userSchema, client.Database)

	// Create a test user
	user := &testutil.TestUser{
		Username: "emptytest",
		Email:    "empty@example.com",
		Role:     "manager",
	}

	err := userModel.Create(ctx, user)
	testutil.AssertNoError(t, err, "Failed to create user")

	// Update role field to empty string
	err = userModel.UpdateById(ctx, user.ID.Hex(), bson.M{"role": ""})
	testutil.AssertNoError(t, err, "Failed to update role to empty string")

	// Find user and verify fields
	updated, err := userModel.FindById(ctx, user.ID.Hex())
	testutil.AssertNoError(t, err, "Failed to find user")

	// Role should be empty, other fields preserved
	testutil.AssertEqual(t, "", updated.Role, "Role should be empty")
	testutil.AssertEqual(t, "emptytest", updated.Username, "Username should be preserved")
	testutil.AssertEqual(t, "empty@example.com", updated.Email, "Email should be preserved")
}
