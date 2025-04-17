package model

import (
	"context"
	"github.com/isimtekin/merhongo/errors"
	"github.com/isimtekin/merhongo/query"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

// FindWithQuery finds documents using a query builder
func (m *Model) FindWithQuery(ctx context.Context, queryBuilder *query.Builder, results interface{}) error {
	if m.Collection == nil {
		return errors.ErrNilCollection
	}

	// Get filter and options from the query builder
	filter, options, err := queryBuilder.Build()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}

	// Execute the query
	cursor, err := m.Collection.Find(ctx, filter, options)
	if err != nil {
		return errors.Wrap(errors.ErrDatabase, "failed to retrieve documents")
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			// Log the error or handle it appropriately
			log.Printf("Failed to close cursor: %v", err)
		}
	}()

	// Decode the results
	err = cursor.All(ctx, results)
	if err != nil {
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
			return errors.ErrNotFound
		}
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
		return 0, errors.Wrap(err, "failed to build query")
	}

	// Execute the count
	count, err := m.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, errors.Wrap(errors.ErrDatabase, "failed to count documents")
	}

	return count, nil
}

// UpdateWithQuery updates documents using a query builder
func (m *Model) UpdateWithQuery(ctx context.Context, queryBuilder *query.Builder, update interface{}) (int64, error) {
	if m.Collection == nil {
		return 0, errors.ErrNilCollection
	}

	// Get filter from the query builder
	filter, _, err := queryBuilder.Build()
	if err != nil {
		return 0, errors.Wrap(err, "failed to build query")
	}

	// Prepare update document
	updateDoc := update
	if _, ok := update.(map[string]interface{}); ok {
		// Wrap with $set if it's a simple map
		updateDoc = map[string]interface{}{"$set": update}
	}

	// Execute the update
	result, err := m.Collection.UpdateMany(ctx, filter, updateDoc)
	if err != nil {
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
		return 0, errors.Wrap(err, "failed to build query")
	}

	// Execute to delete
	result, err := m.Collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, errors.Wrap(errors.ErrDatabase, "failed to delete documents")
	}

	return result.DeletedCount, nil
}
