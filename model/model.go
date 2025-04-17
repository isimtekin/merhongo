// Package model provides MongoDB document model operations
package model

import (
	"context"
	"github.com/isimtekin/merhongo/errors"
	"github.com/isimtekin/merhongo/schema"
	"log"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Model represents a MongoDB collection with its operations
type Model struct {
	Name       string
	Schema     *schema.Schema
	Collection *mongo.Collection
	DB         *mongo.Database
}

// New creates a new model for the given collection
func New(name string, schema *schema.Schema, db *mongo.Database) *Model {
	collName := schema.Collection
	if collName == "" {
		collName = name
	}

	var collection *mongo.Collection
	if db != nil {
		collection = db.Collection(collName)
	}

	model := &Model{
		Name:       name,
		Schema:     schema,
		Collection: collection,
		DB:         db,
	}

	// Only create indexes if db/collection is initialized
	if model.Collection != nil {
		for fieldName, field := range schema.Fields {
			if field.Unique {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				indexModel := mongo.IndexModel{
					Keys:    bson.D{{Key: fieldName, Value: 1}},
					Options: options.Index().SetUnique(true),
				}
				_, err := model.Collection.Indexes().CreateOne(ctx, indexModel)
				cancel()
				if err != nil {
					log.Printf("⚠️ Failed to create index for field '%s': %v", fieldName, err)
				} else {
					log.Printf("✅ Created unique index for field '%s'", fieldName)
				}
			}
		}
	}

	return model
}

// addTimestamps adds timestamps to the document
func (m *Model) addTimestamps(doc interface{}, isNew bool) {
	if !m.Schema.Timestamps {
		return
	}

	val := reflect.ValueOf(doc)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	now := time.Now()

	// Set CreatedAt for new documents
	createdField := val.FieldByName("CreatedAt")
	if createdField.IsValid() && createdField.CanSet() && isNew {
		createdField.Set(reflect.ValueOf(now))
	}

	// Always set UpdatedAt
	updatedField := val.FieldByName("UpdatedAt")
	if updatedField.IsValid() && updatedField.CanSet() {
		updatedField.Set(reflect.ValueOf(now))
	}
}

// applyMiddlewares applies middleware functions to a document
func (m *Model) applyMiddlewares(event string, doc interface{}) error {
	middlewares := m.Schema.Middlewares[event]
	for _, middleware := range middlewares {
		if err := middleware(doc); err != nil {
			return errors.Wrap(errors.ErrMiddleware, err.Error())
		}
	}
	return nil
}

// Create inserts a new document into the collection
func (m *Model) Create(ctx context.Context, doc interface{}) error {
	// Apply pre-save middlewares
	if err := m.applyMiddlewares("save", doc); err != nil {
		return err
	}

	// Validate document against schema
	if err := m.Schema.ValidateDocument(doc); err != nil {
		return errors.Wrap(errors.ErrValidation, err.Error())
	}

	// Add timestamps
	m.addTimestamps(doc, true)

	// Insert document and set ID back to struct
	result, err := m.Collection.InsertOne(ctx, doc)
	if err != nil {
		return errors.Wrap(errors.ErrDatabase, "failed to create document")
	}

	// Set ID back to the struct, if field is named ID and is settable
	val := reflect.ValueOf(doc).Elem()
	idField := val.FieldByName("ID")
	if idField.IsValid() && idField.CanSet() {
		idField.Set(reflect.ValueOf(result.InsertedID))
	}

	return nil
}

// FindById finds a document by its ID
func (m *Model) FindById(ctx context.Context, id string, result interface{}) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.WithDetails(errors.ErrInvalidObjectID, err.Error())
	}

	filter := bson.M{"_id": objectID}
	err = m.Collection.FindOne(ctx, filter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.WrapWithID(errors.ErrNotFound, "document not found", id)
		}
		return errors.Wrap(errors.ErrDatabase, "failed to retrieve document")
	}

	return nil
}

// Find finds documents matching the filter
func (m *Model) Find(ctx context.Context, filter interface{}, results interface{}) error {
	if m.Collection == nil {
		return errors.ErrNilCollection
	}

	cursor, err := m.Collection.Find(ctx, filter)
	if err != nil {
		return errors.Wrap(errors.ErrDatabase, "failed to retrieve documents")
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			// Log the error or handle it appropriately
			log.Printf("Failed to close cursor: %v", err)
		}
	}()

	err = cursor.All(ctx, results)
	if err != nil {
		return errors.Wrap(errors.ErrDecoding, err.Error())
	}

	return nil
}

// FindOne finds a single document matching the filter
func (m *Model) FindOne(ctx context.Context, filter interface{}, result interface{}) error {
	err := m.Collection.FindOne(ctx, filter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.ErrNotFound
		}
		return errors.Wrap(errors.ErrDatabase, "failed to retrieve document")
	}

	return nil
}

// UpdateById updates a document by its ID
func (m *Model) UpdateById(ctx context.Context, id string, update interface{}) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.WithDetails(errors.ErrInvalidObjectID, err.Error())
	}

	filter := bson.M{"_id": objectID}
	result, err := m.Collection.UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		return errors.Wrap(errors.ErrDatabase, "failed to update document")
	}

	if result.MatchedCount == 0 {
		return errors.WrapWithID(errors.ErrNotFound, "document not found", id)
	}

	return nil
}

// DeleteById deletes a document by its ID
func (m *Model) DeleteById(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.WithDetails(errors.ErrInvalidObjectID, err.Error())
	}

	filter := bson.M{"_id": objectID}
	result, err := m.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return errors.Wrap(errors.ErrDatabase, "failed to delete document")
	}

	if result.DeletedCount == 0 {
		return errors.WrapWithID(errors.ErrNotFound, "document not found", id)
	}

	return nil
}

// Count returns the number of documents matching the filter
func (m *Model) Count(ctx context.Context, filter interface{}) (int64, error) {
	if m.Collection == nil {
		return 0, errors.ErrNilCollection
	}

	count, err := m.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, errors.Wrap(errors.ErrDatabase, "failed to count documents")
	}

	return count, nil
}
