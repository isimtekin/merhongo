# Generating Schemas from Structs

Merhongo now supports automatically generating schemas from Go struct definitions using struct tags. This feature helps reduce code duplication and keeps your schema definitions in sync with your struct types.

## Basic Usage

Instead of manually defining a schema, you can now generate one directly from your struct:

```go
import (
    "github.com/isimtekin/merhongo"
    "github.com/isimtekin/merhongo/schema"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "time"
)

// Define your struct with schema tags
type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Username  string             `bson:"username" json:"username" schema:"required,unique"`
    Email     string             `bson:"email" json:"email" schema:"required,unique"`
    Age       int                `bson:"age" json:"age" schema:"min=18,max=100"`
    Role      string             `bson:"role" json:"role" schema:"required"`
    IsActive  bool               `bson:"isActive" json:"isActive"`
    CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
    UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Generate schema from struct
userSchema := schema.GenerateFromStruct(User{},
    schema.WithCollection("users"),
    schema.WithTimestamps(true),
)

// Create model
userModel := merhongo.ModelNew[User]("User", userSchema)
```

## Schema Tags

Add validation rules directly to your structs using the `schema` tag:

| Tag | Example | Description |
|-----|---------|-------------|
| `required` | `schema:"required"` | Field is required |
| `unique` | `schema:"unique"` | Field must be unique (creates index) |
| `min` | `schema:"min=18"` | Minimum value for numbers |
| `max` | `schema:"max=100"` | Maximum value for numbers |

You can combine multiple tags by separating them with commas:

```go
// Field is required, unique, and has min/max constraints
`schema:"required,unique,min=18,max=100"`
```

## Customizing Generated Schemas

After generating a schema, you can customize it further:

```go
// Generate base schema
userSchema := schema.GenerateFromStruct(User{},
    schema.WithCollection("users"),
    schema.WithTimestamps(true),
)

// Add default values
userSchema.Fields["Role"].Default = "user"
userSchema.Fields["IsActive"].Default = true

// Add enum values
userSchema.Fields["Role"].Enum = []interface{}{"user", "admin", "guest"}

// Add custom validation function
userSchema.Fields["Email"].ValidateFunc = func(val interface{}) bool {
    email, ok := val.(string)
    if !ok {
        return false
    }
    return strings.Contains(email, "@")
}

// Add middleware
userSchema.Pre("save", func(doc interface{}) error {
    // Pre-save logic here
    return nil
})
```

## Embedding Struct Fields

The schema generator supports anonymous embedded structs:

```go
type Address struct {
    Street  string `schema:"required"`
    City    string `schema:"required"`
    Country string `schema:"required"`
}

type Customer struct {
    ID      primitive.ObjectID `bson:"_id,omitempty"`
    Name    string             `schema:"required"`
    Email   string             `schema:"required,unique"`
    Address                    // Embedded struct - fields will be included
}

// Generate schema - will include both Customer fields and Address fields
customerSchema := schema.GenerateFromStruct(Customer{})
```

## Special Types Handling

The schema generator handles several special types:

- **Primitive Types**: int, string, bool, etc.
- **Time Types**: time.Time
- **MongoDB Types**: primitive.ObjectID
- **Custom Types**: User-defined types like enums
- **Collection Types**: Slices, maps (basic validation only)
- **Pointer Types**: Pointers to any of the above

## Example with Custom Types

```go
// Define a custom type
type UserRole string

const (
    RoleAdmin UserRole = "admin"
    RoleUser  UserRole = "user"
    RoleGuest UserRole = "guest"
)

// Use the custom type in your struct
type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Name      string             `schema:"required"`
    Role      UserRole           `schema:"required"`
}

// Generate schema
userSchema := schema.GenerateFromStruct(User{})

// Add custom type validation
userSchema.Fields["Role"].Enum = []interface{}{RoleAdmin, RoleUser, RoleGuest}
userSchema.Fields["Role"].Default = RoleUser
```

## Limitations

There are a few limitations to be aware of:

1. **Arrays and Fixed-Size Arrays**: While arrays are supported in the schema, detailed validation for array elements must be handled in a custom validator.

2. **Complex Nested Structures**: For deeply nested structures, you may need to add custom validators.

3. **Interface Types**: Interface fields are supported but require custom validation.

## Advanced: Adding Custom Struct Tag

If you need to add more validation rules via struct tags, you can modify the `parseSchemaTag` function in the schema package.

## Complete Example

Here's a complete example using schema generation from struct:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/isimtekin/merhongo"
	"github.com/isimtekin/merhongo/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Define struct with schema tags
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username" schema:"required,unique"`
	Email     string             `bson:"email" json:"email" schema:"required,unique"`
	Age       int                `bson:"age" json:"age" schema:"min=18,max=100"`
	Role      string             `bson:"role" json:"role" schema:"required"`
	IsActive  bool               `bson:"isActive" json:"isActive"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

func main() {
	// Connect to MongoDB
	client, err := merhongo.Connect("mongodb://localhost:27017", "myapp")
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer merhongo.Disconnect()

	// Generate schema from struct
	userSchema := schema.GenerateFromStruct(User{},
		schema.WithCollection("users"),
		schema.WithTimestamps(true),
	)

	// Add default value for role
	userSchema.Fields["Role"].Default = "user"
	
	// Add middleware
	userSchema.Pre("save", func(doc interface{}) error {
		if user, ok := doc.(*User); ok {
			fmt.Printf("Saving user: %s\n", user.Username)
		}
		return nil
	})

	// Create type-safe model
	userModel := merhongo.ModelNew[User]("User", userSchema)

	// Create a document
	user := &User{
		Username: "john_doe",
		Email:    "john@example.com",
		Age:      30,
		Role:     "user",
		IsActive: true,
	}

	ctx := context.Background()
	err = userModel.Create(ctx, user)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("Created user with ID: %s\n", user.ID.Hex())
}
```