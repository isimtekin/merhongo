# Middleware

Merhongo provides middleware support that allows you to run custom functions before or after database operations. This feature is useful for implementing cross-cutting concerns like validation, logging, data transformation, and more.

## Understanding Middleware

Middleware functions are executed at specific points in the lifecycle of document operations. In Merhongo, middleware can be registered on schemas and will be applied to all operations on models using that schema.

Currently, Merhongo supports "pre" middleware for the "save" event, which runs before a document is saved to the database.

## Adding Middleware to a Schema

```go
import (
    "github.com/isimtekin/merhongo"
    "github.com/isimtekin/merhongo/schema"
    "github.com/isimtekin/merhongo/errors"
    "fmt"
    "time"
)

// Create a schema
userSchema := merhongo.SchemaNew(
    map[string]schema.Field{
        "Username": {Required: true, Unique: true},
        "Email":    {Required: true, Unique: true},
    },
)

// Add pre-save middleware
userSchema.Pre("save", func(doc interface{}) error {
    // Cast to your specific type
    if user, ok := doc.(*User); ok {
        // Implement your middleware logic
        fmt.Printf("About to save user: %s\n", user.Username)
    }
    return nil
})
```

## Common Middleware Use Cases

### Validation

```go
userSchema.Pre("save", func(doc interface{}) error {
    user, ok := doc.(*User)
    if !ok {
        return errors.WithDetails(errors.ErrMiddleware, "invalid document type")
    }
    
    // Validate email format
    if !strings.Contains(user.Email, "@") {
        return errors.WithDetails(errors.ErrValidation, "invalid email format")
    }
    
    return nil
})
```

### Password Hashing

```go
import (
    "golang.org/x/crypto/bcrypt"
)

userSchema.Pre("save", func(doc interface{}) error {
    user, ok := doc.(*User)
    if !ok {
        return errors.WithDetails(errors.ErrMiddleware, "invalid document type")
    }
    
    // Hash the password if it's not already hashed
    if user.Password != "" && !strings.HasPrefix(user.Password, "$2a$") {
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
        if err != nil {
            return errors.Wrap(errors.ErrMiddleware, "failed to hash password")
        }
        user.Password = string(hashedPassword)
    }
    
    return nil
})
```

### Automatic Timestamps

While Merhongo already provides automatic timestamp management, you can implement custom timestamp behavior:

```go
userSchema.Pre("save", func(doc interface{}) error {
    user, ok := doc.(*User)
    if !ok {
        return errors.WithDetails(errors.ErrMiddleware, "invalid document type")
    }
    
    now := time.Now()
    
    // Set timestamps
    if user.ID.IsZero() {
        // New document
        user.CreatedAt = now
    }
    user.UpdatedAt = now
    
    return nil
})
```

### Logging

```go
userSchema.Pre("save", func(doc interface{}) error {
    user, ok := doc.(*User)
    if !ok {
        return nil
    }
    
    // Log document creation/update
    if user.ID.IsZero() {
        log.Printf("Creating new user: %s", user.Username)
    } else {
        log.Printf("Updating user: %s (ID: %s)", user.Username, user.ID.Hex())
    }
    
    return nil
})
```

### Field Sanitization

```go
userSchema.Pre("save", func(doc interface{}) error {
    user, ok := doc.(*User)
    if !ok {
        return nil
    }
    
    // Trim whitespace from string fields
    user.Username = strings.TrimSpace(user.Username)
    user.Email = strings.TrimSpace(user.Email)
    
    // Normalize email to lowercase
    user.Email = strings.ToLower(user.Email)
    
    return nil
})
```

## Error Handling in Middleware

When a middleware function returns an error, the operation is aborted, and the error is returned to the caller. The error is wrapped with `ErrMiddleware` to indicate that it came from middleware:

```go
userSchema.Pre("save", func(doc interface{}) error {
    // Some validation that fails
    return errors.WithDetails(errors.ErrValidation, "validation failed in middleware")
})

// Later when using the model
err := userModel.Create(ctx, user)
if err != nil {
    if errors.IsMiddlewareError(err) {
        fmt.Println("Middleware error:", err)
        // Output might be: "Middleware error: middleware execution failed: validation failed in middleware"
    }
}
```

## Middleware Execution Order

Middleware functions are executed in the order they are added. If you add multiple middleware functions, they will be executed sequentially:

```go
// First middleware
userSchema.Pre("save", logRequest)

// Second middleware
userSchema.Pre("save", validateDocument)

// Third middleware
userSchema.Pre("save", transformDocument)
```

If any middleware returns an error, the execution chain is broken, and the error is returned.

## Best Practices

1. **Keep middleware functions focused**: Each middleware function should have a single responsibility.
2. **Handle type assertions carefully**: Always check that the document is of the expected type.
3. **Return meaningful errors**: Use `errors.WithDetails` or `errors.Wrap` to provide context.
4. **Don't mutate document IDs**: Avoid changing document IDs in middleware.
5. **Be mindful of performance**: Expensive operations in middleware will affect all document operations.