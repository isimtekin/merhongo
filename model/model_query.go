package model

import (
	"context"
	"github.com/isimtekin/merhongo/errors"
	"github.com/isimtekin/merhongo/query"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"reflect"
)

// FindWithQuery finds documents using a query builder
func (m *Model) FindWithQuery(ctx context.Context, queryBuilder *query.Builder, results interface{}) error {
	if m.Collection == nil {
		return errors.ErrNilCollection
	}

	// Get filter and options from the query builder
	filter, options, err := queryBuilder.Build()
	if err != nil {
		log.Printf("⚠️ Failed to build query: %v", err)
		return errors.Wrap(err, "failed to build query")
	}

	// Execute the query
	cursor, err := m.Collection.Find(ctx, filter, options)
	if err != nil {
		log.Printf("⚠️ Failed to retrieve documents with query: %v", err)
		return errors.Wrap(errors.ErrDatabase, "failed to retrieve documents")
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("⚠️ Failed to close cursor: %v", err)
		}
	}()

	// Decode the results
	err = cursor.All(ctx, results)
	if err != nil {
		log.Printf("⚠️ Failed to decode documents: %v", err)
		return errors.Wrap(errors.ErrDecoding, err.Error())
	}

	return nil
}

// FindOneWithQuery finds a single document using a query builder
func (m *Model) FindOneWithQuery(ctx context.Context, queryBuilder *query.Builder, result interface{}) error {

	if m.Collection == nil {
		return errors.ErrNilCollection
	}

	// Get filter and options from the query builder
	filter, findOptions, err := queryBuilder.Build()
	if err != nil {
		log.Printf("⚠️ Failed to build query: %v", err)
		return errors.Wrap(err, "failed to build query")
	}

	// Create FindOneOptions from the parts we need
	findOneOpts := options.FindOne()

	// Copy relevant options from FindOptions to FindOneOptions
	if findOptions.Sort != nil {
		findOneOpts.SetSort(findOptions.Sort)
	}

	if findOptions.Skip != nil {
		findOneOpts.SetSkip(*findOptions.Skip)
	}

	if findOptions.Projection != nil {
		findOneOpts.SetProjection(findOptions.Projection)
	}

	if findOptions.Collation != nil {
		findOneOpts.SetCollation(findOptions.Collation)
	}

	if findOptions.Comment != nil {
		findOneOpts.SetComment(*findOptions.Comment)
	}

	if findOptions.Hint != nil {
		findOneOpts.SetHint(findOptions.Hint)
	}

	if findOptions.Max != nil {
		findOneOpts.SetMax(findOptions.Max)
	}

	if findOptions.Min != nil {
		findOneOpts.SetMin(findOptions.Min)
	}

	// Execute the query
	err = m.Collection.FindOne(ctx, filter, findOneOpts).Decode(result)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			log.Printf("⚠️ No documents found with query: %v", filter)
			return errors.ErrNotFound
		}
		log.Printf("⚠️ Failed to retrieve document with query: %v", err)
		return errors.Wrap(errors.ErrDatabase, "failed to retrieve document")
	}

	return nil
}

// CountWithQuery counts documents using a query builder
func (m *Model) CountWithQuery(ctx context.Context, queryBuilder *query.Builder) (int64, error) {
	if m.Collection == nil {
		return 0, errors.ErrNilCollection
	}

	// Get filter from the query builder
	filter, _, err := queryBuilder.Build()
	if err != nil {
		log.Printf("⚠️ Failed to build query: %v", err)
		return 0, errors.Wrap(err, "failed to build query")
	}

	// Execute the count
	count, err := m.Collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("⚠️ Failed to count documents: %v", err)
		return 0, errors.Wrap(errors.ErrDatabase, "failed to count documents")
	}

	return count, nil
}

// UpdateWithQuery updates documents using a query builder with validation and timestamp handling
func (m *Model) UpdateWithQuery(ctx context.Context, queryBuilder *query.Builder, update interface{}) (int64, error) {
	if m.Collection == nil {
		return 0, errors.ErrNilCollection
	}

	// Get filter from the query builder
	filter, _, err := queryBuilder.Build()
	if err != nil {
		log.Printf("⚠️ Failed to build query: %v", err)
		return 0, errors.Wrap(err, "failed to build query")
	}

	// Prepare update document with timestamp handling
	finalUpdate, err := m.prepareUpdate(update)
	if err != nil {
		log.Printf("⚠️ Failed to prepare update: %v", err)
		return 0, err
	}

	// Validate affected documents if schema and model type are available
	if m.Schema != nil && m.Schema.ModelType != nil {
		// Find documents that will be affected
		cursor, err := m.Collection.Find(ctx, filter)
		if err != nil {
			log.Printf("⚠️ Failed to retrieve documents for validation: %v", err)
			return 0, errors.Wrap(errors.ErrDatabase, "failed to retrieve documents for validation")
		}
		defer cursor.Close(ctx)

		// Validate each document
		for cursor.Next(ctx) {
			var existingDoc bson.M
			if err := cursor.Decode(&existingDoc); err != nil {
				log.Printf("⚠️ Failed to decode document for validation: %v", err)
				return 0, errors.Wrap(errors.ErrDecoding, "failed to decode document")
			}

			// Apply update data to the existing document
			for key, value := range finalUpdate {
				existingDoc[key] = value
			}

			// Create a new instance of the model type
			docType := reflect.TypeOf(m.Schema.ModelType).Elem()
			newInstance := reflect.New(docType).Interface()

			// Convert existingDoc to struct
			bytes, _ := bson.Marshal(existingDoc)
			if err := bson.Unmarshal(bytes, newInstance); err != nil {
				log.Printf("⚠️ Failed to convert to struct for validation: %v", err)
				return 0, errors.Wrap(errors.ErrDecoding, "failed to convert to struct for validation")
			}

			// Validate the full document
			if err := m.Schema.ValidateDocument(newInstance); err != nil {
				log.Printf("⚠️ Document validation failed: %v", err)
				return 0, err
			}
		}

		if err := cursor.Err(); err != nil {
			log.Printf("⚠️ Error during cursor iteration: %v", err)
			return 0, errors.Wrap(errors.ErrDatabase, "error during cursor iteration")
		}
	}

	// Apply the update with the validated data
	updateDoc := map[string]interface{}{"$set": finalUpdate}
	result, err := m.Collection.UpdateMany(ctx, filter, updateDoc)
	if err != nil {
		log.Printf("⚠️ Failed to update documents with query: %v", err)
		return 0, errors.Wrap(errors.ErrDatabase, "failed to update documents")
	}

	return result.ModifiedCount, nil
}

// DeleteWithQuery deletes documents using a query builder
func (m *Model) DeleteWithQuery(ctx context.Context, queryBuilder *query.Builder) (int64, error) {
	if m.Collection == nil {
		return 0, errors.ErrNilCollection
	}

	// Get filter from the query builder
	filter, _, err := queryBuilder.Build()
	if err != nil {
		log.Printf("⚠️ Failed to build query: %v", err)
		return 0, errors.Wrap(err, "failed to build query")
	}

	// Execute to delete
	result, err := m.Collection.DeleteMany(ctx, filter)
	if err != nil {
		log.Printf("⚠️ Failed to delete documents with query: %v", err)
		return 0, errors.Wrap(errors.ErrDatabase, "failed to delete documents")
	}

	return result.DeletedCount, nil
}

// FindWithQuery finds documents using a query builder with type safety
func (m *GenericModel[T]) FindWithQuery(ctx context.Context, queryBuilder *query.Builder) ([]T, error) {
	var results []T
	err := m.Model.FindWithQuery(ctx, queryBuilder, &results)
	if err != nil {
		log.Printf("⚠️ Failed to retrieve documents with query: %v", err)
		return nil, err
	}
	return results, nil
}

// FindOneWithQuery finds a single document using a query builder with type safety
func (m *GenericModel[T]) FindOneWithQuery(ctx context.Context, queryBuilder *query.Builder) (*T, error) {
	result := new(T)
	err := m.Model.FindOneWithQuery(ctx, queryBuilder, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CountWithQuery counts documents using a query builder with type safety
func (m *GenericModel[T]) CountWithQuery(ctx context.Context, queryBuilder *query.Builder) (int64, error) {
	return m.Model.CountWithQuery(ctx, queryBuilder)
}

// UpdateWithQuery updates documents using a query builder with type safety
func (m *GenericModel[T]) UpdateWithQuery(ctx context.Context, queryBuilder *query.Builder, update interface{}) (int64, error) {
	return m.Model.UpdateWithQuery(ctx, queryBuilder, update)
}

// DeleteWithQuery deletes documents using a query builder with type safety
func (m *GenericModel[T]) DeleteWithQuery(ctx context.Context, queryBuilder *query.Builder) (int64, error) {
	return m.Model.DeleteWithQuery(ctx, queryBuilder)
}
