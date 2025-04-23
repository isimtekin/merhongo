package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/isimtekin/merhongo"
	"github.com/isimtekin/merhongo/errors"
	"github.com/isimtekin/merhongo/query"
	"github.com/isimtekin/merhongo/schema"
)

// User represents a user document in MongoDB
type User struct {
	ID        interface{} `bson:"_id,omitempty"`
	Username  string      `bson:"username" schema:"required,unique"`
	Email     string      `bson:"email" schema:"required,unique"`
	Age       int         `bson:"age" schema:"min=18"`
	Active    bool        `bson:"active"`
	CreatedAt time.Time   `bson:"createdAt,omitempty"`
	UpdatedAt time.Time   `bson:"updatedAt,omitempty"`
}

// Product represents a product document in MongoDB
type Product struct {
	ID          interface{} `bson:"_id,omitempty"`
	Name        string      `bson:"name" schema:"required"`
	Description string      `bson:"description"`
	Price       float64     `bson:"price" schema:"required,min=0"`
	InStock     bool        `bson:"inStock"`
	CreatedAt   time.Time   `bson:"createdAt,omitempty"`
	UpdatedAt   time.Time   `bson:"updatedAt,omitempty"`
}

func main() {
	// Example 1: Basic usage with default connection
	basicExample()

	// Example 2: Using multiple connections
	multiConnectionExample()

	// Example 3: Using new generic models with options
	genericModelExample()

	// Example 4: Using schema generation from struct (New!)
	schemaFromStructExample()
}

func basicExample() {
	fmt.Println("\n=== Basic Example ===")

	// Connect to MongoDB using the default connection
	client, err := merhongo.Connect("mongodb://localhost:27017", "merhongo_example")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer merhongo.Disconnect()

	// Define a schema for User
	userSchema := merhongo.SchemaNew(
		map[string]schema.Field{
			"Username": {
				Required: true,
				Unique:   true,
			},
			"Email": {
				Required: true,
				Unique:   true,
				ValidateFunc: func(value interface{}) bool {
					email, ok := value.(string)
					if !ok {
						return false
					}
					// Simple email validation
					return len(email) > 0 && contains(email, "@")
				},
			},
			"Age": {
				Min: 18,
			},
			"Active": {
				Type: true,
			},
		},
		schema.WithCollection("users"),
	)

	// Add a pre-save middleware
	userSchema.Pre("save", func(doc interface{}) error {
		user, ok := doc.(*User)
		if !ok {
			return fmt.Errorf("document is not a User")
		}

		// Set default active state
		if !user.Active {
			user.Active = true
		}

		fmt.Println("Pre-save middleware executed")
		return nil
	})

	// Create a model using the generic model
	userModel := merhongo.ModelNew[User]("User", userSchema, merhongo.ModelOptions{
		Database: client.Database,
	})

	// Create a new user
	ctx := context.Background()
	user := &User{
		Username: "johndoe",
		Email:    "john@example.com",
		Age:      30,
	}

	err = userModel.Create(ctx, user)
	if err != nil {
		if errors.IsValidationError(err) {
			log.Fatalf("Validation error: %v", err)
		}
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("Created user: %+v\n", user)

	// Find a user by ID - note the type-safe return value for generic model
	foundUser, err := userModel.FindById(ctx, user.ID.(string))
	if err != nil {
		if errors.IsNotFound(err) {
			log.Println("User not found")
		} else {
			log.Fatalf("Error finding user: %v", err)
		}
	} else {
		fmt.Printf("Found user: %+v\n", foundUser)
	}

	// Use query builder to find users with type-safe return
	q := merhongo.QueryNew().
		Where("active", true).
		GreaterThan("age", 18)

	users, err := userModel.FindWithQuery(ctx, q)
	if err != nil {
		log.Fatalf("Query error: %v", err)
	}

	fmt.Printf("Found %d active users over 18\n", len(users))

	// Update user using query
	updateCount, err := userModel.UpdateWithQuery(
		ctx,
		query.New().Where("username", "johndoe"),
		map[string]interface{}{"age": 31},
	)
	if err != nil {
		log.Fatalf("Update error: %v", err)
	}

	fmt.Printf("Updated %d users\n", updateCount)
}

func multiConnectionExample() {
	fmt.Println("\n=== Multiple Connections Example ===")

	// Connect to multiple databases
	_, err := merhongo.ConnectWithName("users", "mongodb://localhost:27017", "users_db")
	if err != nil {
		log.Fatalf("Failed to connect to users DB: %v", err)
	}

	_, err = merhongo.ConnectWithName("products", "mongodb://localhost:27017", "products_db")
	if err != nil {
		log.Fatalf("Failed to connect to products DB: %v", err)
	}
	defer merhongo.DisconnectAll()

	// Create schemas
	userSchema := merhongo.SchemaNew(
		map[string]schema.Field{
			"Username": {Required: true, Unique: true},
			"Email":    {Required: true, Unique: true},
		},
		schema.WithCollection("users"),
	)

	productSchema := merhongo.SchemaNew(
		map[string]schema.Field{
			"Name":  {Required: true},
			"Price": {Required: true, Min: 0},
		},
		schema.WithCollection("products"),
	)

	// Create models with specific connections using connection names
	userModel := merhongo.ModelNew[User]("User", userSchema, merhongo.ModelOptions{
		ConnectionName: "users",
	})

	productModel := merhongo.ModelNew[Product]("Product", productSchema, merhongo.ModelOptions{
		ConnectionName: "products",
	})

	// Use models with different connections
	ctx := context.Background()

	// Create a user in the users database
	user := &User{
		Username: "alice",
		Email:    "alice@example.com",
		Age:      25,
		Active:   true,
	}

	err = userModel.Create(ctx, user)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("Created user in users_db: %+v\n", user)

	// Create a product in the products database
	product := &Product{
		Name:        "Smartphone",
		Description: "Latest model",
		Price:       999.99,
		InStock:     true,
	}

	err = productModel.Create(ctx, product)
	if err != nil {
		log.Fatalf("Failed to create product: %v", err)
	}

	fmt.Printf("Created product in products_db: %+v\n", product)

	// Query each database - using type-safe generic methods
	users, err := userModel.Find(ctx, map[string]interface{}{})
	if err != nil {
		log.Fatalf("Error querying users: %v", err)
	}
	fmt.Printf("Found %d users in users_db\n", len(users))

	products, err := productModel.Find(ctx, map[string]interface{}{})
	if err != nil {
		log.Fatalf("Error querying products: %v", err)
	}
	fmt.Printf("Found %d products in products_db\n", len(products))

	// Get connections by name
	usersConn := merhongo.GetConnectionByName("users")
	if usersConn == nil {
		log.Fatal("Failed to get users connection")
	}
	fmt.Println("Successfully retrieved users connection")

	productsConn := merhongo.GetConnectionByName("products")
	if productsConn == nil {
		log.Fatal("Failed to get products connection")
	}
	fmt.Println("Successfully retrieved products connection")
}

// New example showcasing generic models with options
func genericModelExample() {
	fmt.Println("\n=== Generic Model Example with Options ===")

	// Connect to MongoDB
	_, err := merhongo.Connect("mongodb://localhost:27017", "merhongo_generic_example")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer merhongo.Disconnect()

	// Define schemas
	productSchema := merhongo.SchemaNew(
		map[string]schema.Field{
			"Name":  {Required: true},
			"Price": {Required: true, Min: 0},
		},
		schema.WithCollection("advanced_products"),
	)

	// Create model with type safety and custom validator
	validationCalled := false

	productModel := merhongo.ModelNew[Product]("AdvancedProduct", productSchema, merhongo.ModelOptions{
		AutoCreateIndexes: true,
		CustomValidator: func(doc interface{}) error {
			product, ok := doc.(*Product)
			if !ok {
				return fmt.Errorf("document is not a Product")
			}

			// Custom validation logic
			if product.Price <= 0 {
				return fmt.Errorf("price must be positive")
			}

			validationCalled = true
			return nil
		},
	})

	// Use the model
	ctx := context.Background()

	// Create a product with the generic model
	product := &Product{
		Name:        "Luxury Watch",
		Description: "Premium timepiece",
		Price:       1299.99,
		InStock:     true,
	}

	err = productModel.Create(ctx, product)
	if err != nil {
		log.Fatalf("Failed to create product: %v", err)
	}

	fmt.Printf("Created product with generic model: %+v\n", product)
	fmt.Printf("Custom validator was called: %v\n", validationCalled)

	// Query using the type-safe model
	products, err := productModel.FindWithQuery(
		ctx,
		query.New().GreaterThan("price", 1000),
	)
	if err != nil {
		log.Fatalf("Error querying products: %v", err)
	}

	fmt.Printf("Found %d premium products\n", len(products))
}

// New example showcasing schema generation from struct
func schemaFromStructExample() {
	fmt.Println("\n=== Schema From Struct Example ===")

	// Connect to MongoDB
	client, err := merhongo.Connect("mongodb://localhost:27017", "merhongo_schema_from_struct")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer merhongo.Disconnect()

	// Generate schema directly from User struct
	// Note: User struct has schema tags defined at the top of this file
	userSchema := schema.GenerateFromStruct(User{},
		schema.WithCollection("schema_users"),
		schema.WithTimestamps(true),
	)

	// Customize schema after generation - using correct way to update map values
	// The correct way to update a Field in a Schema is to:
	// 1. Get the current field
	// 2. Make a copy and update the copy
	// 3. Assign the updated copy back to the map
	emailField := userSchema.Fields["Email"]
	emailField.ValidateFunc = func(value interface{}) bool {
		email, ok := value.(string)
		if !ok {
			return false
		}
		// Simple email validation
		return len(email) > 0 && contains(email, "@")
	}
	userSchema.Fields["Email"] = emailField

	// Similar approach for setting default value
	activeField := userSchema.Fields["Active"]
	activeField.Default = true
	userSchema.Fields["Active"] = activeField

	// Add middleware
	userSchema.Pre("save", func(doc interface{}) error {
		fmt.Println("Pre-save middleware executed for schema generated from struct")
		return nil
	})

	// Create a model with the generated schema
	userModel := merhongo.ModelNew[User]("StructUser", userSchema, merhongo.ModelOptions{
		Database: client.Database,
	})

	// Create a new user
	ctx := context.Background()
	user := &User{
		Username: "schema_user",
		Email:    "schema@example.com",
		Age:      25,
	}

	err = userModel.Create(ctx, user)
	if err != nil {
		log.Fatalf("Failed to create user with schema from struct: %v", err)
	}

	fmt.Printf("Created user with schema from struct: %+v\n", user)

	// Generate schema from Product struct
	productSchema := schema.GenerateFromStruct(Product{},
		schema.WithCollection("schema_products"),
		schema.WithTimestamps(true),
	)

	// Create model with generated schema
	productModel := merhongo.ModelNew[Product]("StructProduct", productSchema)

	// Create a product
	product := &Product{
		Name:        "Generated Schema Product",
		Description: "Created with schema from struct",
		Price:       199.99,
		InStock:     true,
	}

	err = productModel.Create(ctx, product)
	if err != nil {
		log.Fatalf("Failed to create product with schema from struct: %v", err)
	}

	fmt.Printf("Created product with schema from struct: %+v\n", product)

	// Query using builder
	q := merhongo.QueryNew().
		Where("inStock", true).
		SortBy("price", true)

	products, err := productModel.FindWithQuery(ctx, q)
	if err != nil {
		log.Fatalf("Error querying products with schema from struct: %v", err)
	}

	fmt.Printf("Found %d in-stock products with schema from struct\n", len(products))
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}
