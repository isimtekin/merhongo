// Package schema provides functionality for MongoDB document schemas
package schema

import (
	"fmt"
	"github.com/isimtekin/merhongo/errors"
	"reflect"
	"strings"
)

// Field represents a schema field definition with validation rules
type Field struct {
	Type         interface{}
	Required     bool
	Default      interface{}
	Unique       bool
	Index        bool
	Min          int
	Max          int
	Enum         []interface{}
	ValidateFunc func(interface{}) bool
}

// Schema defines the structure and validation rules for a MongoDB collection
type Schema struct {
	Fields          map[string]Field
	Timestamps      bool
	Collection      string
	Middlewares     map[string][]func(interface{}) error
	CustomValidator func(doc interface{}) error
	// ModelType holds a reference to the model type for validation purposes
	ModelType interface{}
}

// Option is a function that configures a Schema
type Option func(*Schema)

// New creates a new Schema with the specified fields and options
func New(fields map[string]Field, options ...Option) *Schema {
	schema := &Schema{
		Fields:      fields,
		Timestamps:  true,
		Middlewares: make(map[string][]func(interface{}) error),
		ModelType:   nil, // Initially empty
	}

	// Apply all provided options
	for _, option := range options {
		option(schema)
	}

	return schema
}

// WithCollection sets the collection name for the schema
func WithCollection(name string) Option {
	return func(s *Schema) {
		s.Collection = name
	}
}

// WithTimestamps enables or disables automatic timestamps
func WithTimestamps(enable bool) Option {
	return func(s *Schema) {
		s.Timestamps = enable
	}
}

// WithModelType sets the model type for the schema
func WithModelType(modelType interface{}) Option {
	return func(s *Schema) {
		s.ModelType = modelType
	}
}

// Pre adds a middleware function to be executed before the specified event
func (s *Schema) Pre(event string, fn func(interface{}) error) {
	if s.Middlewares[event] == nil {
		s.Middlewares[event] = []func(interface{}) error{}
	}
	s.Middlewares[event] = append(s.Middlewares[event], fn)
}

// ValidateDocument validates a document against the schema
func (s *Schema) ValidateDocument(doc interface{}) error {
	// Use custom validator if provided
	if s.CustomValidator != nil {
		if err := s.CustomValidator(doc); err != nil {
			return errors.Wrap(errors.ErrValidation, err.Error())
		}
		return nil
	}

	// Default validation implementation
	return s.defaultValidation(doc)
}

// defaultValidation performs basic validation based on schema rules
// defaultValidation performs basic validation based on schema rules
func (s *Schema) defaultValidation(doc interface{}) error {
	val := reflect.ValueOf(doc)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Document must be a struct
	if val.Kind() != reflect.Struct {
		return errors.WithDetails(errors.ErrValidation, "document must be a struct")
	}

	// Map to store bson field name to struct field
	bsonToStructField := make(map[string]reflect.Value)

	// Build the map of bson field names to struct fields
	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		structField := t.Field(i)
		bsonTag := structField.Tag.Get("bson")

		if bsonTag != "" {
			parts := strings.Split(bsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				bsonToStructField[parts[0]] = val.Field(i)
			}
		} else {
			// If no bson tag, use the field name
			bsonToStructField[structField.Name] = val.Field(i)
		}
	}

	// Validate required fields
	for fieldName, field := range s.Fields {
		if !field.Required {
			continue
		}

		// Find the field by its BSON name
		docField, exists := bsonToStructField[fieldName]
		if !exists {
			return errors.WithDetails(errors.ErrValidation, fmt.Sprintf("required field '%s' not found in document", fieldName))
		}

		// Check if field is zero value
		if docField.IsZero() {
			return errors.WithDetails(errors.ErrValidation, fmt.Sprintf("required field '%s' is empty", fieldName))
		}
	}

	// Validate field types
	for fieldName, field := range s.Fields {
		docField, exists := bsonToStructField[fieldName]
		if !exists {
			continue
		}

		// Validate Min/Max for numeric fields
		switch docField.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal := docField.Int()
			if field.Min != 0 && intVal < int64(field.Min) {
				return errors.WithDetails(errors.ErrValidation,
					fmt.Sprintf("field '%s' value %d is less than minimum %d", fieldName, intVal, field.Min))
			}
			if field.Max != 0 && intVal > int64(field.Max) {
				return errors.WithDetails(errors.ErrValidation,
					fmt.Sprintf("field '%s' value %d is greater than maximum %d", fieldName, intVal, field.Max))
			}
		case reflect.Float32, reflect.Float64:
			floatVal := docField.Float()
			if field.Min != 0 && floatVal < float64(field.Min) {
				return errors.WithDetails(errors.ErrValidation,
					fmt.Sprintf("field '%s' value %f is less than minimum %d", fieldName, floatVal, field.Min))
			}
			if field.Max != 0 && floatVal > float64(field.Max) {
				return errors.WithDetails(errors.ErrValidation,
					fmt.Sprintf("field '%s' value %f is greater than maximum %d", fieldName, floatVal, field.Max))
			}
		}

		// Validate enum if present
		if len(field.Enum) > 0 {
			found := false
			for _, enumVal := range field.Enum {
				enumReflectVal := reflect.ValueOf(enumVal)
				if reflect.DeepEqual(docField.Interface(), enumReflectVal.Interface()) {
					found = true
					break
				}
			}
			if !found {
				return errors.WithDetails(errors.ErrValidation,
					fmt.Sprintf("field '%s' value is not in the allowed enum values", fieldName))
			}
		}

		// Run custom validation function if present
		if field.ValidateFunc != nil {
			if !field.ValidateFunc(docField.Interface()) {
				return errors.WithDetails(errors.ErrValidation,
					fmt.Sprintf("field '%s' failed custom validation", fieldName))
			}
		}
	}

	return nil
}
