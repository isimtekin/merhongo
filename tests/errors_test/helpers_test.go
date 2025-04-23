package errors_test

import (
	"fmt"
	"github.com/isimtekin/merhongo/errors"
	"strings"
	"testing"
)

func TestIsErrorFunctions(t *testing.T) {
	// Test all the IsX functions with matching errors
	testCases := []struct {
		name     string
		err      error
		checkFn  func(error) bool
		expected bool
	}{
		{
			name:     "IsNotFound with ErrNotFound",
			err:      errors.ErrNotFound,
			checkFn:  errors.IsNotFound,
			expected: true,
		},
		{
			name:     "IsNotFound with wrapped ErrNotFound",
			err:      errors.Wrap(errors.ErrNotFound, "database error"),
			checkFn:  errors.IsNotFound,
			expected: true,
		},
		{
			name:     "IsNotFound with different error",
			err:      errors.ErrValidation,
			checkFn:  errors.IsNotFound,
			expected: false,
		},
		{
			name:     "IsInvalidObjectID with ErrInvalidObjectID",
			err:      errors.ErrInvalidObjectID,
			checkFn:  errors.IsInvalidObjectID,
			expected: true,
		},
		{
			name:     "IsValidationError with ErrValidation",
			err:      errors.ErrValidation,
			checkFn:  errors.IsValidationError,
			expected: true,
		},
		{
			name:     "IsMiddlewareError with ErrMiddleware",
			err:      errors.ErrMiddleware,
			checkFn:  errors.IsMiddlewareError,
			expected: true,
		},
		{
			name:     "IsNilCollectionError with ErrNilCollection",
			err:      errors.ErrNilCollection,
			checkFn:  errors.IsNilCollectionError,
			expected: true,
		},
		{
			name:     "IsDatabaseError with ErrDatabase",
			err:      errors.ErrDatabase,
			checkFn:  errors.IsDatabaseError,
			expected: true,
		},
		{
			name:     "IsConnectionError with ErrConnection",
			err:      errors.ErrConnection,
			checkFn:  errors.IsConnectionError,
			expected: true,
		},
		{
			name:     "IsDecodingError with ErrDecoding",
			err:      errors.ErrDecoding,
			checkFn:  errors.IsDecodingError,
			expected: true,
		},
		{
			name:     "IsError with nil error",
			err:      nil,
			checkFn:  errors.IsNotFound,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.checkFn(tc.err)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestGetErrorDetails(t *testing.T) {
	// Test getting details from various errors
	testCases := []struct {
		name           string
		err            error
		expectedDetail string
	}{
		{
			name:           "nil error",
			err:            nil,
			expectedDetail: "",
		},
		{
			name:           "simple error",
			err:            fmt.Errorf("simple error"),
			expectedDetail: "simple error",
		},
		{
			name:           "wrapped error",
			err:            errors.Wrap(errors.ErrNotFound, "user lookup failed"),
			expectedDetail: "user lookup failed: document not found",
		},
		{
			name:           "error with details",
			err:            errors.WithDetails(errors.ErrValidation, "field 'email' is invalid"),
			expectedDetail: "validation failed: field 'email' is invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			detail := errors.GetErrorDetails(tc.err)
			if detail != tc.expectedDetail {
				t.Errorf("Expected detail '%s', got '%s'", tc.expectedDetail, detail)
			}
		})
	}
}

func TestFormatError(t *testing.T) {
	// Test error formatting for different error types
	testCases := []struct {
		name        string
		err         error
		shouldMatch string
	}{
		{
			name:        "nil error",
			err:         nil,
			shouldMatch: "",
		},
		{
			name:        "not found error",
			err:         errors.ErrNotFound,
			shouldMatch: "[NotFound]",
		},
		{
			name:        "wrapped not found error",
			err:         errors.Wrap(errors.ErrNotFound, "user lookup"),
			shouldMatch: "[NotFound] user lookup",
		},
		{
			name:        "validation error",
			err:         errors.WithDetails(errors.ErrValidation, "required field"),
			shouldMatch: "[Validation]",
		},
		{
			name:        "unknown error type",
			err:         fmt.Errorf("random error"),
			shouldMatch: "[Unknown]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatted := errors.FormatError(tc.err)
			if tc.err != nil && !strings.Contains(formatted, tc.shouldMatch) {
				t.Errorf("Formatted error '%s' should contain '%s'", formatted, tc.shouldMatch)
			}
		})
	}
}

func TestToErrorResponse(t *testing.T) {
	// Test converting errors to structured responses
	testCases := []struct {
		name          string
		err           error
		expectedCode  string
		expectedIsSet bool
	}{
		{
			name:          "nil error",
			err:           nil,
			expectedCode:  "unknown",
			expectedIsSet: true,
		},
		{
			name:          "not found error",
			err:           errors.ErrNotFound,
			expectedCode:  "not_found",
			expectedIsSet: true,
		},
		{
			name:          "invalid object id",
			err:           errors.ErrInvalidObjectID,
			expectedCode:  "invalid_id",
			expectedIsSet: true,
		},
		{
			name:          "validation error",
			err:           errors.ErrValidation,
			expectedCode:  "validation_error",
			expectedIsSet: true,
		},
		{
			name:          "middleware error",
			err:           errors.ErrMiddleware,
			expectedCode:  "middleware_error",
			expectedIsSet: true,
		},
		{
			name:          "nil collection error",
			err:           errors.ErrNilCollection,
			expectedCode:  "collection_error",
			expectedIsSet: true,
		},
		{
			name:          "database error",
			err:           errors.ErrDatabase,
			expectedCode:  "database_error",
			expectedIsSet: true,
		},
		{
			name:          "connection error",
			err:           errors.ErrConnection,
			expectedCode:  "connection_error",
			expectedIsSet: true,
		},
		{
			name:          "decoding error",
			err:           errors.ErrDecoding,
			expectedCode:  "decoding_error",
			expectedIsSet: true,
		},
		{
			name:          "unknown error",
			err:           fmt.Errorf("unknown error"),
			expectedCode:  "unknown_error",
			expectedIsSet: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := errors.ToErrorResponse(tc.err)

			// Check the error code
			if response.Code != tc.expectedCode {
				t.Errorf("Expected code '%s', got '%s'", tc.expectedCode, response.Code)
			}

			// Check that message is set
			if tc.expectedIsSet && response.Message == "" {
				t.Errorf("Expected message to be set")
			}

			// For errors with details, check that details are included
			if tc.err != nil && tc.err != errors.ErrNotFound && tc.err != errors.ErrValidation {
				if strings.Contains(tc.err.Error(), ":") && response.Details == "" {
					t.Errorf("Expected details to be extracted from error")
				}
			}
		})
	}
}

func TestErrorResponseStructure(t *testing.T) {
	// Create a detailed error
	err := errors.WithDetails(errors.ErrValidation, "field 'email' is invalid")

	// Convert to response
	response := errors.ToErrorResponse(err)

	// Check structure
	if response.Code != "validation_error" {
		t.Errorf("Expected code 'validation_error', got '%s'", response.Code)
	}

	if response.Message != "Validation failed" {
		t.Errorf("Expected message 'Validation failed', got '%s'", response.Message)
	}

	if !strings.Contains(response.Details, "field 'email' is invalid") {
		t.Errorf("Expected details to contain the field message")
	}
}
