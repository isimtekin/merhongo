package errors_test

import (
	stderrors "errors"
	"fmt"
	"github.com/isimtekin/merhongo/errors"
	"testing"
)

func TestErrors_StandardErrors(t *testing.T) {
	// Ensure all standard errors are defined
	standardErrors := []error{
		errors.ErrNotFound,
		errors.ErrInvalidObjectID,
		errors.ErrValidation,
		errors.ErrMiddleware,
		errors.ErrNilCollection,
		errors.ErrDatabase,
		errors.ErrConnection,
		errors.ErrDecoding,
	}

	for _, err := range standardErrors {
		if err == nil {
			t.Errorf("Standard error should not be nil")
		}
		if err.Error() == "" {
			t.Errorf("Standard error should have a message")
		}
	}
}

func TestWithDetails(t *testing.T) {
	// Test adding details to a standard error
	baseErr := errors.ErrValidation
	details := "field 'name' is required"
	err := errors.WithDetails(baseErr, details)

	// Check that the error message contains both the base error and details
	expectedMsg := fmt.Sprintf("%s: %s", baseErr.Error(), details)
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Check that errors.Is works with the wrapped error
	if !stderrors.Is(err, baseErr) {
		t.Errorf("errors.Is should return true for the base error")
	}
}

func TestWrap(t *testing.T) {
	// Test wrapping an error with a message
	baseErr := fmt.Errorf("original error")
	message := "failed to process request"
	err := errors.Wrap(baseErr, message)

	// Check that the error message contains both the message and original error
	expectedMsg := fmt.Sprintf("%s: %s", message, baseErr.Error())
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Check that errors.Is works with the wrapped error
	if !stderrors.Is(err, baseErr) {
		t.Errorf("errors.Is should return true for the base error")
	}

	// Test that wrap returns nil for nil error
	if errors.Wrap(nil, message) != nil {
		t.Errorf("Wrap should return nil for nil error")
	}
}

func TestWrapWithID(t *testing.T) {
	// Test wrapping an error with a message and ID
	baseErr := errors.ErrNotFound
	message := "failed to find user"
	id := "507f1f77bcf86cd799439011"
	err := errors.WrapWithID(baseErr, message, id)

	// Check that the error message contains the message, ID, and original error
	expectedMsg := fmt.Sprintf("%s (ID: %s): %s", message, id, baseErr.Error())
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Check that errors.Is works with the wrapped error
	if !stderrors.Is(err, baseErr) {
		t.Errorf("errors.Is should return true for the base error")
	}

	// Test that WrapWithID returns nil for nil error
	if errors.WrapWithID(nil, message, id) != nil {
		t.Errorf("WrapWithID should return nil for nil error")
	}
}

func TestNestedWrapping(t *testing.T) {
	// Test nested error wrapping
	baseErr := errors.ErrValidation
	err1 := errors.WithDetails(baseErr, "field validation failed")
	err2 := errors.Wrap(err1, "processing error")
	err3 := errors.WrapWithID(err2, "user operation failed", "user123")

	// Check that errors.Is works with deeply nested errors
	if !stderrors.Is(err3, baseErr) {
		t.Errorf("errors.Is should return true for the base error in nested wrapping")
	}

	// Check that the error message has all the components
	if err3.Error() == "" {
		t.Errorf("Nested wrapped error should have a message")
	}
}
