package errors

import (
	"errors"
	"fmt"
	"strings"
)

// IsNotFound checks if an error is or wraps ErrNotFound
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsInvalidObjectID checks if an error is or wraps ErrInvalidObjectID
func IsInvalidObjectID(err error) bool {
	return errors.Is(err, ErrInvalidObjectID)
}

// IsValidationError checks if an error is or wraps ErrValidation
func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidation)
}

// IsMiddlewareError checks if an error is or wraps ErrMiddleware
func IsMiddlewareError(err error) bool {
	return errors.Is(err, ErrMiddleware)
}

// IsNilCollectionError checks if an error is or wraps ErrNilCollection
func IsNilCollectionError(err error) bool {
	return errors.Is(err, ErrNilCollection)
}

// IsDatabaseError checks if an error is or wraps ErrDatabase
func IsDatabaseError(err error) bool {
	return errors.Is(err, ErrDatabase)
}

// IsConnectionError checks if an error is or wraps ErrConnection
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrConnection)
}

// IsDecodingError checks if an error is or wraps ErrDecoding
func IsDecodingError(err error) bool {
	return errors.Is(err, ErrDecoding)
}

// IsSchemaValidationError checks specifically for schema validation errors
func IsSchemaValidationError(err error) bool {
	// First check if it's a validation error at all
	if !IsValidationError(err) {
		return false
	}

	// Check if the error message contains schema validation related text
	errStr := err.Error()
	return strings.Contains(errStr, "field") ||
		strings.Contains(errStr, "required") ||
		strings.Contains(errStr, "validation") ||
		strings.Contains(errStr, "minimum") ||
		strings.Contains(errStr, "maximum") ||
		strings.Contains(errStr, "empty")
}

// IsTimestampError checks for errors related to timestamps (if we add specific checks)
func IsTimestampError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "timestamp") ||
		strings.Contains(errStr, "createdAt") ||
		strings.Contains(errStr, "updatedAt")
}

// ValidationErrorDetails extracts more detailed information from validation errors
func ValidationErrorDetails(err error) string {
	if !IsValidationError(err) {
		return ""
	}

	// Extract details after the error type prefix
	errStr := err.Error()
	parts := strings.SplitN(errStr, ":", 2)
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}

	return errStr
}

// GetErrorDetails returns the detailed message from the error if available
func GetErrorDetails(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

// FormatError formats an error for logging or display
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	// Start with the error message
	msg := err.Error()

	// Add error type classification
	var errType string
	switch {
	case IsNotFound(err):
		errType = "NotFound"
	case IsInvalidObjectID(err):
		errType = "InvalidObjectID"
	case IsValidationError(err):
		errType = "Validation"
	case IsMiddlewareError(err):
		errType = "Middleware"
	case IsNilCollectionError(err):
		errType = "NilCollection"
	case IsDatabaseError(err):
		errType = "Database"
	case IsConnectionError(err):
		errType = "Connection"
	case IsDecodingError(err):
		errType = "Decoding"
	default:
		errType = "Unknown"
	}

	return fmt.Sprintf("[%s] %s", errType, msg)
}

// ErrorResponse represents a structured error response that can be returned to clients
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ToErrorResponse converts an error to a structured response
func ToErrorResponse(err error) ErrorResponse {
	if err == nil {
		return ErrorResponse{
			Code:    "unknown",
			Message: "Unknown error",
		}
	}

	var code string
	var message string

	switch {
	case IsNotFound(err):
		code = "not_found"
		message = "Resource not found"
	case IsInvalidObjectID(err):
		code = "invalid_id"
		message = "Invalid identifier format"
	case IsValidationError(err):
		code = "validation_error"
		message = "Validation failed"
	case IsMiddlewareError(err):
		code = "middleware_error"
		message = "Processing error"
	case IsNilCollectionError(err):
		code = "collection_error"
		message = "Collection not available"
	case IsDatabaseError(err):
		code = "database_error"
		message = "Database operation failed"
	case IsConnectionError(err):
		code = "connection_error"
		message = "Database connection error"
	case IsDecodingError(err):
		code = "decoding_error"
		message = "Failed to decode data"
	default:
		code = "unknown_error"
		message = "An unexpected error occurred"
	}

	// Get detailed message, but clean it up
	details := err.Error()
	for _, baseErr := range []error{
		ErrNotFound, ErrInvalidObjectID, ErrValidation,
		ErrMiddleware, ErrNilCollection, ErrDatabase,
		ErrConnection, ErrDecoding,
	} {
		details = strings.Replace(details, baseErr.Error()+": ", "", 1)
	}

	return ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
}
