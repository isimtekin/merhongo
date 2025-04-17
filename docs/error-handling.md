# Error Handling in Merhongo

Merhongo provides a comprehensive error handling system to make it easier to identify, handle, and respond to different types of errors that may occur during database operations.

## Standard Error Types

Merhongo defines several standard error types in the `errors` package:

- `ErrNotFound`: Document not found in the collection
- `ErrInvalidObjectID`: Invalid MongoDB ObjectID format
- `ErrValidation`: Validation error for document fields
- `ErrMiddleware`: Error occurred in middleware execution
- `ErrNilCollection`: Operation attempted on a nil collection
- `ErrDatabase`: General database operation error
- `ErrConnection`: MongoDB connection error
- `ErrDecoding`: Error decoding documents from MongoDB

## Checking Error Types

You can check error types using the standard Go `errors.Is()` function or the helper functions provided in the `errors` package:

```go
import (
    "github.com/isimtekin/merhongo/errors"
)

// Using errors.Is
if errors.Is(err, errors.ErrNotFound) {
    // Handle not found error
}

// Using helper functions
if errors.IsNotFound(err) {
    // Handle not found error
}
```

## Error Handling Examples

### Basic Error Checking

```go
user := &User{Username: "johndoe", Email: "john@example.com"}
err := userModel.Create(ctx, user)
if err != nil {
    if errors.IsValidationError(err) {
        fmt.Println("Validation error:", err)
        return
    }
    fmt.Println("Failed to create user:", err)
    return
}
```

### Error Details

All errors contain detailed information that you can access through the error message:

```go
err := userModel.FindById(ctx, "invalid-id", &result)
if err != nil {
    details := errors.GetErrorDetails(err)
    fmt.Println("Error details:", details)
    return
}
```

### Structured Error Responses for HTTP APIs

You can convert errors to structured responses easily:

```go
import (
    "encoding/json"
    "net/http"
    "github.com/isimtekin/merhongo/errors"
)

func handleGetUser(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    
    var user User
    err := userModel.FindById(r.Context(), id, &user)
    if err != nil {
        response := errors.ToErrorResponse(err)
        w.Header().Set("Content-Type", "application/json")
        
        // Set appropriate status code based on error type
        if errors.IsNotFound(err) {
            w.WriteHeader(http.StatusNotFound)
        } else if errors.IsInvalidObjectID(err) {
            w.WriteHeader(http.StatusBadRequest)
        } else {
            w.WriteHeader(http.StatusInternalServerError)
        }
        
        json.NewEncoder(w).Encode(response)
        return
    }
    
    // Return the user as JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

## Creating Custom Errors

You can create custom errors that wrap the standard errors:

```go
import (
    "github.com/isimtekin/merhongo/errors"
)

// Create a specific validation error
func validateUsername(username string) error {
    if len(username) < 3 {
        return errors.WithDetails(errors.ErrValidation, 
            "username must be at least 3 characters long")
    }
    return nil
}

// Wrapping an error with context
func (s *UserService) GetUserProfile(id string) (*Profile, error) {
    var user User
    err := userModel.FindById(ctx, id, &user)
    if err != nil {
        return nil, errors.Wrap(err, "failed to fetch user profile")
    }
    
    // Create and return profile
    return &Profile{
        DisplayName: user.Username,
        Email: user.Email,
    }, nil
}
```

## Best Practices

1. **Use standard error types**: Stick to the standard error types provided by the package when possible.
2. **Add context**: Use `errors.Wrap()` or `errors.WithDetails()` to add context to errors.
3. **Check error types**: Use `errors.Is()` or the helper functions to check for specific error types.
4. **Provide useful error messages**: Make error messages descriptive and helpful for debugging.
5. **Return appropriate HTTP status codes**: Map Merhongo errors to appropriate HTTP status codes in web applications.