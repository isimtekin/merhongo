// Package query provides utilities for building MongoDB queries
package query

import (
	"github.com/isimtekin/merhongo/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Operator constants for MongoDB query operators
const (
	OpEqual        = "$eq"
	OpNotEqual     = "$ne"
	OpGreaterThan  = "$gt"
	OpGreaterEqual = "$gte"
	OpLessThan     = "$lt"
	OpLessEqual    = "$lte"
	OpIn           = "$in"
	OpNotIn        = "$nin"
	OpExists       = "$exists"
	OpRegex        = "$regex"
)

// Builder helps to build MongoDB queries
type Builder struct {
	filter bson.M
	sort   bson.D
	limit  int64
	skip   int64
	err    error
}

// New creates a new query builder
func New() *Builder {
	return &Builder{
		filter: bson.M{},
		sort:   bson.D{},
	}
}

// WithError creates a new query builder that starts with an error
// This is useful for chaining error handling
func WithError(err error) *Builder {
	builder := New()
	builder.err = err
	return builder
}

// Error returns any error that occurred during query building
func (b *Builder) Error() error {
	return b.err
}

// Where adds filter conditions
func (b *Builder) Where(key string, value interface{}) *Builder {
	if b.err != nil {
		return b
	}

	if key == "" {
		b.err = errors.WithDetails(errors.ErrValidation, "key cannot be empty")
		return b
	}

	b.filter[key] = value
	return b
}

// WhereOperator adds a filter with a specific operator
func (b *Builder) WhereOperator(key string, operator string, value interface{}) *Builder {
	if b.err != nil {
		return b
	}

	if key == "" {
		b.err = errors.WithDetails(errors.ErrValidation, "key cannot be empty")
		return b
	}

	if operator == "" {
		b.err = errors.WithDetails(errors.ErrValidation, "operator cannot be empty")
		return b
	}

	existing, exists := b.filter[key]
	if exists {
		if existingMap, ok := existing.(bson.M); ok {
			existingMap[operator] = value
			b.filter[key] = existingMap
		} else {
			// Convert direct value to operator map
			b.filter[key] = bson.M{operator: value}
		}
	} else {
		b.filter[key] = bson.M{operator: value}
	}

	return b
}

// Equals adds an equals condition
func (b *Builder) Equals(key string, value interface{}) *Builder {
	return b.WhereOperator(key, OpEqual, value)
}

// NotEquals adds a not equals condition
func (b *Builder) NotEquals(key string, value interface{}) *Builder {
	return b.WhereOperator(key, OpNotEqual, value)
}

// GreaterThan adds a greater than condition
func (b *Builder) GreaterThan(key string, value interface{}) *Builder {
	return b.WhereOperator(key, OpGreaterThan, value)
}

// GreaterThanOrEqual adds a greater than or equal condition
func (b *Builder) GreaterThanOrEqual(key string, value interface{}) *Builder {
	return b.WhereOperator(key, OpGreaterEqual, value)
}

// LessThan adds a less than condition
func (b *Builder) LessThan(key string, value interface{}) *Builder {
	return b.WhereOperator(key, OpLessThan, value)
}

// LessThanOrEqual adds a less than or equal condition
func (b *Builder) LessThanOrEqual(key string, value interface{}) *Builder {
	return b.WhereOperator(key, OpLessEqual, value)
}

// In adds an in condition
func (b *Builder) In(key string, values interface{}) *Builder {
	return b.WhereOperator(key, OpIn, values)
}

// NotIn adds a not in condition
func (b *Builder) NotIn(key string, values interface{}) *Builder {
	return b.WhereOperator(key, OpNotIn, values)
}

// Exists adds an exists condition
func (b *Builder) Exists(key string, exists bool) *Builder {
	return b.WhereOperator(key, OpExists, exists)
}

// Regex adds a regular expression condition
func (b *Builder) Regex(key string, pattern string, options ...string) *Builder {
	if b.err != nil {
		return b
	}

	if key == "" {
		b.err = errors.WithDetails(errors.ErrValidation, "key cannot be empty")
		return b
	}

	regexDoc := bson.M{OpRegex: pattern}

	// Add options if provided
	if len(options) > 0 && options[0] != "" {
		regexDoc["$options"] = options[0]
	}

	// Check if we already have conditions for this field
	existing, exists := b.filter[key]
	if exists {
		if existingMap, ok := existing.(bson.M); ok {
			// Merge with existing conditions
			for k, v := range regexDoc {
				existingMap[k] = v
			}
		} else {
			b.filter[key] = regexDoc
		}
	} else {
		b.filter[key] = regexDoc
	}

	return b
}

// SortBy adds sort criteria
func (b *Builder) SortBy(key string, ascending bool) *Builder {
	if b.err != nil {
		return b
	}

	if key == "" {
		b.err = errors.WithDetails(errors.ErrValidation, "sort key cannot be empty")
		return b
	}

	var value int
	if ascending {
		value = 1
	} else {
		value = -1
	}

	// Check if we already have a sort for this key
	for i, sort := range b.sort {
		if sort.Key == key {
			// Update existing sort
			b.sort[i].Value = value
			return b
		}
	}

	// Add new sort
	b.sort = append(b.sort, bson.E{Key: key, Value: value})
	return b
}

// Limit sets the maximum number of results
func (b *Builder) Limit(limit int64) *Builder {
	if b.err != nil {
		return b
	}

	if limit < 0 {
		b.err = errors.WithDetails(errors.ErrValidation, "limit cannot be negative")
		return b
	}

	b.limit = limit
	return b
}

// Skip sets the number of results to skip
func (b *Builder) Skip(skip int64) *Builder {
	if b.err != nil {
		return b
	}

	if skip < 0 {
		b.err = errors.WithDetails(errors.ErrValidation, "skip cannot be negative")
		return b
	}

	b.skip = skip
	return b
}

// GetFilter returns the filter
func (b *Builder) GetFilter() (bson.M, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.filter, nil
}

// GetOptions returns the query options
func (b *Builder) GetOptions() (*options.FindOptions, error) {
	if b.err != nil {
		return nil, b.err
	}

	opts := options.Find()

	if len(b.sort) > 0 {
		opts.SetSort(b.sort)
	}

	if b.limit > 0 {
		opts.SetLimit(b.limit)
	}

	if b.skip > 0 {
		opts.SetSkip(b.skip)
	}

	return opts, nil
}

// Build returns both the filter and options, or an error if one occurred
func (b *Builder) Build() (bson.M, *options.FindOptions, error) {
	if b.err != nil {
		return nil, nil, b.err
	}

	opts := options.Find()

	if len(b.sort) > 0 {
		opts.SetSort(b.sort)
	}

	if b.limit > 0 {
		opts.SetLimit(b.limit)
	}

	if b.skip > 0 {
		opts.SetSkip(b.skip)
	}

	return b.filter, opts, nil
}

// MergeFilter merges another filter into this builder
func (b *Builder) MergeFilter(filter bson.M) *Builder {
	if b.err != nil {
		return b
	}

	if filter == nil {
		return b
	}

	for key, value := range filter {
		if existing, exists := b.filter[key]; exists {
			// If both are bson.M, we can merge them
			if existingMap, ok1 := existing.(bson.M); ok1 {
				if valueMap, ok2 := value.(bson.M); ok2 {
					// Merge the maps
					for k, v := range valueMap {
						existingMap[k] = v
					}
					continue
				}
			}
			// Otherwise, just overwrite
		}
		b.filter[key] = value
	}

	return b
}
