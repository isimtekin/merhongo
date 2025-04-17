// Package errors provides standardized error handling for Merhongo
package errors

import (
	"errors"
	"fmt"
)

// Standard errors that can be checked with errors.Is
var (
	// ErrNotFound indicates a document was not found
	ErrNotFound = errors.New("document not found")

	// ErrInvalidObjectID indicates an invalid MongoDB ObjectID
	ErrInvalidObjectID = errors.New("invalid ObjectID")

	// ErrValidation indicates a validation error
	ErrValidation = errors.New("validation failed")

	// ErrMiddleware indicates a middleware execution error
	ErrMiddleware = errors.New("middleware execution failed")

	// ErrNilCollection indicates operation on a nil collection
	ErrNilCollection = errors.New("collection is nil")

	// ErrDatabase indicates a general database operation error
	ErrDatabase = errors.New("database operation failed")

	// ErrConnection indicates a MongoDB connection error
	ErrConnection = errors.New("MongoDB connection failed")

	// ErrDecoding indicates an error decoding documents
	ErrDecoding = errors.New("failed to decode documents")
)

// WithDetails adds detailed information to a standard error
func WithDetails(err error, details string) error {
	return fmt.Errorf("%w: %s", err, details)
}

// Wrap wraps an error with additional context message
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapWithID wraps an error and includes the document ID in the message
func WrapWithID(err error, message string, id string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s (ID: %s): %w", message, id, err)
}
