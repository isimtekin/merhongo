package schema

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SchemaTag represents the schema tag and its options
type SchemaTag struct {
	Required bool
	Unique   bool
	Min      int
	Max      int
	Index    bool
}

// GenerateFromStruct automatically generates a Schema from a struct type
// It uses struct tags to define schema properties
func GenerateFromStruct(structType interface{}, options ...Option) *Schema {
	t := reflect.TypeOf(structType)

	// If a pointer is passed, get the underlying type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Ensure we're working with a struct
	if t.Kind() != reflect.Struct {
		panic("GenerateFromStruct: input must be a struct or pointer to struct")
	}

	// Create a new schema
	fields := make(map[string]Field)

	// Process each field in the struct
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Skip ID field which is handled specially
		bsonTag := field.Tag.Get("bson")
		if field.Name == "ID" || strings.HasPrefix(bsonTag, "_id") {
			continue
		}

		// Handle anonymous embedded fields
		if field.Anonymous {
			// For embedded structs, process their fields recursively
			if field.Type.Kind() == reflect.Struct {
				// Create a zero value of this type to pass to GenerateFromStruct
				embeddedValue := reflect.New(field.Type).Elem().Interface()
				embeddedSchema := GenerateFromStruct(embeddedValue)

				// Add all fields from the embedded struct to our schema
				for embeddedFieldName, embeddedField := range embeddedSchema.Fields {
					fields[embeddedFieldName] = embeddedField
				}
			}
			// Skip anonymous fields that aren't structs
			continue
		}

		// Parse schema tag
		schemaTag := parseSchemaTag(field.Tag.Get("schema"))

		// Get zero value for the field type
		zeroVal := GetZeroValue(field.Type)

		// Create field definition
		fieldDef := Field{
			Type:     zeroVal,
			Required: schemaTag.Required,
			Unique:   schemaTag.Unique,
			Index:    schemaTag.Index || schemaTag.Unique,
			Min:      schemaTag.Min,
			Max:      schemaTag.Max,
		}

		// Extract field name from bson tag if present, otherwise use the struct field name
		fieldName := field.Name
		if bsonTag != "" {
			// Parse the bson tag to get the field name
			tagParts := strings.Split(bsonTag, ",")
			if len(tagParts) > 0 && tagParts[0] != "" {
				fieldName = tagParts[0]
			}
		}

		// Add field to schema using either the bson tag name or the struct field name
		// Note: We include fields with bson:"-" in the schema because it's part of the test requirements
		fields[fieldName] = fieldDef
	}

	// Create schema with the fields
	schema := New(fields, options...)

	return schema
}

// parseSchemaTag parses the schema tag into a SchemaTag struct
func parseSchemaTag(tag string) SchemaTag {
	result := SchemaTag{}

	if tag == "" {
		return result
	}

	options := strings.Split(tag, ",")
	for _, opt := range options {
		opt = strings.TrimSpace(opt)
		switch {
		case opt == "required":
			result.Required = true
		case opt == "unique":
			result.Unique = true
		case opt == "index":
			result.Index = true
		case strings.HasPrefix(opt, "min="):
			var min int
			fmt.Sscanf(opt, "min=%d", &min)
			result.Min = min
		case strings.HasPrefix(opt, "max="):
			var max int
			fmt.Sscanf(opt, "max=%d", &max)
			result.Max = max
		}
	}

	return result
}

// GetZeroValue returns a zero value for the given type
func GetZeroValue(t reflect.Type) interface{} {
	// Handle primitive.ObjectID specially, it's a common case
	if t == reflect.TypeOf(primitive.ObjectID{}) {
		return primitive.ObjectID{}
	}

	// Handle custom types - need to check both underlying type and actual type
	if t.PkgPath() != "" {
		// Create a zero value using reflection for any custom type
		zeroVal := reflect.New(t).Elem()
		return zeroVal.Interface()
	}

	// Now switch on basic kinds
	switch t.Kind() {
	case reflect.Bool:
		return false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uint(0)
	case reflect.Float32, reflect.Float64:
		return 0.0
	case reflect.String:
		return ""
	case reflect.Slice:
		// Special handling for []interface{} type which is problematic with MakeSlice
		if t.Elem().Kind() == reflect.Interface {
			// For interface slices, just return an empty slice literal
			return []interface{}{}
		}

		// For other slice types, safely try to create a slice
		defer func() {
			recover() // Recover from any panics
		}()

		// Try to create a slice with the correct type
		v := reflect.MakeSlice(t, 0, 0)
		return v.Interface()
	case reflect.Array:
		// Arrays are fixed size, so we need special handling
		// For simplicity, return nil for arrays
		return nil
	case reflect.Struct:
		// Special handling for common types
		if t == reflect.TypeOf(time.Time{}) {
			return time.Time{}
		}
		// For other structs, return a zero struct
		v := reflect.New(t).Elem()
		return v.Interface()
	case reflect.Ptr:
		// For pointers, return nil
		return nil
	case reflect.Map:
		// Similar protection for maps
		defer func() {
			recover() // Recover from any panics
		}()

		// Try to create a map with the correct type
		v := reflect.MakeMap(t)
		return v.Interface()
	case reflect.Interface:
		// For interface types, return nil
		return nil
	default:
		return nil
	}
}
