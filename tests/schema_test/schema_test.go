package schema_test

import (
	"github.com/isimtekin/merhongo/schema"
	"testing"
)

func TestSchemaCreation(t *testing.T) {
	fields := map[string]schema.Field{
		"Email": {
			Required: true,
			Unique:   true,
		},
	}
	s := schema.New(fields, schema.WithCollection("users"))

	if s.Fields["Email"].Required != true {
		t.Errorf("Expected Email field to be required")
	}

	if s.Collection != "users" {
		t.Errorf("Expected collection name to be 'users'")
	}
}

func TestWithTimestampsOption(t *testing.T) {
	// Default should be true
	s := schema.New(map[string]schema.Field{})
	if !s.Timestamps {
		t.Error("expected timestamps to be true by default")
	}

	// Now disable it
	s = schema.New(
		map[string]schema.Field{},
		schema.WithTimestamps(false),
	)

	if s.Timestamps {
		t.Error("expected timestamps to be disabled")
	}
}

func TestPreMiddlewareRegistration(t *testing.T) {
	s := schema.New(map[string]schema.Field{})

	triggered := false
	s.Pre("save", func(doc interface{}) error {
		triggered = true
		return nil
	})

	if s.Middlewares["save"] == nil || len(s.Middlewares["save"]) != 1 {
		t.Error("expected one pre-save middleware to be registered")
	}

	// Execute the middleware manually to ensure it's functional
	err := s.Middlewares["save"][0](nil)
	if err != nil || !triggered {
		t.Error("middleware function was not executed properly")
	}
}
