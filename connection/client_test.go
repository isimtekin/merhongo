package connection

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"testing"
)

func TestConnectAndDisconnect(t *testing.T) {
	client, err := Connect("mongodb://localhost:27017", "merhongo_test")
	if err != nil {
		t.Fatalf("connection failed: %v", err)
	}
	err = client.Disconnect()
	if err != nil {
		t.Fatalf("disconnect failed: %v", err)
	}
}

func TestExecuteTransaction(t *testing.T) {
	client, err := Connect("mongodb://localhost:27017", "merhongo_test")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() {
		if err := client.Disconnect(); err != nil {
			t.Logf("Failed to disconnect: %v", err)
		}
	}()

	// Test a successful transaction with a no-op operation
	err = client.ExecuteTransaction(context.Background(), func(sc mongo.SessionContext) error {
		return nil
	})

	if err != nil {
		t.Errorf("expected successful transaction, got error: %v", err)
	}
}

func TestConnect_InvalidURI(t *testing.T) {
	_, err := Connect("mongodb://invalidhost:27017", "merhongo_test")
	if err == nil {
		t.Errorf("Expected error for invalid URI, got nil")
	}
}

func TestExecuteTransaction_Success(t *testing.T) {
	client, err := Connect("mongodb://localhost:27017", "merhongo_test")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() {
		if err := client.Disconnect(); err != nil {
			t.Logf("Failed to disconnect: %v", err)
		}
	}()

	err = client.ExecuteTransaction(context.Background(), func(sc mongo.SessionContext) error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
}

func TestExecuteTransaction_FnError(t *testing.T) {
	client, err := Connect("mongodb://localhost:27017", "merhongo_test")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() {
		if err := client.Disconnect(); err != nil {
			t.Logf("Failed to disconnect: %v", err)
		}
	}()

	err = client.ExecuteTransaction(context.Background(), func(sc mongo.SessionContext) error {
		return errors.New("fake failure in fn")
	})

	if err == nil || !strings.Contains(err.Error(), "fake failure") {
		t.Errorf("Expected fn error to be returned, got: %v", err)
	}
}
