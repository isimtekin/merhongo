# Merhongo

<div align="center">
  <img src="./merhongo-logo.svg" alt="Merhongo Logo" />
</div>
<div align="center">


[![Go Reference](https://pkg.go.dev/badge/github.com/isimtekin/merhongo.svg)](https://pkg.go.dev/github.com/isimtekin/merhongo)
[![Go Report Card](https://goreportcard.com/badge/github.com/isimtekin/merhongo)](https://goreportcard.com/report/github.com/isimtekin/merhongo)
[![Test Coverage](https://img.shields.io/badge/coverage-85%25-brightgreen)](https://github.com/isimtekin/merhongo)
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
✅ **Connection Management**: Singleton pattern with support for multiple named connections  
✅ **High Test Coverage**: 85% of code covered by tests  
✅ **Comprehensive Documentation**: Detailed examples and guides

## Requirements

- **Go**: Version 1.23.5 or higher
- **MongoDB**: Compatible with MongoDB 4.4 or higher
- **MongoDB Go Driver**: Version 1.17.3 (managed automatically via go modules)

## Installation

```bash
go get github.com/isimtekin/merhongo
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/isimtekin/merhongo"
	"github.com/isimtekin/merhongo/errors"
	"github.com/isimtekin/merhongo/schema"
)

// Define your document structure
type User struct {
	ID        interface{} `bson:"_id,omitempty"`
	Username  string      `bson:"username"`
	Email     string      `bson:"email"`
	Age       int         `bson:"age"`
	CreatedAt time.Time   `bson:"createdAt"`
	UpdatedAt time.Time   `bson:"updatedAt"`
}

func main() {
	// Connect to MongoDB
	client, err := merhongo.Connect("mongodb://localhost:27017", "my_database")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer merhongo.Disconnect()

	// Define a schema
	userSchema := merhongo.SchemaNew(
		map[string]schema.Field{
			"Username": {
				Type:     "",
				Required: true,
				Unique:   true,
			},
			"Email": {
				Type:     "",
				Required: true,
			},
			"Age": {
				Type: 0,
				Min:  18,
			},
		},
		schema.WithCollection("users"),
	)

	// Create a model
	userModel := merhongo.ModelNew("User", userSchema, client.Database)

	// Create a new user
	ctx := context.Background()
	user := &User{
		Username: "johndoe",
		Email:    "john@example.com",
		Age:      30,
	}

	err = userModel.Create(ctx, user)
	if err != nil {
		// Use the error helpers for more specific error handling
		if errors.IsValidationError(err) {
			log.Fatalf("Validation error: %v", err)
		}
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("Created user: %+v\n", user)
	
	// Find a user by ID
	var foundUser User
	err = userModel.FindById(ctx, user.ID.(string), &foundUser)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Println("User not found")
		} else {
			log.Fatalf("Error finding user: %v", err)
		}
	}
}
```

## Using the Connection Manager

Merhongo provides a powerful connection management system:

```go
// Default connection
client, err := merhongo.Connect("mongodb://localhost:27017", "my_database")
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer merhongo.Disconnect()

// Access the default connection anywhere in your code
defaultClient := merhongo.GetConnection()

// Multiple named connections
usersClient, err := merhongo.ConnectWithName("users", "mongodb://localhost:27017", "users_db")
if err != nil {
    log.Fatalf("Failed to connect to users DB: %v", err)
}

productsClient, err := merhongo.ConnectWithName("products", "mongodb://localhost:27017", "products_db")
if err != nil {
    log.Fatalf("Failed to connect to products DB: %v", err)
}

// Disconnect all connections at once
defer merhongo.DisconnectAll()

// Access named connections anywhere in your code
usersConnection := merhongo.GetConnectionByName("users")
productsConnection := merhongo.GetConnectionByName("products")
```

## Using the Query Builder

Merhongo provides a powerful query builder for creating MongoDB queries:

```go
import (
	"context"
	"fmt"
	"github.com/isimtekin/merhongo"
)

func findActiveUsers(ctx context.Context, userModel *model.Model) {
	// Create a query using the builder
	q := merhongo.QueryNew().
		Where("active", true).
		GreaterThan("age", 18).
		In("role", []string{"user", "admin"}).
		SortBy("username", true).
		Limit(10)
	
	// Find documents with the query
	var users []User
	err := userModel.FindWithQuery(ctx, q, &users)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Found %d users\n", len(users))
	
	// Count documents with the query
	count, err := userModel.CountWithQuery(ctx, q)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Total count: %d\n", count)
	
	// Advanced query with regex
	q = merhongo.QueryNew().
		Regex("username", "^j", "i").  // usernames starting with 'j', case insensitive
		GreaterThanOrEqual("age", 21).
		Exists("email", true).         // must have email field
		SortBy("createdAt", false)     // newest first
	
	err = userModel.FindWithQuery(ctx, q, &users)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	// Update all matched documents
	modifiedCount, err := userModel.UpdateWithQuery(
		ctx, 
		merhongo.QueryNew().Where("active", false),
		map[string]interface{}{"active": true, "updatedAt": time.Now()},
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Activated %d users\n", modifiedCount)
}
```

## Error Handling

Merhongo provides a robust error handling system with standard error types and helper functions:

```go
import "github.com/isimtekin/merhongo/errors"

// Check specific error types
if errors.IsNotFound(err) {
    // Handle document not found
}

// Get structured error responses (useful for HTTP APIs)
response := errors.ToErrorResponse(err)
```

See the [error handling documentation](docs/error-handling.md) for more details.

## Test Coverage

Merhongo is thoroughly tested to ensure reliability:

| Package     | Coverage |
|-------------|----------|
| merhongo    | 100% |
| connection  | 100% |
| model       | 89% |
| schema      | 49% |
| query       | 85% |
| errors      | 100% |
| **Overall** | **84%** |

The high test coverage helps ensure that Merhongo is stable and reliable for production use.

## Development and Testing

### Setting up MongoDB with Docker Compose

The easiest way to set up MongoDB for development and testing is using Docker Compose:

```bash
# Start MongoDB and Mongo Express
docker-compose up -d

# Check the status of the containers
docker-compose ps

# Access Mongo Express web interface (for database management)
# Open http://localhost:8081 in your browser (default credentials: admin/pass)
```

### Using the Makefile

The project includes a Makefile to help with common development tasks:

```bash
# Run all tests
make test

# Run tests with coverage and see coverage statistics
make cover

# Generate HTML coverage report and open in browser
make cover-html

# Format Go code
make fmt

# Check if code is properly formatted (useful for CI/CD)
make check-fmt

# Run linter (requires golangci-lint to be installed)
make lint

# MongoDB container operations
make mongo-start    # Start MongoDB
make mongo-stop     # Stop MongoDB
make mongo-restart  # Restart MongoDB
make mongo-logs     # View MongoDB logs

# Docker Compose operations
make docker-up      # Start all services
make docker-down    # Stop all services
make docker-logs    # View all logs

# Clean up test artifacts
make clean
```

## Project Structure

```
merhongo/
├── merhongo.go          # Top-level package with singleton access
├── connection/          # MongoDB connection management
├── schema/              # Schema definition and validation
├── model/               # Models and CRUD operations
├── query/               # Query building utilities
├── errors/              # Error handling system
├── example/             # Example applications
├── docs/                # Documentation
├── docker-compose.yml   # Docker Compose configuration
├── Makefile             # Build and test automation
├── go.mod               # Go module definition
└── README.md            # This file
```

## Comparison with Other Libraries

| Feature                 | Merhongo | mgo | mongo-go-driver | mgm   |
|-------------------------|----------|-----|-----------------|-------|
| Schema Validation       | ✅       | ❌   | ❌               | ⚠️ Limited |
| Middleware Support      | ✅       | ❌   | ❌               | ✅     |
| Query Builder           | ✅       | ❌   | ❌               | ❌     |
| Error Types             | ✅       | ⚠️ Limited | ⚠️ Limited    | ❌     |
| Automatic Timestamps    | ✅       | ❌   | ❌               | ✅     |
| Connection Management   | ✅       | ⚠️ Limited | ⚠️ Limited    | ❌     |
| Active Development      | ✅       | ❌   | ✅               | ⚠️ Limited |

## Roadmap

- **v0.2.0**: ✅ Connection management with singleton pattern, enhanced structure
- **v0.3.0**: Aggregation pipeline builder
- **v0.4.0**: Population (references) support
- **v0.5.0**: Transactions and bulk operations helpers
- **v1.0.0**: API stabilization and performance optimizations

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgements

- The [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver) team
- The [Mongoose](https://mongoosejs.com/) project for inspiration
