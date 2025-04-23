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

// GenericModel extends Model with type-safe operations for a specific document type
type GenericModel[T any] struct {
	*Model
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
			if field.Index || field.Unique {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				indexOptions := options.Index()
				if field.Unique {
					indexOptions.SetUnique(true)
				}

				indexModel := mongo.IndexModel{
					Keys:    bson.D{{Key: fieldName, Value: 1}},
					Options: indexOptions,
				}
				_, err := model.Collection.Indexes().CreateOne(ctx, indexModel)
				cancel()
				if err != nil {
					log.Printf("⚠️ Failed to create index for field '%s': %v", fieldName, err)
				} else {
					indexType := "index"
					if field.Unique {
						indexType = "unique index"
					}
					log.Printf("✅ Created %s for field '%s'", indexType, fieldName)
				}
			}
		}
	}

	return model
}

// NewGeneric creates a new generic model with type-safe operations
func NewGeneric[T any](name string, schema *schema.Schema, db *mongo.Database) *GenericModel[T] {
	// Set the model type in the schema for validation purposes
	var modelType T
	if schema != nil {
		schema.ModelType = &modelType
	}

	return &GenericModel[T]{
		Model: New(name, schema, db),
	}
}

// prepareUpdate prepares update data with timestamp handling
func (m *Model) prepareUpdate(update interface{}) (map[string]interface{}, error) {
	var finalUpdate map[string]interface{}

	// Convert update to map format
	switch u := update.(type) {
	case map[string]interface{}:
		finalUpdate = make(map[string]interface{})
		for k, v := range u {
			finalUpdate[k] = v
		}
	case bson.M:
		finalUpdate = make(map[string]interface{})
		for k, v := range u {
			finalUpdate[k] = v
		}
	case bson.D:
		finalUpdate = make(map[string]interface{})
		for _, e := range u {
			finalUpdate[e.Key] = e.Value
		}
	default:
		// Support for other types can be added using reflection
		// but for now let's support only common types
		return nil, errors.WithDetails(errors.ErrValidation, "update must be a map or bson type")
	}

	// Remove createdAt field (should not be modified)
	delete(finalUpdate, "createdAt")
	delete(finalUpdate, "CreatedAt")

	// Add/update updatedAt field if timestamps are enabled
	if m.Schema != nil && m.Schema.Timestamps {
		finalUpdate["updatedAt"] = time.Now()
	}

	return finalUpdate, nil
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
		log.Printf("⚠️ Failed to insert document: %v", err)
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

// Create inserts a new document with type safety
func (m *GenericModel[T]) Create(ctx context.Context, doc *T) error {
	return m.Model.Create(ctx, doc)
}

// FindById finds a document by its ID
func (m *Model) FindById(ctx context.Context, id string, result interface{}) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("⚠️ Invalid ObjectID format: %s - %v", id, err)
		return errors.WithDetails(errors.ErrInvalidObjectID, err.Error())
	}

	filter := bson.M{"_id": objectID}
	err = m.Collection.FindOne(ctx, filter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("⚠️ Document not found with ID: %s", id)
			return errors.WrapWithID(errors.ErrNotFound, "document not found", id)
		}
		log.Printf("⚠️ Failed to retrieve document with ID %s: %v", id, err)
		return errors.Wrap(errors.ErrDatabase, "failed to retrieve document")
	}

	return nil
}

// FindById finds a document by its ID with type safety
func (m *GenericModel[T]) FindById(ctx context.Context, id string) (*T, error) {
	result := new(T)
	err := m.Model.FindById(ctx, id, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Find finds documents matching the filter
func (m *Model) Find(ctx context.Context, filter interface{}, results interface{}) error {
	if m.Collection == nil {
		return errors.ErrNilCollection
	}

	cursor, err := m.Collection.Find(ctx, filter)
	if err != nil {
		log.Printf("⚠️ Failed to retrieve documents: %v", err)
		return errors.Wrap(errors.ErrDatabase, "failed to retrieve documents")
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("⚠️ Failed to close cursor: %v", err)
		}
	}()

	err = cursor.All(ctx, results)
	if err != nil {
		log.Printf("⚠️ Failed to decode documents: %v", err)
		return errors.Wrap(errors.ErrDecoding, err.Error())
	}

	return nil
}

// Find finds documents matching the filter with type safety
func (m *GenericModel[T]) Find(ctx context.Context, filter interface{}) ([]T, error) {
	var results []T
	err := m.Model.Find(ctx, filter, &results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// FindOne finds a single document matching the filter
func (m *Model) FindOne(ctx context.Context, filter interface{}, result interface{}) error {
	err := m.Collection.FindOne(ctx, filter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("⚠️ Document not found with filter: %v", filter)
			return errors.ErrNotFound
		}
		log.Printf("⚠️ Failed to retrieve document: %v", err)
		return errors.Wrap(errors.ErrDatabase, "failed to retrieve document")
	}

	return nil
}

// FindOne finds a single document matching the filter with type safety
func (m *GenericModel[T]) FindOne(ctx context.Context, filter interface{}) (*T, error) {
	result := new(T)
	err := m.Model.FindOne(ctx, filter, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateById updates a document by its ID with validation and timestamp handling
func (m *Model) UpdateById(ctx context.Context, id string, update interface{}) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("⚠️ Invalid ObjectID format: %s - %v", id, err)
		return errors.WithDetails(errors.ErrInvalidObjectID, err.Error())
	}

	// 1. First find the existing document
	filter := bson.M{"_id": objectID}
	result := m.Collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			log.Printf("⚠️ Document not found with ID: %s", id)
			return errors.WrapWithID(errors.ErrNotFound, "document not found", id)
		}
		log.Printf("⚠️ Failed to retrieve document with ID %s for update: %v", id, result.Err())
		return errors.Wrap(errors.ErrDatabase, "failed to retrieve document")
	}

	// 2. Load the existing document as a map
	var existingDoc bson.M
	if err := result.Decode(&existingDoc); err != nil {
		log.Printf("⚠️ Failed to decode document with ID %s: %v", id, err)
		return errors.Wrap(errors.ErrDecoding, "failed to decode document")
	}

	// 3. Prepare update data (handle timestamps)
	finalUpdate, err := m.prepareUpdate(update)
	if err != nil {
		return err
	}

	// 4. Apply update data to the existing document
	for key, value := range finalUpdate {
		existingDoc[key] = value
	}

	// 5. Validate the full updated document
	if m.Schema != nil && m.Schema.ModelType != nil {
		// Create a new instance of the model type
		docType := reflect.TypeOf(m.Schema.ModelType).Elem()
		newInstance := reflect.New(docType).Interface()

		// Convert existingDoc to struct
		bytes, _ := bson.Marshal(existingDoc)
		if err := bson.Unmarshal(bytes, newInstance); err != nil {
			log.Printf("⚠️ Failed to convert document to struct for validation: %v", err)
			return errors.Wrap(errors.ErrDecoding, "failed to convert to struct for validation")
		}

		// Validate the document
		if err := m.Schema.ValidateDocument(newInstance); err != nil {
			log.Printf("⚠️ Document validation failed: %v", err)
			return err
		}
	}

	// 6. Apply the update
	_, err = m.Collection.UpdateOne(ctx, filter, bson.M{"$set": finalUpdate})
	if err != nil {
		log.Printf("⚠️ Failed to update document with ID %s: %v", id, err)
		return errors.Wrap(errors.ErrDatabase, "failed to update document")
	}

	return nil
}

// UpdateById updates a document by its ID with type safety
func (m *GenericModel[T]) UpdateById(ctx context.Context, id string, update interface{}) error {
	return m.Model.UpdateById(ctx, id, update)
}

// DeleteById deletes a document by its ID
func (m *Model) DeleteById(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("⚠️ Invalid ObjectID format: %s - %v", id, err)
		return errors.WithDetails(errors.ErrInvalidObjectID, err.Error())
	}

	filter := bson.M{"_id": objectID}
	result, err := m.Collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Printf("⚠️ Failed to delete document with ID %s: %v", id, err)
		return errors.Wrap(errors.ErrDatabase, "failed to delete document")
	}

	if result.DeletedCount == 0 {
		log.Printf("⚠️ Document not found with ID: %s", id)
		return errors.WrapWithID(errors.ErrNotFound, "document not found", id)
	}

	return nil
}

// DeleteById deletes a document by its ID with type safety
func (m *GenericModel[T]) DeleteById(ctx context.Context, id string) error {
	return m.Model.DeleteById(ctx, id)
}

// Count returns the number of documents matching the filter
func (m *Model) Count(ctx context.Context, filter interface{}) (int64, error) {
	if m.Collection == nil {
		return 0, errors.ErrNilCollection
	}

	count, err := m.Collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("⚠️ Failed to count documents: %v", err)
		return 0, errors.Wrap(errors.ErrDatabase, "failed to count documents")
	}

	return count, nil
}

// Count returns the number of documents matching the filter with type safety
func (m *GenericModel[T]) Count(ctx context.Context, filter interface{}) (int64, error) {
	return m.Model.Count(ctx, filter)
}
