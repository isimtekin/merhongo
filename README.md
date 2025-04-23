# Merhongo

<div align="center">
  <img src="./media/merhongo-logo.png" alt="Merhongo Logo" />
</div>
<div align="center">

[![Go Reference](https://pkg.go.dev/badge/github.com/isimtekin/merhongo.svg)](https://pkg.go.dev/github.com/isimtekin/merhongo)
[![Go Report Card](https://goreportcard.com/badge/github.com/isimtekin/merhongo)](https://goreportcard.com/report/github.com/isimtekin/merhongo)
[![Test Coverage](https://img.shields.io/badge/coverage-84-brightgreen)](https://github.com/isimtekin/merhongo)
[![CI](https://github.com/isimtekin/merhongo/actions/workflows/ci.yml/badge.svg)](https://github.com/isimtekin/merhongo/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/isimtekin/merhongo/branch/main/graph/badge.svg)](https://codecov.io/gh/isimtekin/merhongo)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/isimtekin/merhongo/blob/main/LICENSE)

**A simple, robust, and intuitive MongoDB driver for Go, inspired by Mongoose.**

</div>

## Overview

Merhongo combines the power of the official MongoDB Go driver with an intuitive API inspired by Mongoose. It provides schema validation, middleware support, elegant query building, robust error handling, and much more to simplify working with MongoDB in your Go applications.

## Features

✅ **Schema Definition & Validation**: Define document structure with validation rules  
✅ **Intuitive Models**: Create clean, reusable models for collections  
✅ **Powerful Query Builder**: Build MongoDB queries with a fluent, chainable API  
✅ **Middleware Support**: Pre/post operation hooks for advanced workflows  
✅ **Robust Error Handling**: Standardized error types and helpful utilities  
✅ **Automatic Timestamps**: Built-in createdAt/updatedAt field management  
✅ **Type-Safe Generic Models**: Create type-safe models with Go's generics  
✅ **Flexible Configuration**: Configure models with rich options API  
✅ **Auto-Generate Schemas**: Generate schemas directly from struct definitions  
✅ **High Test Coverage**: 84% of code covered by tests  
✅ **Comprehensive Documentation**: Detailed examples and guides

## Requirements

- **Go**: Version 1.18 or higher (for generics support)
- **MongoDB**: Compatible with MongoDB 4.0 or higher (for transaction support)
- **MongoDB Go Driver**: Version 1.17.3 or higher (managed automatically via go modules)

## Installation

```bash
go get github.com/isimtekin/merhongo
```

## Documentation

Comprehensive documentation is available in the [docs](./docs) directory:

- [Getting Started](./docs/getting-started.md) - Basic usage and examples
- [Query Building](./docs/query-building.md) - Building complex MongoDB queries
- [Schema Validation](./docs/schema-validation.md) - Defining validation rules
- [Schema from Struct](./docs/schema-from-struct.md) - Generating schemas from structs
- [Middleware](./docs/middleware.md) - Adding hooks for operations
- [Error Handling](./docs/error-handling.md) - Working with Merhongo errors
- [Transactions](./docs/transactions.md) - Using MongoDB transactions
- [API Reference](./docs/api-reference.md) - Detailed API documentation
- [FAQ](./docs/faq.md) - Frequently asked questions

## Quick Start

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

	// Create a type-safe model with generics
	userModel := merhongo.ModelNew[User]("User", userSchema)

	// Create a document
	user := &User{
		Username: "john_doe",
		Email:    "john@example.com",
		Age:      30,
	}

	ctx := context.Background()
	err = userModel.Create(ctx, user)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("Created user with ID: %s\n", user.ID.Hex())

	// Query with builder
	queryBuilder := merhongo.QueryNew().
		Where("username", "john_doe").
		SortBy("createdAt", false)

	foundUser, err := userModel.FindOneWithQuery(ctx, queryBuilder)
	if err != nil {
		log.Fatalf("Failed to find user: %v", err)
	}

	fmt.Printf("Found user: %s (%s)\n", foundUser.Username, foundUser.Email)
}
```

## Advanced Usage Examples

### Creating Models with Options

```go
// Define model with options
userModel := merhongo.ModelNew[User]("User", userSchema, merhongo.ModelOptions{
    ConnectionName: "analytics",      // Use a specific named connection
    AutoCreateIndexes: true,          // Automatically create indexes
    CustomValidator: func(doc interface{}) error {
        // Custom validation logic
        return nil
    },
})
```

### Using Middleware

```go
// Add pre-save middleware
userSchema.Pre("save", func(doc interface{}) error {
    // Cast to your type if needed
    if user, ok := doc.(*User); ok {
        // Do something with the user before saving
        fmt.Printf("About to save user: %s\n", user.Username)
    }
    return nil
})
```

### Query Building

```go
query := merhongo.QueryNew().
    Where("age", 30).
    GreaterThan("createdAt", lastWeek).
    In("status", []string{"active", "pending"}).
    SortBy("username", true).
    Limit(10).
    Skip(20)

users, err := userModel.FindWithQuery(ctx, query)
```

### Transactions

```go
err := client.ExecuteTransaction(ctx, func(sc mongo.SessionContext) error {
    // Create a user
    err := userModel.Create(sc, user)
    if err != nil {
        return err
    }
    
    // Create a related document
    err = profileModel.Create(sc, profile)
    if err != nil {
        return err
    }
    
    return nil
})
```

## Error Handling

```go
err := userModel.FindById(ctx, "invalid-id", &user)
if merhongo.errors.IsInvalidObjectID(err) {
    // Handle invalid ID error
}

err = userModel.FindById(ctx, validId, &user)
if merhongo.errors.IsNotFound(err) {
    // Handle not found error
}
```

## Test Coverage

Merhongo is thoroughly tested to ensure reliability:

| Package     | Coverage |
|-------------|----------|
| merhongo    | 89%      |
| connection  | 100%     |
| model       | 89%      |
| schema      | 84%      |
| query       | 85%      |
| errors      | 100%     |
| **Overall** | **84%**  |

## Comparison with Other Libraries

| Feature                 | Merhongo | mgo | mongo-go-driver | mgm   |
|-------------------------|----------|-----|-----------------|-------|
| Schema Validation       | ✅       | ❌   | ❌               | ⚠️ Limited |
| Schema from Struct      | ✅       | ❌   | ❌               | ❌     |
| Middleware Support      | ✅       | ❌   | ❌               | ✅     |
| Query Builder           | ✅       | ❌   | ❌               | ❌     |
| Error Types             | ✅       | ⚠️ Limited | ⚠️ Limited    | ❌     |
| Type-Safe Generic Models| ✅       | ❌   | ❌               | ❌     |
| Automatic Timestamps    | ✅       | ❌   | ❌               | ✅     |
| Active Development      | ✅       | ❌   | ✅               | ⚠️ Limited |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.