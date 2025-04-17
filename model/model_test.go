package model

import (
	"context"
	"fmt"
	"github.com/isimtekin/merhongo/connection"
	"github.com/isimtekin/merhongo/errors"
	"github.com/isimtekin/merhongo/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"strings"
	"testing"
	"time"
)

type User struct {
	ID        interface{} `bson:"_id,omitempty"`
	Username  string      `bson:"username"`
	Email     string      `bson:"email"`
	Age       int         `bson:"age"`
	IsActive  bool        `bson:"isActive"`
	CreatedAt time.Time   `bson:"createdAt"`
	UpdatedAt time.Time   `bson:"updatedAt"`
}

func TestUserModel_CreateAndFindOne(t *testing.T) {
	ctx := context.Background()

	client, err := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	if err != nil {
		t.Fatalf("DB connection failed: %v", err)
	}
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Printf("Failed to disconnect: %v", err)
		}
	}()

	// Clean up before test
	_ = client.Database.Collection("users").Drop(ctx)

	userSchema := schema.New(
		map[string]schema.Field{
			"Username": {Required: true, Unique: true},
			"Email":    {Required: true, Unique: true},
			"Age":      {Min: 18, Max: 100},
		},
		schema.WithCollection("users"),
	)

	userModel := New("User", userSchema, client.Database)

	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Age:      27,
	}

	err = userModel.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var result User
	err = userModel.FindOne(ctx, map[string]interface{}{"username": "testuser"}, &result)
	if err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}

	if result.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, result.Email)
	}
}

func TestModel_CRUDOperations(t *testing.T) {
	ctx := context.Background()
	client, _ := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Printf("Failed to disconnect: %v", err)
		}
	}()

	type User struct {
		ID        interface{} `bson:"_id,omitempty"`
		Username  string      `bson:"username"`
		Age       int         `bson:"age"`
		CreatedAt time.Time   `bson:"createdAt"`
		UpdatedAt time.Time   `bson:"updatedAt"`
	}

	userSchema := schema.New(map[string]schema.Field{
		"Username": {Required: true, Unique: true},
		"Age":      {Min: 0},
	}, schema.WithCollection("users"))

	userModel := New("User", userSchema, client.Database)
	_ = userModel.Collection.Drop(ctx)

	user := &User{Username: "john_doe", Age: 28}
	err := userModel.Create(ctx, user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Test Count
	count, err := userModel.Count(ctx, bson.M{"username": "john_doe"})
	if err != nil || count != 1 {
		t.Errorf("expected 1 user, got %d", count)
	}

	// Test FindById
	id := user.ID.(primitive.ObjectID).Hex()
	var found User
	err = userModel.FindById(ctx, id, &found)
	if err != nil || found.Username != "john_doe" {
		t.Errorf("expected to find user by ID")
	}

	// Test UpdateById
	err = userModel.UpdateById(ctx, id, bson.M{"age": 30})
	if err != nil {
		t.Errorf("failed to update user: %v", err)
	}

	// Test DeleteById
	err = userModel.DeleteById(ctx, id)
	if err != nil {
		t.Errorf("failed to delete user: %v", err)
	}
}

// This is the failing test that needs to be fixed
func TestModel_Create_WithMiddlewareError(t *testing.T) {
	ctx := context.Background()
	client, _ := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Printf("Failed to disconnect: %v", err)
		}
	}()

	type Dummy struct {
		ID   interface{} `bson:"_id,omitempty"`
		Name string      `bson:"name"`
	}

	s := schema.New(map[string]schema.Field{
		"Name": {Required: true},
	}, schema.WithCollection("dummies"))

	// Add middleware that always fails
	s.Pre("save", func(doc interface{}) error {
		return fmt.Errorf("middleware failed")
	})

	m := New("Dummy", s, client.Database)

	doc := &Dummy{Name: "test"}
	err := m.Create(ctx, doc)

	// Use the helper function instead of checking the exact error message
	if !errors.IsMiddlewareError(err) {
		t.Errorf("expected middleware failure, got %v", err)
	}

	// If you want to check that the details contain the original error message
	if !strings.Contains(err.Error(), "middleware failed") {
		t.Errorf("expected error to contain 'middleware failed', got: %v", err)
	}
}

// Test for validation error
func TestModel_Create_WithValidationError(t *testing.T) {
	ctx := context.Background()
	client, _ := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Printf("Failed to disconnect: %v", err)
		}
	}()

	type Dummy struct {
		ID   interface{} `bson:"_id,omitempty"`
		Name string      `bson:"name"`
	}

	s := schema.New(map[string]schema.Field{
		"Name": {Required: true},
	}, schema.WithCollection("dummies"))

	// Override validation temporarily
	s.CustomValidator = func(doc interface{}) error {
		return errors.ErrValidation
	}

	m := New("Dummy", s, client.Database)

	doc := &Dummy{Name: "fail"}
	err := m.Create(ctx, doc)

	// Use the helper function instead of errors.Is
	if !errors.IsValidationError(err) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

// Test for nil collection error
func TestModel_Find_InvalidCollection(t *testing.T) {
	ctx := context.Background()

	type Dummy struct {
		ID string `bson:"_id,omitempty"`
	}

	s := schema.New(map[string]schema.Field{}, schema.WithCollection("invalid"))
	m := New("Dummy", s, nil) // nil client.Database

	var results []Dummy
	err := m.Find(ctx, bson.M{}, &results)

	// Use the helper function instead of errors.Is
	if !errors.IsNilCollectionError(err) {
		t.Errorf("expected ErrNilCollection, got %v", err)
	}
}

// Test for the not found error
func TestModel_FindOne_NotFound(t *testing.T) {
	ctx := context.Background()
	client, _ := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Printf("Failed to disconnect: %v", err)
		}
	}()

	type Dummy struct {
		ID string `bson:"_id,omitempty"`
	}

	s := schema.New(map[string]schema.Field{}, schema.WithCollection("dummies"))
	m := New("Dummy", s, client.Database)

	var result Dummy
	err := m.FindOne(ctx, bson.M{"username": "not_exists"}, &result)

	// Use the helper function instead of errors.Is
	if !errors.IsNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestModel_UpdateById_InvalidObjectID(t *testing.T) {
	ctx := context.Background()
	client, _ := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Printf("Failed to disconnect: %v", err)
		}
	}()

	s := schema.New(map[string]schema.Field{}, schema.WithCollection("dummies"))
	m := New("Dummy", s, client.Database)

	err := m.UpdateById(ctx, "invalid_object_id", bson.M{"name": "updated"})

	if err == nil || !strings.Contains(err.Error(), "invalid ObjectID") {
		t.Errorf("expected ObjectID parse error, got %v", err)
	}
}

func TestModel_UpdateById_DocumentNotFound(t *testing.T) {
	ctx := context.Background()
	client, _ := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Printf("Failed to disconnect: %v", err)
		}
	}()

	s := schema.New(map[string]schema.Field{}, schema.WithCollection("dummies"))
	m := New("Dummy", s, client.Database)

	// Use a valid but non-existent ObjectID
	id := primitive.NewObjectID().Hex()

	err := m.UpdateById(ctx, id, bson.M{"name": "new"})

	if err == nil || !strings.Contains(err.Error(), "document not found") {
		t.Errorf("expected document not found error, got %v", err)
	}
}

// Test for invalid ObjectID error
func TestModel_FindById_InvalidObjectID(t *testing.T) {
	ctx := context.Background()
	client, _ := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Printf("Failed to disconnect: %v", err)
		}
	}()

	s := schema.New(map[string]schema.Field{}, schema.WithCollection("dummies"))
	m := New("Dummy", s, client.Database)

	var result map[string]interface{}
	err := m.FindById(ctx, "not-an-objectid", &result)

	// Use the helper function instead of errors.Is
	if !errors.IsInvalidObjectID(err) {
		t.Errorf("expected ErrInvalidObjectID, got: %v", err)
	}
}

func TestModel_FindById_DocumentNotFound(t *testing.T) {
	ctx := context.Background()
	client, _ := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Printf("Failed to disconnect: %v", err)
		}
	}()

	s := schema.New(map[string]schema.Field{}, schema.WithCollection("dummies"))
	m := New("Dummy", s, client.Database)

	var result map[string]interface{}
	id := primitive.NewObjectID().Hex() // var ama koleksiyonda yok
	err := m.FindById(ctx, id, &result)

	if err == nil || !strings.Contains(err.Error(), "document not found") {
		t.Errorf("expected document not found error, got: %v", err)
	}
}

func TestModel_Find_DecodeError(t *testing.T) {
	ctx := context.Background()
	client, _ := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Printf("Failed to disconnect: %v", err)
		}
	}()

	s := schema.New(map[string]schema.Field{}, schema.WithCollection("dummies"))
	m := New("Dummy", s, client.Database)

	// Geçerli bir belge ekle
	_ = m.Collection.Drop(ctx)
	_, _ = m.Collection.InsertOne(ctx, bson.M{"username": "john"})

	// Geçersiz results: slice olmayan veri → decode error tetikler
	var invalidTarget map[string]interface{}
	err := m.Find(ctx, bson.M{}, &invalidTarget)

	if err == nil || !strings.Contains(err.Error(), "failed to decode documents") {
		t.Errorf("expected decode error, got %v", err)
	}
}

func TestModel_DeleteById_InvalidObjectID(t *testing.T) {
	ctx := context.Background()
	client, _ := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Printf("Failed to disconnect: %v", err)
		}
	}()

	s := schema.New(map[string]schema.Field{}, schema.WithCollection("dummies"))
	m := New("Dummy", s, client.Database)

	err := m.DeleteById(ctx, "invalid_object_id")

	if err == nil || !strings.Contains(err.Error(), "invalid ObjectID") {
		t.Errorf("expected ObjectID parse error, got %v", err)
	}
}

func TestModel_DeleteById_DocumentNotFound(t *testing.T) {
	ctx := context.Background()
	client, _ := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Printf("Failed to disconnect: %v", err)
		}
	}()

	s := schema.New(map[string]schema.Field{}, schema.WithCollection("dummies"))
	m := New("Dummy", s, client.Database)

	// Var olmayan ama geçerli bir ObjectID kullan
	id := primitive.NewObjectID().Hex()
	err := m.DeleteById(ctx, id)

	if err == nil || !strings.Contains(err.Error(), "document not found") {
		t.Errorf("expected 'document not found' error, got %v", err)
	}
}

func TestModel_Count_Success(t *testing.T) {
	ctx := context.Background()
	client, _ := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Printf("Failed to disconnect: %v", err)
		}
	}()

	s := schema.New(map[string]schema.Field{
		"Name": {Required: true},
	}, schema.WithCollection("dummies"))

	m := New("Dummy", s, client.Database)

	_ = m.Collection.Drop(ctx) // temizle
	_, _ = m.Collection.InsertOne(ctx, bson.M{"name": "test"})
	_, _ = m.Collection.InsertOne(ctx, bson.M{"name": "test"})

	count, err := m.Count(ctx, bson.M{"name": "test"})
	if err != nil {
		t.Errorf("Count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestModel_Count_Error(t *testing.T) {
	s := schema.New(map[string]schema.Field{}, schema.WithCollection("dummies"))
	m := New("Dummy", s, nil)

	_, err := m.Count(context.Background(), bson.M{})
	if err == nil || !strings.Contains(err.Error(), "collection is nil") {
		t.Errorf("expected error from nil collection, got %v", err)
	}
}

func TestNewModel_NoDB(t *testing.T) {
	schema := schema.New(map[string]schema.Field{
		"Name": {Unique: true},
	})

	model := New("testNoDB", schema, nil)
	if model.Collection != nil {
		t.Errorf("expected Collection to be nil")
	}
}

func TestNewModel_NoUniqueIndex(t *testing.T) {
	schema := schema.New(map[string]schema.Field{
		"Name": {Required: true},
	})

	client, _ := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	model := New("test_no_unique", schema, client.Database)

	if model == nil || model.Collection == nil {
		t.Fatal("model or collection is nil")
	}
}

// Burada sadece index oluşturuldu mu diye genel test yapılır
func TestNewModel_UniqueIndexSuccess(t *testing.T) {
	schema := schema.New(map[string]schema.Field{
		"Email": {Unique: true},
	})

	client, _ := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	model := New("test_index_success", schema, client.Database)

	// check if index exists could be added with listIndexes
	if model.Collection == nil {
		t.Error("expected collection to be initialized")
	}
}

type TimestampedDoc struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt time.Time          `bson:"createdAt,omitempty"`
	UpdatedAt time.Time          `bson:"updatedAt,omitempty"`
}

func TestModel_AddTimestamps_Disabled(t *testing.T) {
	doc := &TimestampedDoc{}
	model := &Model{
		Schema: &schema.Schema{Timestamps: false},
	}
	model.addTimestamps(doc, true)

	if !doc.CreatedAt.IsZero() || !doc.UpdatedAt.IsZero() {
		t.Error("expected no timestamps to be set when disabled")
	}
}

func TestModel_AddTimestamps_NewDoc(t *testing.T) {
	doc := &TimestampedDoc{}
	model := &Model{
		Schema: &schema.Schema{Timestamps: true},
	}
	model.addTimestamps(doc, true)

	if doc.CreatedAt.IsZero() || doc.UpdatedAt.IsZero() {
		t.Error("expected CreatedAt and UpdatedAt to be set for new doc")
	}
}

func TestModel_AddTimestamps_ExistingDoc(t *testing.T) {
	doc := &TimestampedDoc{}
	model := &Model{
		Schema: &schema.Schema{Timestamps: true},
	}
	model.addTimestamps(doc, false)

	if !doc.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to remain zero for existing doc")
	}
	if doc.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}
