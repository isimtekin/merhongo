# Schema Validation

Merhongo provides a powerful schema validation system that lets you define rules for your MongoDB documents. This helps ensure data integrity and consistency in your database.

## Creating a Schema

A schema in Merhongo defines the structure of your documents and validation rules:

```go
import (
    "github.com/isimtekin/merhongo"
    "github.com/isimtekin/merhongo/schema"
)

// Create a schema for users
userSchema := merhongo.SchemaNew(
    map[string]schema.Field{
        "Username": {Required: true, Unique: true},
        "Email":    {Required: true, Unique: true},
        "Age":      {Min: 18, Max: 100},
        "Role":     {Enum: []interface{}{"user", "admin", "guest"}},
    },
    schema.WithCollection("users"),
    schema.WithTimestamps(true),
)
```

## Available Field Validation Rules

Merhongo's `Field` struct provides several validation rules:

| Rule | Type | Description |
|------|------|-------------|
| `Type` | `interface{}` | Expected type for the field |
| `Required` | `bool` | Whether the field is required |
| `Default` | `interface{}` | Default value if not provided |
| `Unique` | `bool` | Whether the field should be unique |
| `Min` | `int` | Minimum value for numbers |
| `Max` | `int` | Maximum value for numbers |
| `Enum` | `[]interface{}` | List of allowed values |
| `ValidateFunc` | `func(interface{}) bool` | Custom validation function |

## Validation Rules Examples

### Required Fields

```go
"Username": {Required: true}
```

This ensures that the `Username` field must be present and not empty in every document.

### Unique Fields

```go
"Email": {Required: true, Unique: true}
```

This ensures that the `Email` field must be present and have a unique value across all documents. Merhongo will automatically create a unique index for these fields.

### Numeric Constraints

```go
"Age": {Min: 18, Max: 100}
```

This ensures that the `Age` field must be at least 18 and at most 100.

### Enum Values

```go
"Role": {Enum: []interface{}{"user", "admin", "guest"}}
```

This ensures that the `Role` field must be one of the specified values.

### Custom Validation Function

```go
"Password": {
    Required: true,
    ValidateFunc: func(val interface{}) bool {
        password, ok := val.(string)
        if !ok {
            return false
        }
        // Password must be at least 8 characters
        return len(password) >= 8
    },
}
```

This adds a custom validation function that checks if the password is at least 8 characters long.

## Schema Options

### Collection Name

```go
schema.WithCollection("users")
```

This specifies the MongoDB collection name for this schema.

### Automatic Timestamps

```go
schema.WithTimestamps(true)
```

When enabled, Merhongo automatically sets `CreatedAt` and `UpdatedAt` fields on documents. Your model must have these fields defined:

```go
type User struct {
    // ...other fields
    CreatedAt time.Time `bson:"createdAt,omitempty"`
    UpdatedAt time.Time `bson:"updatedAt,omitempty"`
}
```

## Custom Validation

You can add custom validation logic to your schema:

```go
// Define schema
userSchema := merhongo.SchemaNew(
    map[string]schema.Field{
        "Username": {Required: true},
        "Email":    {Required: true},
        "Password": {Required: true},
    },
)

// Set custom validator
userSchema.CustomValidator = func(doc interface{}) error {
    user, ok := doc.(*User)
    if !ok {
        return errors.WithDetails(errors.ErrValidation, "invalid document type")
    }
    
    // Check password strength
    if len(user.Password) < 8 {
        return errors.WithDetails(errors.ErrValidation, "password must be at least 8 characters")
    }
    
    // Check email format
    if !strings.Contains(user.Email, "@") {
        return errors.WithDetails(errors.ErrValidation, "invalid email format")
    }
    
    return nil
}
```

## Validating Documents

Validation happens automatically when you use the model's `Create` method:

```go
user := &User{
    Username: "john_doe",
    Email:    "john@example.com",
    Age:      16, // This will trigger a validation error (min: 18)
}

err := userModel.Create(ctx, user)
if err != nil {
    if errors.IsValidationError(err) {
        fmt.Println("Validation error:", err)
        // Output: Validation error: validation failed: field 'Age' value 16 is less than minimum 18
    }
}
```

## Schema Middleware

You can add middleware functions to be executed before validating a document:

```go
// Add pre-save middleware to hash passwords
userSchema.Pre("save", func(doc interface{}) error {
    if user, ok := doc.(*User); ok {
        // Hash the password if it's not already hashed
        if !strings.HasPrefix(user.Password, "$2a$") {
            hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
            if err != nil {
                return errors.Wrap(errors.ErrMiddleware, "failed to hash password")
            }
            user.Password = string(hashedPassword)
        }
    }
    return nil
})
```

For more information on middleware, see the [Middleware](middleware.md) documentation.

## Advanced Usage: Schema with Model Options

You can create a model with additional options:

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

This allows you to override the schema's validator or use a specific database connection.