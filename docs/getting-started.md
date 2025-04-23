# Getting Started with Merhongo

Merhongo is a MongoDB library for Go that provides an intuitive API inspired by Mongoose. This guide will help you get started with Merhongo in your Go applications.

## Installation

To install Merhongo, use the following command:

```bash
go get github.com/isimtekin/merhongo
```

## Basic Usage

### Connecting to MongoDB

First, establish a connection to your MongoDB instance:

```go
package main

import (
	"log"
	"github.com/isimtekin/merhongo"
)

func main() {
	// Connect to MongoDB
	client, err := merhongo.Connect("mongodb://localhost:27017", "mydatabase")
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Always disconnect when done
	defer merhongo.Disconnect()

	// Use the client...
}
```

You can also create multiple named connections:

```go
// Create a primary connection
client1, err := merhongo.ConnectWithName("primary", "mongodb://localhost:27017", "primary_db")
if err != nil {
log.Fatalf("Failed to connect to primary DB: %v", err)
}

// Create a secondary connection
client2, err := merhongo.ConnectWithName("analytics", "mongodb://localhost:27018", "analytics_db")
if err != nil {
log.Fatalf("Failed to connect to analytics DB: %v", err)
}

// Get a specific connection later
analyticsClient := merhongo.GetConnectionByName("analytics")
```

### Defining Schemas

You have two ways to define schemas in Merhongo:

#### Option 1: Manual Schema Definition

Manually define the structure and validation rules:

```go
import (
    "github.com/isimtekin/merhongo"
    "github.com/isimtekin/merhongo/schema"
)

// Define a schema for users
userSchema := merhongo.SchemaNew(
    map[string]schema.Field{
        "Username": {Required: true, Unique: true},
        "Email":    {Required: true, Unique: true},
        "Age":      {Min: 18, Max: 100},
        "Role":     {Enum: []interface{}{"user", "admin", "guest"}},
    },
    schema.WithCollection("users"),     // Specify collection name
    schema.WithTimestamps(true),        // Enable automatic timestamps
)
```

#### Option 2: Generate Schema from Struct (New!)

Alternatively, you can generate a schema directly from your struct using tags:

```go
// Define your model struct with schema tags
type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Username  string             `bson:"username" json:"username" schema:"required,unique"`
    Email     string             `bson:"email" json:"email" schema:"required,unique"`
    Age       int                `bson:"age" json:"age" schema:"min=18,max=100"`
    Role      string             `bson:"role" json:"role" schema:"required"`
    CreatedAt time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
    UpdatedAt time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`
}

// Generate schema from struct
userSchema := schema.GenerateFromStruct(User{},
    schema.WithCollection("users"),
    schema.WithTimestamps(true),
)

// Optionally add more configurations after generation
userSchema.Fields["Role"].Enum = []interface{}{"user", "admin", "guest"}
userSchema.Fields["Role"].Default = "user"
```

### Creating Models

Models represent MongoDB collections with defined schemas:

```go
import (
    "go.mongodb.org/mongo-driver/bson/primitive"
    "time"
)

// Define your model struct (if not already defined)
type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Username  string             `bson:"username" json:"username"`
    Email     string             `bson:"email" json:"email"`
    Age       int                `bson:"age" json:"age"`
    Role      string             `bson:"role" json:"role"`
    CreatedAt time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
    UpdatedAt time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`
}

// Create a type-safe model with generics
userModel := merhongo.ModelNew[User]("User", userSchema)
```

### CRUD Operations

Merhongo provides intuitive CRUD operations for working with documents:

```go
import (
    "context"
    "fmt"
)

func main() {
    // ... connection code ...
    
    ctx := context.Background()
    
    // Create a new user
    user := &User{
        Username: "johndoe",
        Email:    "john@example.com",
        Age:      30,
        Role:     "user",
    }
    
    err := userModel.Create(ctx, user)
    if err != nil {
        log.Fatalf("Failed to create user: %v", err)
    }
    
    fmt.Printf("Created user with ID: %s\n", user.ID.Hex())
    
    // Find a user by ID
    foundUser, err := userModel.FindById(ctx, user.ID.Hex())
    if err != nil {
        log.Fatalf("Failed to find user: %v", err)
    }
    
    // Find users with filter
    activeUsers, err := userModel.Find(ctx, map[string]interface{}{"role": "user"})
    if err != nil {
        log.Fatalf("Failed to find users: %v", err)
    }
    
    fmt.Printf("Found %d active users\n", len(activeUsers))
    
    // Update a user
    err = userModel.UpdateById(ctx, user.ID.Hex(), map[string]interface{}{
        "age": 31,
    })
    if err != nil {
        log.Fatalf("Failed to update user: %v", err)
    }
    
    // Delete a user
    err = userModel.DeleteById(ctx, user.ID.Hex())
    if err != nil {
        log.Fatalf("Failed to delete user: %v", err)
    }
}
```

### Using the Query Builder

Merhongo provides a fluent query builder for more complex queries:

```go
// Create a query
query := merhongo.QueryNew().
    Where("age", 30).
    GreaterThan("createdAt", lastWeek).
    In("role", []string{"user", "admin"}).
    SortBy("username", true). // true = ascending
    Limit(10).
    Skip(20)

// Find users using the query
users, err := userModel.FindWithQuery(ctx, query)
if err != nil {
    log.Fatalf("Query failed: %v", err)
}

fmt.Printf("Found %d users\n", len(users))
```

## Complete Example

Here's a complete example showing how to use Merhongo with schema generation from struct:

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

// Define your model struct with schema tags
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

	// Add default value and enum for role
	userSchema.Fields["Role"].Default = "user"
	userSchema.Fields["Role"].Enum = []interface{}{"user", "admin", "guest"}

	// Create a type-safe model
	userModel := merhongo.ModelNew[User]("User", userSchema)

	// Create context
	ctx := context.Background()

	// Create a new user
	user := &User{
		Username: "johndoe",
		Email:    "john@example.com",
		Age:      30,
		Role:     "user",
		IsActive: true,
	}

	// Save to database
	err = userModel.Create(ctx, user)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("Created user with ID: %s\n", user.ID.Hex())

	// Query using the builder
	query := merhongo.QueryNew().
		Where("isActive", true).
		SortBy("username", true)

	// Find active users
	activeUsers, err := userModel.FindWithQuery(ctx, query)
	if err != nil {
		log.Fatalf("Failed to query users: %v", err)
	}

	fmt.Printf("Found %d active users\n", len(activeUsers))

	// Print user details
	for _, u := range activeUsers {
		fmt.Printf("- %s (%s)\n", u.Username, u.Email)
	}
}
```

## Next Steps

Now that you understand the basics of Merhongo, you can explore more advanced features:

- [Query Building](query-building.md) - Learn how to build complex MongoDB queries
- [Schema Validation](schema-validation.md) - Define complex validation rules
- [Schema from Struct](schema-from-struct.md) - Generate schemas from struct definitions
- [Middleware](middleware.md) - Add hooks to execute code before/after operations
- [Error Handling](error-handling.md) - Learn about Merhongo's error handling system
- [Transactions](transactions.md) - Perform multiple operations atomically