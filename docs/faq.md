# Frequently Asked Questions

This document addresses common questions and issues that users might encounter when using Merhongo.

## General Questions

### What is Merhongo?

Merhongo is a MongoDB library for Go that provides an intuitive API inspired by Mongoose (the popular MongoDB library for Node.js). It offers schema validation, middleware support, transaction handling, and a fluent query builder to simplify working with MongoDB in Go applications.

### What are the main benefits of using Merhongo over the official MongoDB driver?

Merhongo builds on top of the official MongoDB Go driver and adds several benefits:

- Schema definition and validation
- Schema generation from struct definitions
- Middleware support for pre/post operation hooks
- Intuitive model-based approach to collections
- Fluent query builder for constructing MongoDB queries
- Standardized error handling
- Automatic document timestamps
- Type-safe generic models with Go's generics

### What versions of Go and MongoDB are supported?

Merhongo requires:
- Go 1.18 or later (for generics support)
- MongoDB 4.0 or later (for transaction support)
- MongoDB Go Driver 1.17.3 or later

## Schema Management

### How do I create a schema for my MongoDB collection?

You have two options:

**Option 1: Manual Schema Definition**
```go
userSchema := merhongo.SchemaNew(
    map[string]schema.Field{
        "Username": {Required: true, Unique: true},
        "Email":    {Required: true, Unique: true},
        "Age":      {Min: 18, Max: 100},
    },
    schema.WithCollection("users"),
    schema.WithTimestamps(true),
)
```

**Option 2: Generate Schema from Struct (New!)**
```go
// Define struct with schema tags
type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Username  string             `schema:"required,unique"`
    Email     string             `schema:"required,unique"`
    Age       int                `schema:"min=18,max=100"`
}

// Generate schema from struct
userSchema := schema.GenerateFromStruct(User{},
    schema.WithCollection("users"),
    schema.WithTimestamps(true),
)
```

### What schema tags are supported when generating schemas from structs?

The following tags are supported:
- `required` - Field is required
- `unique` - Field must be unique (creates index)
- `min=X` - Minimum value for numbers
- `max=X` - Maximum value for numbers

Example:
```go
type User struct {
    Username  string  `schema:"required,unique"`
    Age       int     `schema:"min=18,max=100"`
}
```

### How do I handle embedded structs when generating schemas?

Embedded anonymous structs are automatically included in the generated schema:

```go
type Address struct {
    Street  string `schema:"required"`
    City    string `schema:"required"`
}

type User struct {
    ID      primitive.ObjectID `bson:"_id,omitempty"`
    Name    string             `schema:"required"`
    Address                    // Embedded struct - fields will be included
}

// The generated schema will include Street and City fields
```

## Connection Management

### How do I connect to MongoDB with authentication?

```go
// Connect with authentication
client, err := merhongo.Connect("mongodb://username:password@localhost:27017/mydatabase", "mydatabase")
```

### How do I use MongoDB connection options?

For advanced connection options, you can use the MongoDB connection string URI format:

```go
// Connect with options
uri := "mongodb://localhost:27017/mydatabase?retryWrites=true&w=majority&maxPoolSize=20"
client, err := merhongo.Connect(uri, "mydatabase")
```

### How do I manage multiple database connections?

```go
// Create multiple named connections
mainClient, err := merhongo.ConnectWithName("main", "mongodb://localhost:27017", "main_db")
analyticsClient, err := merhongo.ConnectWithName("analytics", "mongodb://localhost:27018", "analytics_db")

// Get connections by name
mainConnection := merhongo.GetConnectionByName("main")
analyticsConnection := merhongo.GetConnectionByName("analytics")

// Disconnect all connections
merhongo.DisconnectAll()
```

## Models and Schemas

### How do I create a model with indices?

```go
// Define schema with unique fields (creates indices automatically)
userSchema := merhongo.SchemaNew(
    map[string]schema.Field{
        "Username": {Required: true, Unique: true}, // Creates unique index
        "Email":    {Required: true, Unique: true}, // Creates unique index
    },
)

// Or with schema generation:
type User struct {
    Username string `schema:"required,unique"` // Creates unique index
    Email    string `schema:"required,unique"` // Creates unique index
}
userSchema := schema.GenerateFromStruct(User{})

// Create model
userModel := merhongo.ModelNew[User]("User", userSchema)
```

### How do I handle array fields in my schema?

MongoDB supports array fields, but Merhongo doesn't have specific validation for array elements. You can validate arrays in a custom validator:

```go
type Post struct {
    ID      primitive.ObjectID `bson:"_id,omitempty"`
    Title   string             `bson:"title" schema:"required"`
    Tags    []string           `bson:"tags"`
}

// Generate base schema
postSchema := schema.GenerateFromStruct(Post{})

// Add custom validator
postSchema.CustomValidator = func(doc interface{}) error {
    post, ok := doc.(*Post)
    if !ok {
        return errors.WithDetails(errors.ErrValidation, "invalid document type")
    }
    
    // Validate tags array
    if len(post.Tags) > 10 {
        return errors.WithDetails(errors.ErrValidation, "too many tags (max 10)")
    }
    
    for _, tag := range post.Tags {
        if len(tag) < 2 {
            return errors.WithDetails(errors.ErrValidation, "tag too short (min 2 chars)")
        }
    }
    
    return nil
}
```

### How do I implement soft delete with Merhongo?

You can implement soft delete using a middleware and query builder:

```go
// Add a field to your model
type User struct {
    // ... other fields
    DeletedAt *time.Time `bson:"deletedAt,omitempty"`
}

// Add middleware to filter out deleted documents
userSchema.Pre("find", func(doc interface{}) error {
    // Add deleted filter to all queries
    if filter, ok := doc.(bson.M); ok {
        filter["deletedAt"] = nil
    }
    return nil
})

// Soft delete method
func SoftDelete(ctx context.Context, userModel *merhongo.GenericModel[User], id string) error {
    now := time.Now()
    return userModel.UpdateById(ctx, id, bson.M{"deletedAt": now})
}

// Find only including deleted documents
func FindWithDeleted(ctx context.Context, userModel *merhongo.GenericModel[User]) ([]User, error) {
    // Create query without the deleted filter
    q := merhongo.QueryNew()
    // No need to add any conditions - will return all documents
    
    return userModel.FindWithQuery(ctx, q)
}
```

## Query Building

### How do I perform complex queries with logical operators (AND, OR)?

For complex queries with logical operators, you can use the `MergeFilter` method with MongoDB's `$and` and `$or` operators:

```go
import (
    "go.mongodb.org/mongo-driver/bson"
)

// Create a base query
q := query.New().Where("status", "active")

// Add complex OR condition
orCondition := bson.M{"$or": []bson.M{
    {"age": bson.M{"$lt": 18}},
    {"role": "junior"},
}}

// Merge with base query (implicitly AND)
q.MergeFilter(orCondition)

// Execute the query
users, err := userModel.FindWithQuery(ctx, q)
```

The resulting query will find documents where status is "active" AND (age < 18 OR role is "junior").

### How do I use the query builder with pagination?

```go
func GetPaginatedUsers(page, pageSize int) ([]User, int64, error) {
    ctx := context.Background()
    
    // Create query with pagination
    q := query.New().
        Where("active", true).
        SortBy("createdAt", false).
        Skip(int64((page - 1) * pageSize)).
        Limit(int64(pageSize))
    
    // Get total count (without pagination)
    countQuery := query.New().Where("active", true)
    total, err := userModel.CountWithQuery(ctx, countQuery)
    if err != nil {
        return nil, 0, err
    }
    
    // Get paginated results
    users, err := userModel.FindWithQuery(ctx, q)
    if err != nil {
        return nil, 0, err
    }
    
    return users, total, nil
}
```

## Error Handling

### How do I handle "document not found" errors properly?

```go
import (
    "github.com/isimtekin/merhongo/errors"
)

// Find a document
user, err := userModel.FindById(ctx, id)
if err != nil {
    if errors.IsNotFound(err) {
        // Handle not found case
        return nil, fmt.Errorf("user not found: %s", id)
    }
    // Handle other errors
    return nil, fmt.Errorf("database error: %v", err)
}
```

### How can I handle duplicate key errors?

When using unique fields, MongoDB will return a duplicate key error if you try to insert a document with a value that already exists:

```go
err := userModel.Create(ctx, user)
if err != nil {
    // Check for specific error message patterns
    errMsg := err.Error()
    if strings.Contains(errMsg, "duplicate key error") {
        if strings.Contains(errMsg, "username") {
            return errors.WithDetails(errors.ErrValidation, "username already taken")
        } else if strings.Contains(errMsg, "email") {
            return errors.WithDetails(errors.ErrValidation, "email already registered")
        }
        return errors.WithDetails(errors.ErrValidation, "duplicate value for unique field")
    }
    return err
}
```

## Performance and Best Practices

### How can I optimize Merhongo for performance?

1. **Use projection**: Limit fields returned by queries
```go
   q := query.New().Where("active", true)
   opts := options.Find().SetProjection(bson.M{"username": 1, "email": 1})
   ```

2. **Create indexes for frequent queries**:
```go
   indexModel := mongo.IndexModel{
       Keys: bson.D{{"createdAt", -1}},
   }
   _, err := userModel.Collection.Indexes().CreateOne(ctx, indexModel)
   ```

3. **Batch operations for bulk changes**:
```go
   models := []mongo.WriteModel{
       mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": id1}).SetUpdate(bson.M{"$set": bson.M{"status": "active"}}),
       mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": id2}).SetUpdate(bson.M{"$set": bson.M{"status": "active"}}),
   }
   result, err := userModel.Collection.BulkWrite(ctx, models)
   ```

4. **Use transactions judiciously**: Only use transactions when atomicity is required

### What are some common pitfalls to avoid?

1. **Not handling errors properly**: Always check and handle errors from database operations
2. **Not validating input data**: Ensure proper validation before database operations
3. **Using transactions for single operations**: Transactions add overhead
4. **Not using indexes for frequently queried fields**: This can lead to slow queries
5. **Not closing connections**: Always defer disconnect calls
6. **Using synchronous operations in high-concurrency environments**: Consider using workers or channels

### How can I optimize schema generation from structs?

1. **Be specific with schema tags**: Only add tags where validation is needed
2. **Reuse embedded structs**: Common fields can be grouped into embedded structs
3. **Add custom validators**: For complex validation scenarios
4. **Customize after generation**: Add default values, enums, and custom validators after generation