package schema_test

import (
	"context"
	"fmt"
	schema2 "github.com/isimtekin/merhongo/schema"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ----------------- Test Structs -----------------

// TestStruct for basic schema generation testing
type TestStruct struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `schema:"required,unique"`
	Age         int                `schema:"required,min=18,max=100"`
	Email       string             `schema:"required,unique"`
	Active      bool               `schema:"required"`
	Description string
	Tags        []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AdvancedTestStruct with various field types and tags
type AdvancedTestStruct struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	Name            string             `schema:"required,unique"`
	Age             int                `schema:"min=18,max=100"`
	Email           *string            `schema:"required"`
	Score           float64            `schema:"min=0,max=10"`
	IsActive        bool               `schema:"required"`
	Tags            []string
	Metadata        map[string]string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CustomID        primitive.ObjectID `schema:"required"`
	IntArray        []int              `schema:"required"`
	StringMap       map[string]int
	NestedStruct    NestedStruct
	NestedStructPtr *NestedStruct `schema:"required"`
	IntPointer      *int
	//AnySlice        []interface{}
}

// NestedStruct for testing nested structs in schema generation
type NestedStruct struct {
	Field1 string `schema:"required"`
	Field2 int    `schema:"min=1,max=10"`
}

// AnonymousFieldStruct for testing embedded anonymous fields
type AnonymousFieldStruct struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty"`
	NestedStruct                     // Anonymous embedded struct
	string       `schema:"required"` // Anonymous primitive type (unsupported)
}

// MultiTagStruct for testing various tag combinations
type MultiTagStruct struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `schema:"required,unique,min=2,max=100"` // Multiple options
	Description string             `schema:""`                              // Empty schema tag
	Status      string             `schema:"unknown_tag"`                   // Unknown tag option
	Age         int                `schema:"min=18,min=21"`                 // Duplicate tag (last one should win)
	Score       float64            `schema:"max=10,min=20"`                 // Conflicting constraints
}

// Custom type for testing custom types
type CustomEnum string

const (
	EnumValue1 CustomEnum = "value1"
	EnumValue2 CustomEnum = "value2"
)

// CustomTypeStruct for testing custom types and enum handling
type CustomTypeStruct struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	EnumType CustomEnum         `schema:"required"`
}

// NoIDStruct doesn't have an ID field
type NoIDStruct struct {
	Name string `schema:"required"`
	Age  int    `schema:"min=18"`
}

// IDWithDifferentNameStruct has ID field with different name
type IDWithDifferentNameStruct struct {
	Identifier primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `schema:"required"`
}

// RecursiveStruct contains a reference to itself (potential infinite loop)
type RecursiveStruct struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `schema:"required"`
	Parent   *RecursiveStruct   // Pointer to same type
	Children []RecursiveStruct  // Slice of same type
}

// Empty struct for testing edge cases
type EmptyStruct struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
}

// Struct with only private fields
type PrivateFieldsStruct struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	name  string             `schema:"required"` // Private field
	email string             `schema:"required"` // Private field
}

// UserForTest defines a test struct for real-world usage testing
type UserForTest struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username" schema:"required,unique"`
	Email     string             `bson:"email" json:"email" schema:"required,unique"`
	Age       int                `bson:"age" json:"age" schema:"min=13,max=120"`
	IsActive  bool               `bson:"isActive" json:"isActive"`
	Role      string             `bson:"role" json:"role" schema:"required"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// ----------------- Basic Tests -----------------

func TestGenerateFromStruct(t *testing.T) {
	schema := schema2.GenerateFromStruct(TestStruct{}, schema2.WithCollection("test_collection"))

	// Check collection name
	if schema.Collection != "test_collection" {
		t.Errorf("Expected collection name to be 'test_collection', got '%s'", schema.Collection)
	}

	// Check that fields were properly generated
	tests := []struct {
		fieldName string
		required  bool
		unique    bool
		min       int
		max       int
	}{
		{"Name", true, true, 0, 0},
		{"Age", true, false, 18, 100},
		{"Email", true, true, 0, 0},
		{"Active", true, false, 0, 0},
		{"Description", false, false, 0, 0},
		{"Tags", false, false, 0, 0},
	}

	for _, tc := range tests {
		field, exists := schema.Fields[tc.fieldName]
		if !exists {
			t.Errorf("Expected field '%s' to exist in schema", tc.fieldName)
			continue
		}

		if field.Required != tc.required {
			t.Errorf("Field '%s': expected Required=%v, got %v", tc.fieldName, tc.required, field.Required)
		}

		if field.Unique != tc.unique {
			t.Errorf("Field '%s': expected Unique=%v, got %v", tc.fieldName, tc.unique, field.Unique)
		}

		if field.Min != tc.min {
			t.Errorf("Field '%s': expected Min=%d, got %d", tc.fieldName, tc.min, field.Min)
		}

		if field.Max != tc.max {
			t.Errorf("Field '%s': expected Max=%d, got %d", tc.fieldName, tc.max, field.Max)
		}
	}

	// Check that ID field is not included
	if _, exists := schema.Fields["ID"]; exists {
		t.Error("ID field should not be included in schema fields")
	}

	// Check that timestamps are enabled by default
	if !schema.Timestamps {
		t.Error("Expected timestamps to be enabled by default")
	}
}

// Test with pointer to struct
func TestGenerateFromStructPointer(t *testing.T) {
	schema := schema2.GenerateFromStruct(&TestStruct{})

	if len(schema.Fields) == 0 {
		t.Error("Expected fields to be generated from struct pointer")
	}

	// Check one field to ensure it worked
	if field, exists := schema.Fields["Name"]; !exists || !field.Required {
		t.Error("Expected 'Name' field to exist and be required")
	}
}

// Test with non-struct type (should panic)
func TestGenerateFromNonStruct(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when using non-struct type")
		}
	}()

	schema2.GenerateFromStruct("not a struct")
}

// Test with custom options
func TestGenerateWithOptions(t *testing.T) {
	schema := schema2.GenerateFromStruct(TestStruct{},
		schema2.WithCollection("custom_collection"),
		schema2.WithTimestamps(false),
	)

	if schema.Collection != "custom_collection" {
		t.Errorf("Expected collection name to be 'custom_collection', got '%s'", schema.Collection)
	}

	if schema.Timestamps {
		t.Error("Expected timestamps to be disabled")
	}
}

// ----------------- Advanced Field Tests -----------------

func TestGenerateAdvancedFieldTypes(t *testing.T) {
	schema := schema2.GenerateFromStruct(AdvancedTestStruct{})

	// Check various field types
	testCases := []struct {
		fieldName string
		fieldType reflect.Kind
		required  bool
		unique    bool
		min       int
		max       int
	}{
		{"Name", reflect.String, true, true, 0, 0},
		{"Age", reflect.Int, false, false, 18, 100},
		{"Score", reflect.Float64, false, false, 0, 10},
		{"IsActive", reflect.Bool, true, false, 0, 0},
		{"Tags", reflect.Slice, false, false, 0, 0},
		{"Metadata", reflect.Map, false, false, 0, 0},
		{"CustomID", reflect.Struct, true, false, 0, 0},
		{"IntArray", reflect.Slice, true, false, 0, 0},
		{"StringMap", reflect.Map, false, false, 0, 0},
	}

	for _, tc := range testCases {
		field, exists := schema.Fields[tc.fieldName]
		if !exists {
			t.Errorf("Expected field '%s' to exist in schema", tc.fieldName)
			continue
		}

		if field.Required != tc.required {
			t.Errorf("Field '%s': expected Required=%v, got %v", tc.fieldName, tc.required, field.Required)
		}

		if field.Unique != tc.unique {
			t.Errorf("Field '%s': expected Unique=%v, got %v", tc.fieldName, tc.unique, field.Unique)
		}

		if field.Min != tc.min {
			t.Errorf("Field '%s': expected Min=%d, got %d", tc.fieldName, tc.min, field.Min)
		}

		if field.Max != tc.max {
			t.Errorf("Field '%s': expected Max=%d, got %d", tc.fieldName, tc.max, field.Max)
		}
	}
}

func TestGeneratePointerFields(t *testing.T) {
	schema := schema2.GenerateFromStruct(AdvancedTestStruct{})

	// Test pointer fields
	emailField, exists := schema.Fields["Email"]
	if !exists {
		t.Fatal("Expected 'Email' field (pointer to string) to exist in schema")
	}

	if !emailField.Required {
		t.Errorf("Expected Email field to be required")
	}

	// Check nested struct pointer
	nestedPtrField, exists := schema.Fields["NestedStructPtr"]
	if !exists {
		t.Fatal("Expected 'NestedStructPtr' field to exist in schema")
	}

	if !nestedPtrField.Required {
		t.Errorf("Expected NestedStructPtr field to be required")
	}
}

func TestGenerateCustomTypes(t *testing.T) {
	schema := schema2.GenerateFromStruct(CustomTypeStruct{})

	// Test custom enum type
	enumField, exists := schema.Fields["EnumType"]
	if !exists {
		t.Fatal("Expected 'EnumType' field to exist in schema")
	}

	if !enumField.Required {
		t.Errorf("Expected EnumType field to be required")
	}

	// Check the type - we expect it to be an empty CustomEnum
	customEnum, ok := enumField.Type.(CustomEnum)
	if !ok {
		t.Errorf("Expected EnumType.Type to be of type CustomEnum, got %T", enumField.Type)
	} else if customEnum != CustomEnum("") {
		t.Errorf("Expected EnumType.Type to be empty CustomEnum, got %q", customEnum)
	}
}

// ----------------- Edge Cases Tests -----------------

func TestGenerateEmptyStruct(t *testing.T) {
	schema := schema2.GenerateFromStruct(EmptyStruct{})

	// Should have no fields (ID is excluded)
	if len(schema.Fields) != 0 {
		t.Errorf("Expected empty struct to generate no fields, got %d", len(schema.Fields))
	}
}

func TestGeneratePrivateFields(t *testing.T) {
	schema := schema2.GenerateFromStruct(PrivateFieldsStruct{})

	// Private fields should be ignored
	if len(schema.Fields) != 0 {
		t.Errorf("Expected private fields to be ignored, got %d fields", len(schema.Fields))
	}
}

func TestAnonymousFields(t *testing.T) {
	schema := schema2.GenerateFromStruct(AnonymousFieldStruct{})

	// Check if anonymous struct fields are properly handled
	field1, exists := schema.Fields["Field1"]
	if !exists {
		t.Error("Expected embedded Field1 to be included in schema")
	} else if !field1.Required {
		t.Error("Expected embedded Field1 to maintain its required flag")
	}

	field2, exists := schema.Fields["Field2"]
	if !exists {
		t.Error("Expected embedded Field2 to be included in schema")
	} else if field2.Min != 1 || field2.Max != 10 {
		t.Errorf("Expected embedded Field2 to maintain min=1,max=10, got min=%d,max=%d", field2.Min, field2.Max)
	}

	// Anonymous primitive type should be ignored
	if _, exists := schema.Fields["string"]; exists {
		t.Error("Expected anonymous primitive type to be ignored")
	}
}

func TestMultipleTagOptions(t *testing.T) {
	schema := schema2.GenerateFromStruct(MultiTagStruct{})

	// Field with multiple valid options
	nameField, exists := schema.Fields["Name"]
	if !exists {
		t.Fatal("Expected 'Name' field to exist in schema")
	}
	if !nameField.Required || !nameField.Unique || nameField.Min != 2 || nameField.Max != 100 {
		t.Errorf("Expected Name to have required=true, unique=true, min=2, max=100; got %+v", nameField)
	}

	// Field with empty schema tag
	descField, exists := schema.Fields["Description"]
	if !exists {
		t.Fatal("Expected 'Description' field to exist in schema")
	}
	if descField.Required || descField.Unique || descField.Min != 0 || descField.Max != 0 {
		t.Errorf("Expected Description to have all default values; got %+v", descField)
	}

	// Field with unknown tag option
	statusField, exists := schema.Fields["Status"]
	if !exists {
		t.Fatal("Expected 'Status' field to exist in schema")
	}
	if statusField.Required || statusField.Unique || statusField.Min != 0 || statusField.Max != 0 {
		t.Errorf("Expected Status to have all default values despite unknown tag; got %+v", statusField)
	}

	// Field with duplicate tags (last one should win)
	ageField, exists := schema.Fields["Age"]
	if !exists {
		t.Fatal("Expected 'Age' field to exist in schema")
	}
	if ageField.Min != 21 {
		t.Errorf("Expected Age min to be 21 (last tag value), got %d", ageField.Min)
	}

	// Field with conflicting constraints
	scoreField, exists := schema.Fields["Score"]
	if !exists {
		t.Fatal("Expected 'Score' field to exist in schema")
	}
	if scoreField.Min != 20 || scoreField.Max != 10 {
		t.Errorf("Expected Score to have min=20, max=10; got min=%d, max=%d", scoreField.Min, scoreField.Max)
	}
}

func TestNoIDField(t *testing.T) {
	schema := schema2.GenerateFromStruct(NoIDStruct{})

	// Should process normally without ID field
	if len(schema.Fields) != 2 {
		t.Errorf("Expected 2 fields in NoIDStruct schema, got %d", len(schema.Fields))
	}

	nameField, exists := schema.Fields["Name"]
	if !exists || !nameField.Required {
		t.Error("Expected Name field to exist and be required")
	}

	ageField, exists := schema.Fields["Age"]
	if !exists || ageField.Min != 18 {
		t.Error("Expected Age field to exist with min=18")
	}
}

func TestIDWithDifferentName(t *testing.T) {
	schema := schema2.GenerateFromStruct(IDWithDifferentNameStruct{})

	// Should not include the ID field with a different name in the schema
	if _, exists := schema.Fields["Identifier"]; exists {
		t.Error("Expected Identifier field to be excluded as it's marked as _id")
	}

	// Should still include the Name field
	nameField, exists := schema.Fields["Name"]
	if !exists || !nameField.Required {
		t.Error("Expected Name field to exist and be required")
	}
}

func TestRecursiveStructs(t *testing.T) {
	// This test primarily checks that we don't get stuck in infinite recursion
	schema := schema2.GenerateFromStruct(RecursiveStruct{})

	// Should handle recursive structures without crashing
	if len(schema.Fields) != 3 {
		t.Errorf("Expected 3 fields in RecursiveStruct schema, got %d", len(schema.Fields))
	}

	_, exists := schema.Fields["Parent"]
	if !exists {
		t.Error("Expected Parent field to exist in schema")
	}

	_, exists = schema.Fields["Children"]
	if !exists {
		t.Error("Expected Children field to exist in schema")
	}

	nameField, exists := schema.Fields["Name"]
	if !exists || !nameField.Required {
		t.Error("Expected Name field to exist and be required")
	}
}

func TestSpecialBsonTags(t *testing.T) {
	// Test special bson tag handling
	type SpecialBsonTagStruct struct {
		ID          primitive.ObjectID `bson:"_id,omitempty"`
		Name        string             `bson:"fullName" schema:"required"`             // Different field name
		Description string             `bson:"desc,omitempty" schema:""`               // With omitempty
		InternalID  string             `bson:"-" schema:"required"`                    // Ignored in BSON
		CreatedAt   time.Time          `bson:"created_at,omitempty" schema:"required"` // Snake case with omitempty
	}

	schema := schema2.GenerateFromStruct(SpecialBsonTagStruct{})

	// All fields except ID should be in the schema with their BSON names
	expectedFields := []string{"fullName", "desc", "-", "created_at"}
	for _, fieldName := range expectedFields {
		if _, exists := schema.Fields[fieldName]; !exists {
			t.Errorf("Expected field '%s' to exist in schema", fieldName)
		}
	}

	// Required fields should be marked as required
	requiredFields := []string{"fullName", "-", "created_at"}
	for _, fieldName := range requiredFields {
		field, exists := schema.Fields[fieldName]
		if !exists {
			t.Errorf("Field '%s' missing from schema", fieldName)
			continue
		}

		if !field.Required {
			t.Errorf("Expected field '%s' to be required", fieldName)
		}
	}
}

func TestInvalidTagValues(t *testing.T) {
	// Test with invalid tag values
	type InvalidTagStruct struct {
		ID       primitive.ObjectID `bson:"_id,omitempty"`
		Age      int                `schema:"min=invalid"` // Invalid min
		MaxScore int                `schema:"max=invalid"` // Invalid max
	}

	schema := schema2.GenerateFromStruct(InvalidTagStruct{})

	// Should not cause a panic, but min/max should be 0
	ageField, exists := schema.Fields["Age"]
	if !exists {
		t.Fatal("Expected 'Age' field to exist in schema")
	}

	if ageField.Min != 0 {
		t.Errorf("Expected invalid min tag to result in 0, got %d", ageField.Min)
	}

	maxScoreField, exists := schema.Fields["MaxScore"]
	if !exists {
		t.Fatal("Expected 'MaxScore' field to exist in schema")
	}

	if maxScoreField.Max != 0 {
		t.Errorf("Expected invalid max tag to result in 0, got %d", maxScoreField.Max)
	}
}

func TestZeroValueGeneration(t *testing.T) {
	type ZeroValueStruct struct {
		ID          primitive.ObjectID `bson:"_id,omitempty"`
		BoolField   bool               `schema:"required"`
		IntField    int                `schema:"min=1"`
		FloatField  float64            `schema:"min=0.1"`
		StringField string             `schema:"required"`
		SliceField  []int              `schema:"required"`
		MapField    map[string]int     `schema:"required"`
		TimeField   time.Time          `schema:"required"`
		ObjectField primitive.ObjectID `schema:"required"`
	}

	schema := schema2.GenerateFromStruct(ZeroValueStruct{})

	// Check zero values for different types
	testCases := []struct {
		fieldName string
		expected  interface{}
	}{
		{"BoolField", false},
		{"IntField", 0},
		{"FloatField", 0.0},
		{"StringField", ""},
		// Note: Can't directly compare slice and map instances, just check type
	}

	for _, tc := range testCases {
		field, exists := schema.Fields[tc.fieldName]
		if !exists {
			t.Errorf("Expected field '%s' to exist in schema", tc.fieldName)
			continue
		}

		if field.Type != tc.expected {
			t.Errorf("Field '%s': expected zero value %v, got %v", tc.fieldName, tc.expected, field.Type)
		}
	}

	// For complex types, check the type rather than value
	sliceField, exists := schema.Fields["SliceField"]
	if !exists {
		t.Error("Expected SliceField to exist in schema")
	} else if _, ok := sliceField.Type.([]int); !ok {
		t.Errorf("Expected SliceField.Type to be []int, got %T", sliceField.Type)
	}

	mapField, exists := schema.Fields["MapField"]
	if !exists {
		t.Error("Expected MapField to exist in schema")
	} else if _, ok := mapField.Type.(map[string]int); !ok {
		t.Errorf("Expected MapField.Type to be map[string]int, got %T", mapField.Type)
	}

	timeField, exists := schema.Fields["TimeField"]
	if !exists {
		t.Error("Expected TimeField to exist in schema")
	} else if _, ok := timeField.Type.(time.Time); !ok {
		t.Errorf("Expected TimeField.Type to be time.Time, got %T", timeField.Type)
	}

	// Check ObjectID field specifically
	objectField, exists := schema.Fields["ObjectField"]
	if !exists {
		t.Error("Expected ObjectField to exist in schema")
	} else {
		// Check if it's actually an ObjectID
		_, ok := objectField.Type.(primitive.ObjectID)
		if !ok {
			t.Errorf("Expected ObjectField.Type to be primitive.ObjectID, got %T", objectField.Type)
		}
	}
}

// ----------------- Implementation Tests -----------------

func TestCustomizeSchemaAfterGeneration(t *testing.T) {
	// Generate schema, then customize it
	schema := schema2.GenerateFromStruct(AdvancedTestStruct{})

	// Add enum values
	if field, exists := schema.Fields["Name"]; exists {
		field.Enum = []interface{}{"John", "Jane", "Joe"}
		schema.Fields["Name"] = field
	}

	// Add custom validator
	if field, exists := schema.Fields["Email"]; exists {
		field.ValidateFunc = func(val interface{}) bool {
			if valPtr, ok := val.(*string); ok && valPtr != nil {
				email := *valPtr
				return len(email) > 0 && strings.Contains(email, "@")
			}
			return false
		}
		schema.Fields["Email"] = field
	}

	// Add default value
	if field, exists := schema.Fields["IsActive"]; exists {
		field.Default = true
		schema.Fields["IsActive"] = field
	}

	// Verify customizations
	if len(schema.Fields["Name"].Enum) != 3 {
		t.Errorf("Expected Name field to have 3 enum values")
	}

	if schema.Fields["IsActive"].Default != true {
		t.Errorf("Expected IsActive field to have default=true")
	}

	if schema.Fields["Email"].ValidateFunc == nil {
		t.Errorf("Expected Email field to have a validator function")
	}
}

func TestSchemaGenerationWithValidation(t *testing.T) {
	// Generate schema from struct
	userSchema := schema2.GenerateFromStruct(UserForTest{},
		schema2.WithCollection("users"),
		schema2.WithTimestamps(true),
	)

	// Add default value for role - BSON alan adı kullanılıyor
	if field, exists := userSchema.Fields["role"]; exists {
		field.Default = "user"
		userSchema.Fields["role"] = field
	}

	// Test cases
	testCases := []struct {
		name        string
		user        UserForTest
		shouldError bool
		errorField  string // BSON field name olarak kullanılacak
	}{
		{
			name: "Valid user",
			user: UserForTest{
				Username: "testuser",
				Email:    "test@example.com",
				Age:      25,
				Role:     "user",
				IsActive: true,
			},
			shouldError: false,
		},
		{
			name: "Missing required field - username",
			user: UserForTest{
				Email:    "test@example.com",
				Age:      25,
				Role:     "user",
				IsActive: true,
			},
			shouldError: true,
			errorField:  "username", // BSON field name
		},
		{
			name: "Missing required field - email",
			user: UserForTest{
				Username: "testuser",
				Age:      25,
				Role:     "user",
				IsActive: true,
			},
			shouldError: true,
			errorField:  "email", // BSON field name
		},
		{
			name: "Age below minimum",
			user: UserForTest{
				Username: "testuser",
				Email:    "test@example.com",
				Age:      10, // Below min of 13
				Role:     "user",
				IsActive: true,
			},
			shouldError: true,
			errorField:  "age", // BSON field name
		},
		{
			name: "Age above maximum",
			user: UserForTest{
				Username: "testuser",
				Email:    "test@example.com",
				Age:      130, // Above max of 120
				Role:     "user",
				IsActive: true,
			},
			shouldError: true,
			errorField:  "age", // BSON field name
		},
		{
			name: "Missing required field - role",
			user: UserForTest{
				Username: "testuser",
				Email:    "test@example.com",
				Age:      25,
				IsActive: true,
			},
			shouldError: true,
			errorField:  "role", // BSON field name
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := userSchema.ValidateDocument(&tc.user)

			if tc.shouldError && err == nil {
				t.Errorf("Expected validation error for field %s, but got none", tc.errorField)
			}

			if !tc.shouldError && err != nil {
				t.Errorf("Expected no validation error, but got: %v", err)
			}

			if tc.shouldError && err != nil {
				if tc.errorField != "" {
					errMsg := err.Error()
					if !strings.Contains(errMsg, tc.errorField) {
						t.Errorf("Expected error for field %s, but got error: %v", tc.errorField, err)
					}
				}
			}
		})
	}
}

func TestSchemaGenerationWithMiddleware(t *testing.T) {
	// Generate schema from struct
	userSchema := schema2.GenerateFromStruct(UserForTest{},
		schema2.WithCollection("users"),
		schema2.WithTimestamps(true),
	)

	// Add a middleware to set default values
	middlewareCalled := false
	userSchema.Pre("save", func(doc interface{}) error {
		middlewareCalled = true
		if user, ok := doc.(*UserForTest); ok {
			// Set defaults
			if user.Role == "" {
				user.Role = "user"
			}
			if user.Age == 0 {
				user.Age = 18
			}
		}
		return nil
	})

	// Create a user without some fields
	user := &UserForTest{
		Username: "testuser",
		Email:    "test@example.com",
		// Age not set
		// Role not set
	}

	// Call middleware directly to test
	err := userSchema.Middlewares["save"][0](user)

	// Validate results
	if err != nil {
		t.Errorf("Middleware returned error: %v", err)
	}

	if !middlewareCalled {
		t.Error("Middleware was not called")
	}

	if user.Role != "user" {
		t.Errorf("Expected middleware to set default role='user', got '%s'", user.Role)
	}

	if user.Age != 18 {
		t.Errorf("Expected middleware to set default age=18, got %d", user.Age)
	}
}

func TestValidateWithCustomValidator(t *testing.T) {
	// Generate schema from struct
	userSchema := schema2.GenerateFromStruct(UserForTest{},
		schema2.WithCollection("users"),
		schema2.WithTimestamps(true),
	)

	// Add custom validator that validates email format
	customValidatorCalled := false
	userSchema.CustomValidator = func(doc interface{}) error {
		customValidatorCalled = true
		if user, ok := doc.(*UserForTest); ok {
			// Simple email validation
			if user.Email != "" && !strings.Contains(user.Email, "@") {
				return fmt.Errorf("invalid email format")
			}
		}
		return nil
	}

	// Valid user
	validUser := &UserForTest{
		Username: "testuser",
		Email:    "test@example.com",
		Age:      25,
		Role:     "user",
	}

	// Invalid user (bad email)
	invalidUser := &UserForTest{
		Username: "testuser",
		Email:    "invalid-email", // No @
		Age:      25,
		Role:     "user",
	}

	// Test valid user
	err := userSchema.ValidateDocument(validUser)
	if err != nil {
		t.Errorf("Expected no validation error for valid user, got: %v", err)
	}

	if !customValidatorCalled {
		t.Error("Custom validator was not called for valid user")
	}

	// Reset flag
	customValidatorCalled = false

	// Test invalid user
	err = userSchema.ValidateDocument(invalidUser)
	if err == nil {
		t.Error("Expected validation error for invalid email, but got none")
	}

	if !customValidatorCalled {
		t.Error("Custom validator was not called for invalid user")
	}
}

// ----------------- MongoDB Integration Tests -----------------

// Integration test flag - set to true to run MongoDB integration tests
// They will be skipped by default
var runIntegrationTests = false
var mongoTestURI = "mongodb://localhost:27017"
var mongoTestDB = "merhongo_test_schema_generator"

// TestMongoDBSchemaIntegration performs an integration test with a real MongoDB connection
func TestMongoDBSchemaIntegration(t *testing.T) {
	if !runIntegrationTests {
		t.Skip("Skipping MongoDB integration tests")
	}

	// Connect to MongoDB
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoTestURI))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Ping to ensure connection is established
	if err = client.Ping(ctx, nil); err != nil {
		t.Fatalf("Failed to ping MongoDB: %v", err)
	}

	// Get test database and collection
	db := client.Database(mongoTestDB)

	// Generate schema from struct
	userSchema := schema2.GenerateFromStruct(UserForTest{},
		schema2.WithCollection("users_integration_test"),
		schema2.WithTimestamps(true),
	)

	// Use the collection name from schema
	coll := db.Collection(userSchema.Collection)

	// Clear collection before test
	coll.Drop(ctx)

	// Create unique indexes based on schema
	for fieldName, field := range userSchema.Fields {
		if field.Unique {
			indexModel := mongo.IndexModel{
				Keys:    bson.D{{Key: fieldName, Value: 1}},
				Options: options.Index().SetUnique(true),
			}
			_, err := coll.Indexes().CreateOne(ctx, indexModel)
			if err != nil {
				t.Fatalf("Failed to create index for field '%s': %v", fieldName, err)
			}
		}
	}

	// Test user creation
	user := UserForTest{
		ID:       primitive.NewObjectID(),
		Username: "integrationtest",
		Email:    "integration@test.com",
		Age:      30,
		Role:     "user",
		IsActive: true,
	}

	// Validate user against schema
	err = userSchema.ValidateDocument(&user)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Insert user into MongoDB
	_, err = coll.InsertOne(ctx, user)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Retrieve user from MongoDB
	var retrievedUser UserForTest
	err = coll.FindOne(ctx, bson.M{"username": "integrationtest"}).Decode(&retrievedUser)
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}

	// Verify retrieved user
	if retrievedUser.Username != user.Username {
		t.Errorf("Expected username '%s', got '%s'", user.Username, retrievedUser.Username)
	}

	if retrievedUser.Email != user.Email {
		t.Errorf("Expected email '%s', got '%s'", user.Email, retrievedUser.Email)
	}

	// Test duplicate key error (unique constraint)
	duplicateUser := UserForTest{
		ID:       primitive.NewObjectID(),
		Username: "integrationtest", // Same username (should fail)
		Email:    "another@test.com",
		Age:      30,
		Role:     "user",
		IsActive: true,
	}

	_, err = coll.InsertOne(ctx, duplicateUser)
	if err == nil {
		t.Error("Expected error for duplicate username, but got none")
	}

	// Clean up
	coll.Drop(ctx)
}

// Helper functions for string tests
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// TestArrayHandling tests how arrays are handled in schema generation
func TestArrayHandling(t *testing.T) {
	// Define a struct with array fields
	type ArrayStruct struct {
		ID          primitive.ObjectID `bson:"_id,omitempty"`
		StringArray [3]string          `schema:"required"`
		IntArray    [5]int             `schema:"min=1"`
		BoolArray   [2]bool
		ObjArray    [3]primitive.ObjectID
	}

	schema := schema2.GenerateFromStruct(ArrayStruct{})

	// Check if array fields exist in schema
	_, exists := schema.Fields["StringArray"]
	if !exists {
		t.Error("Expected StringArray field to exist in schema")
	}

	_, exists = schema.Fields["IntArray"]
	if !exists {
		t.Error("Expected IntArray field to exist in schema")
	}

	_, exists = schema.Fields["BoolArray"]
	if !exists {
		t.Error("Expected BoolArray field to exist in schema")
	}

	_, exists = schema.Fields["ObjArray"]
	if !exists {
		t.Error("Expected ObjArray field to exist in schema")
	}

	// Check array field types
	// For arrays, we expect getZeroValue to return nil as specified in the code
	for _, fieldName := range []string{"StringArray", "IntArray", "BoolArray", "ObjArray"} {
		field, exists := schema.Fields[fieldName]
		if !exists {
			continue
		}

		if field.Type != nil {
			t.Errorf("Expected %s field type to be nil (for arrays), got %v", fieldName, field.Type)
		}
	}

	// Verify other properties are still set correctly
	stringArrayField, _ := schema.Fields["StringArray"]
	if !stringArrayField.Required {
		t.Error("Expected StringArray to have Required=true")
	}

	intArrayField, _ := schema.Fields["IntArray"]
	if intArrayField.Min != 1 {
		t.Errorf("Expected IntArray to have Min=1, got %d", intArrayField.Min)
	}
}

// TestStructHandling tests how different struct types are handled in schema generation
func TestStructHandling(t *testing.T) {
	// Define a struct with various struct fields
	type NestedStructA struct {
		FieldA string `schema:"required"`
		FieldB int    `schema:"min=10"`
	}

	type NestedStructB struct {
		FieldC bool
		FieldD []string
	}

	type ComplexStructsTest struct {
		ID           primitive.ObjectID `bson:"_id,omitempty"`
		TimeField    time.Time          `schema:"required"`
		NestedA      NestedStructA
		NestedB      NestedStructB    `schema:"required"`
		CustomStruct CustomTypeStruct // Using existing type from other tests
	}

	schema := schema2.GenerateFromStruct(ComplexStructsTest{})

	// Check if struct fields exist
	_, exists := schema.Fields["TimeField"]
	if !exists {
		t.Error("Expected TimeField to exist in schema")
	}

	_, exists = schema.Fields["NestedA"]
	if !exists {
		t.Error("Expected NestedA to exist in schema")
	}

	_, exists = schema.Fields["NestedB"]
	if !exists {
		t.Error("Expected NestedB to exist in schema")
	}

	_, exists = schema.Fields["CustomStruct"]
	if !exists {
		t.Error("Expected CustomStruct to exist in schema")
	}

	// Test time.Time handling
	timeField, _ := schema.Fields["TimeField"]
	if _, ok := timeField.Type.(time.Time); !ok {
		t.Errorf("Expected TimeField.Type to be time.Time, got %T", timeField.Type)
	}

	// For other structs, should return zero value of that struct type
	nestedAField, _ := schema.Fields["NestedA"]
	if reflect.TypeOf(nestedAField.Type) != reflect.TypeOf(NestedStructA{}) {
		t.Errorf("Expected NestedA.Type to be NestedStructA, got %T", nestedAField.Type)
	}

	nestedBField, _ := schema.Fields["NestedB"]
	if reflect.TypeOf(nestedBField.Type) != reflect.TypeOf(NestedStructB{}) {
		t.Errorf("Expected NestedB.Type to be NestedStructB, got %T", nestedBField.Type)
	}

	// Verify other properties are still set correctly
	if !timeField.Required {
		t.Error("Expected TimeField to have Required=true")
	}

	if !nestedBField.Required {
		t.Error("Expected NestedB to have Required=true")
	}
}

// TestReflectionEdgeCases tests edge cases in the reflection handling
func TestReflectionEdgeCases(t *testing.T) {
	// Test direct calls to getZeroValue for edge cases

	// Test array type
	arrayType := reflect.TypeOf([3]string{})
	arrayZero := schema2.GetZeroValue(arrayType)
	if arrayZero != nil {
		t.Errorf("Expected nil for array type, got %v", arrayZero)
	}

	// Test time.Time
	timeType := reflect.TypeOf(time.Time{})
	timeZero := schema2.GetZeroValue(timeType)
	if _, ok := timeZero.(time.Time); !ok {
		t.Errorf("Expected time.Time for time type, got %T", timeZero)
	}

	// Test struct with reflection.New
	type TestStruct struct {
		Field1 string
		Field2 int
	}

	structType := reflect.TypeOf(TestStruct{})
	structZero := schema2.GetZeroValue(structType)

	// Check that returned value is of correct type
	if _, ok := structZero.(TestStruct); !ok {
		t.Errorf("Expected TestStruct for struct type, got %T", structZero)
	}

	// Check it's a zero value (empty fields)
	testStruct, _ := structZero.(TestStruct)
	if testStruct.Field1 != "" || testStruct.Field2 != 0 {
		t.Errorf("Expected zero struct, got Field1=%q, Field2=%d",
			testStruct.Field1, testStruct.Field2)
	}
}

// TestSpecialTimeHandling tests specifically the time.Time handling code path
func TestSpecialTimeHandling(t *testing.T) {
	// Test struct with time.Time field
	type TimeStruct struct {
		ID       primitive.ObjectID `bson:"_id,omitempty"`
		Created  time.Time          `schema:"required"`
		Updated  time.Time
		Optional *time.Time `schema:"required"`
	}

	schema := schema2.GenerateFromStruct(TimeStruct{})

	// Check if time fields exist in schema
	createdField, exists := schema.Fields["Created"]
	if !exists {
		t.Fatal("Expected Created field to exist in schema")
	}

	updatedField, exists := schema.Fields["Updated"]
	if !exists {
		t.Fatal("Expected Updated field to exist in schema")
	}

	// Test time.Time special handling - should return empty time.Time{}
	// First check type
	if _, ok := createdField.Type.(time.Time); !ok {
		t.Errorf("Expected Created.Type to be time.Time, got %T", createdField.Type)
	}

	// Then check it's the zero value
	timeVal, _ := createdField.Type.(time.Time)
	zeroTime := time.Time{}
	if !timeVal.Equal(zeroTime) {
		t.Errorf("Expected Created.Type to be zero time, got %v", timeVal)
	}

	// Similarly for Updated field
	if _, ok := updatedField.Type.(time.Time); !ok {
		t.Errorf("Expected Updated.Type to be time.Time, got %T", updatedField.Type)
	}

	// Check that Required flag is properly set
	if !createdField.Required {
		t.Error("Expected Created field to be required")
	}

	if updatedField.Required {
		t.Error("Expected Updated field to not be required")
	}

	// Test direct call to getZeroValue with time.Time type
	timeType := reflect.TypeOf(time.Time{})
	timeZero := schema2.GetZeroValue(timeType)

	// Check if it returns the correct type
	if _, ok := timeZero.(time.Time); !ok {
		t.Fatalf("Expected getZeroValue to return time.Time for time.Time type, got %T", timeZero)
	}

	// Check if it's the zero time value
	returnedTime, _ := timeZero.(time.Time)
	if !returnedTime.Equal(zeroTime) {
		t.Errorf("Expected zero time, got %v", returnedTime)
	}
}

// TestReflectNewElementHandling tests the reflect.New(t).Elem() code path
func TestReflectNewElementHandling(t *testing.T) {
	// Define some custom struct types
	type SimpleStruct struct {
		Field1 string
		Field2 int
		Field3 bool
	}

	type ComplexStruct struct {
		Name     string
		Value    float64
		Enabled  bool
		SubItems []string
	}

	// Test direct calls to getZeroValue
	simpleType := reflect.TypeOf(SimpleStruct{})
	simpleZero := schema2.GetZeroValue(simpleType)

	// Check if it returns the correct type
	if _, ok := simpleZero.(SimpleStruct); !ok {
		t.Errorf("Expected SimpleStruct, got %T", simpleZero)
	}

	// Check if it's a zero value
	simpleVal, _ := simpleZero.(SimpleStruct)
	if simpleVal.Field1 != "" || simpleVal.Field2 != 0 || simpleVal.Field3 != false {
		t.Errorf("Expected zero struct, got %+v", simpleVal)
	}

	// Test with more complex struct
	complexType := reflect.TypeOf(ComplexStruct{})
	complexZero := schema2.GetZeroValue(complexType)

	// Check type
	if _, ok := complexZero.(ComplexStruct); !ok {
		t.Errorf("Expected ComplexStruct, got %T", complexZero)
	}

	// Check zero values
	complexVal, _ := complexZero.(ComplexStruct)
	if complexVal.Name != "" || complexVal.Value != 0 ||
		complexVal.Enabled != false || complexVal.SubItems != nil {
		t.Errorf("Expected zero struct, got %+v", complexVal)
	}

	// Now test with a struct in schema
	type StructFieldStruct struct {
		ID      primitive.ObjectID `bson:"_id,omitempty"`
		Simple  SimpleStruct       `schema:"required"`
		Complex ComplexStruct
	}

	schema := schema2.GenerateFromStruct(StructFieldStruct{})

	// Check fields exist
	simpleField, exists := schema.Fields["Simple"]
	if !exists {
		t.Fatal("Expected Simple field to exist in schema")
	}

	complexField, exists := schema.Fields["Complex"]
	if !exists {
		t.Fatal("Expected Complex field to exist in schema")
	}

	// Check field types
	if _, ok := simpleField.Type.(SimpleStruct); !ok {
		t.Errorf("Expected Simple.Type to be SimpleStruct, got %T", simpleField.Type)
	}

	if _, ok := complexField.Type.(ComplexStruct); !ok {
		t.Errorf("Expected Complex.Type to be ComplexStruct, got %T", complexField.Type)
	}

	// Check that it's zero values
	simpleStructVal, _ := simpleField.Type.(SimpleStruct)
	if simpleStructVal.Field1 != "" || simpleStructVal.Field2 != 0 || simpleStructVal.Field3 != false {
		t.Errorf("Expected zero struct for Simple field, got %+v", simpleStructVal)
	}

	// Check Required flag is preserved
	if !simpleField.Required {
		t.Error("Expected Simple field to be required")
	}

	if complexField.Required {
		t.Error("Expected Complex field to not be required")
	}
}
