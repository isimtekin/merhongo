# Error Handling

Merhongo provides a comprehensive error handling system to make it easier to identify, handle, and respond to different types of errors that may occur during database operations.

## Standard Error Types

Merhongo defines several standard error types in the `errors` package:

| Error Type | Description |
|------------|-------------|
| `ErrNotFound` | Document not found in the collection |
| `ErrInvalidObjectID` | Invalid MongoDB ObjectID format |
| `ErrValidation` | Validation error for document fields |
| `ErrMiddleware` | Error occurred in middleware execution |
| `ErrNilCollection` | Operation attempted on a nil collection |
| `ErrDatabase` | General database operation error |
| `ErrConnection` | MongoDB connection error |
| `ErrDecoding` | Error decoding documents from MongoDB |

## Checking Error Types

You can check error types using the standard Go `errors.Is()` function or the helper functions provided in the `errors` package:

```go
import (
    "github.com/isimtekin/merhongo/errors"
)

// Using standard Go errors.Is
if errors.Is(err, errors.ErrNotFound) {
    // Handle not found error
}

// Using Merhongo helper functions
if errors.IsNotFound(err) {
    // Handle not found error
}
if errors.IsValidationError(err) {
    // Handle validation error
}
if errors.IsInvalidObjectID(err) {
    // Handle invalid ObjectID error
}
```

The available helper functions are:

- `IsNotFound(err error) bool`
- `IsInvalidObjectID(err error) bool`
- `IsValidationError(err error) bool`
- `IsMiddlewareError(err error) bool`
- `IsNilCollectionError(err error) bool`
- `IsDatabaseError(err error) bool`
- `IsConnectionError(err error) bool`
- `IsDecodingError(err error) bool`

## Creating Custom Errors

Merhongo provides utility functions for wrapping or enhancing errors:

### WithDetails

Adds detailed information to a standard error:

```go
// Standard error with details
err := errors.WithDetails(errors.ErrValidation, "username must be at least 3 characters")

// Example in validation
func validateUsername(username string) error {
    if len(username) < 3 {
        return errors.WithDetails(errors.ErrValidation, 
            "username must be at least 3 characters long")
    }
    return nil
}
```

### Wrap

Wraps an error with an additional context message:

```go
// Wrap an error with context
err := errors.Wrap(originalError, "failed to process user data")

// Example in a service method
func (s *UserService) ProcessUserData(userData []byte) error {
    user, err := unmarshalUser(userData)
    if err != nil {
        return errors.Wrap(err, "could not parse user data")
    }
    // ...
}
```

### WrapWithID

Wraps an error and includes a document ID in the message:

```go
// Wrap an error with document ID
err := errors.WrapWithID(errors.ErrNotFound, "user not found", userID)

// Example in a service method
func (s *UserService) GetUserByID(id string) (*User, error) {
    user, err := s.userModel.FindById(ctx, id)
    if err != nil {
        return nil, errors.WrapWithID(err, "failed to retrieve user", id)
    }
    return user, nil
}
```

## Getting Error Details

You can get detailed information from an error:

```go
// Get full error details
details := errors.GetErrorDetails(err)
fmt.Println("Error details:", details)
```

## Formatting Errors for Display

The `FormatError` function formats an error for logging or display:

```go
// Format error for display
formattedError := errors.FormatError(err)
fmt.Println(formattedError)
// Output example: "[NotFound] document not found: user with ID '123' not found"
```

## Structured Error Responses for APIs

The `ToErrorResponse` function converts errors to a structured format suitable for API responses:

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

The `ErrorResponse` struct has the following fields:

```go
type ErrorResponse struct {
    Code    string `json:"code"`    // Error code like "not_found", "validation_error"
    Message string `json:"message"` // Human-readable message
    Details string `json:"details,omitempty"` // Detailed information
}
```

## Best Practices

1. **Use standard error types**: Stick to the standard error types provided by the package when possible.
2. **Add context**: Use `errors.Wrap()` or `errors.WithDetails()` to add context to errors.
3. **Check error types**: Use `errors.Is()` or the helper functions to check for specific error types.
4. **Provide useful error messages**: Make error messages descriptive and helpful for debugging.
5. **Return appropriate HTTP status codes**: Map Merhongo errors to appropriate HTTP status codes in web applications.
6. **Centralize error handling**: Create error handling middleware for web applications.
7. **Log detailed errors**: Log detailed errors on the server but return sanitized errors to clients.