# Query Building

Merhongo provides a powerful and expressive query builder that helps you construct MongoDB queries with a fluent, chainable API. This approach makes complex queries more readable and maintainable.

## Basic Query Operations

### Creating a Query Builder

```go
import (
    "github.com/isimtekin/merhongo"
    "github.com/isimtekin/merhongo/query"
)

// Create a new query builder
q := merhongo.QueryNew()
// or
q := query.New()
```

### Simple Equality Conditions

```go
// Find users with age = 30
q := query.New().Where("age", 30)

// Execute the query
users, err := userModel.FindWithQuery(ctx, q)
```

### Comparison Operators

```go
// Greater than
q := query.New().GreaterThan("age", 25)

// Less than
q := query.New().LessThan("age", 50)

// Greater than or equal
q := query.New().GreaterThanOrEqual("age", 18)

// Less than or equal
q := query.New().LessThanOrEqual("age", 65)

// Not equals
q := query.New().NotEquals("status", "inactive")
```

### Array Operators

```go
// In - match any value in an array
q := query.New().In("role", []string{"admin", "moderator"})

// Not In - exclude matches from an array
q := query.New().NotIn("status", []string{"deleted", "banned"})
```

### Logical Operators

You can combine multiple conditions which are implicitly joined with AND logic:

```go
// Find active users over 25
q := query.New().
    Where("active", true).
    GreaterThan("age", 25)
```

### Existence Check

```go
// Documents where email field exists
q := query.New().Exists("email", true)

// Documents where phone field doesn't exist
q := query.New().Exists("phone", false)
```

### Regular Expression Search

```go
// Names starting with "john", case insensitive
q := query.New().Regex("name", "^john", "i")
```

## Query Options

### Sorting

```go
// Sort by age ascending
q := query.New().SortBy("age", true)

// Sort by registration date descending
q := query.New().SortBy("createdAt", false)

// Multiple sort criteria
q := query.New().
    SortBy("role", true).      // Sort by role ascending
    SortBy("createdAt", false) // Then by createdAt descending
```

### Pagination

```go
// Limit results to 10 documents
q := query.New().Limit(10)

// Skip the first 20 documents
q := query.New().Skip(20)

// Typical pagination implementation
q := query.New().
    Skip((page - 1) * pageSize).
    Limit(pageSize)
```

## Combining Multiple Conditions

You can chain multiple conditions to create complex queries:

```go
// Find active users who are either admins or over 30 years old
q := query.New().
    Where("active", true).
    In("role", []string{"admin", "moderator"}).
    GreaterThanOrEqual("age", 30).
    Skip(0).
    Limit(10).
    SortBy("createdAt", false)
```

## Executing Queries

### Find Multiple Documents

```go
// Using the standard model
var users []User
err := userModel.FindWithQuery(ctx, q, &users)

// Using the generic model
users, err := userModel.FindWithQuery(ctx, q)
```

### Find a Single Document

```go
// Using the standard model
var user User
err := userModel.FindOneWithQuery(ctx, q, &user)

// Using the generic model
user, err := userModel.FindOneWithQuery(ctx, q)
```

### Count Documents

```go
count, err := userModel.CountWithQuery(ctx, q)
```

### Update Documents

```go
// Define the update
update := map[string]interface{}{
    "status": "active",
    "updatedBy": "system",
}

// Update all matching documents
modifiedCount, err := userModel.UpdateWithQuery(ctx, q, update)
```

### Delete Documents

```go
deletedCount, err := userModel.DeleteWithQuery(ctx, q)
```

## Error Handling

The query builder validates conditions and reports errors:

```go
q := query.New().
    Where("", "invalid") // This will create an error

// The error will be returned when you build or execute the query
_, _, err := q.Build()
if err != nil {
    // Handle validation error
}
```

## Advanced Usage: Merging Filters

You can merge existing BSON filters into your query:

```go
// Create a query
q := query.New().Where("active", true)

// Add a complex filter
additionalFilter := bson.M{"$or": []bson.M{
    {"role": "admin"},
    {"age": bson.M{"$gte": 30}},
}}

// Merge the filters
q.MergeFilter(additionalFilter)

// Execute the query
users, err := userModel.FindWithQuery(ctx, q)
```

This is particularly useful when you need to build queries dynamically based on user input or other conditions.