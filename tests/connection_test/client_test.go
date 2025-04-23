package connection_test

import (
	"context"
	"errors"
	"github.com/isimtekin/merhongo/connection"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
)

// --- Mock definitions using testify/mock ---

// MockSessionContext is a mock implementation of mongo.SessionContext
type MockSessionContext struct {
	mock.Mock
	context.Context
}

// StartTransaction starts a transaction with the given options.
func (m *MockSessionContext) StartTransaction(opts ...*options.TransactionOptions) error {
	args := m.Called(opts)
	return args.Error(0)
}

// CommitTransaction commits the transaction.
func (m *MockSessionContext) CommitTransaction(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// AbortTransaction aborts the transaction.
func (m *MockSessionContext) AbortTransaction(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// WithTransaction executes the provided callback within a transaction.
func (m *MockSessionContext) WithTransaction(ctx context.Context, fn func(mongo.SessionContext) (interface{}, error), opts ...*options.TransactionOptions) (interface{}, error) {
	args := m.Called(ctx, fn, opts)
	return args.Get(0), args.Error(1)
}

// EndSession ends the session.
func (m *MockSessionContext) EndSession(ctx context.Context) {
	m.Called(ctx)
}

// ClusterTime returns the cluster time.
func (m *MockSessionContext) ClusterTime() bson.Raw {
	args := m.Called()
	if rf, ok := args.Get(0).(bson.Raw); ok {
		return rf
	}
	return bson.Raw{}
}

// OperationTime returns the operation time.
func (m *MockSessionContext) OperationTime() *primitive.Timestamp {
	args := m.Called()
	if ts, ok := args.Get(0).(*primitive.Timestamp); ok {
		return ts
	}
	return nil
}

// Client returns the client.
func (m *MockSessionContext) Client() *mongo.Client {
	args := m.Called()
	if cl, ok := args.Get(0).(*mongo.Client); ok {
		return cl
	}
	return nil
}

// ID returns the session ID.
func (m *MockSessionContext) ID() bson.Raw {
	args := m.Called()
	if rf, ok := args.Get(0).(bson.Raw); ok {
		return rf
	}
	return bson.Raw{}
}

// AdvanceClusterTime advances the cluster time.
func (m *MockSessionContext) AdvanceClusterTime(clusterTime bson.Raw) error {
	args := m.Called(clusterTime)
	return args.Error(0)
}

// AdvanceOperationTime advances the operation time.
func (m *MockSessionContext) AdvanceOperationTime(operationTime *primitive.Timestamp) error {
	args := m.Called(operationTime)
	return args.Error(0)
}

// SetCausalConsistency sets the causal consistency.
func (m *MockSessionContext) SetCausalConsistency(causalConsistency bool) {
	m.Called(causalConsistency)
}

// Session returns the underlying session.
func (m *MockSessionContext) Session() mongo.Session {
	args := m.Called()
	return args.Get(0).(mongo.Session)
}

// MockMongoClient is a mock implementation of the methods we need from mongo.Client
type MockMongoClient struct {
	mock.Mock
}

func (m *MockMongoClient) UseSession(ctx context.Context, fn func(mongo.SessionContext) error) error {
	args := m.Called(ctx, fn)

	// If there's a function to call, execute it with our mock session
	if mockFn, ok := args.Get(0).(func(context.Context, func(mongo.SessionContext) error) error); ok && mockFn != nil {
		return mockFn(ctx, fn)
	}

	return args.Error(1)
}

// --- Standard connection tests using real MongoDB ---

func TestConnectAndDisconnect(t *testing.T) {
	client, err := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	if err != nil {
		t.Fatalf("connection failed: %v", err)
	}
	err = client.Disconnect()
	if err != nil {
		t.Fatalf("disconnect failed: %v", err)
	}
}

func TestConnect_InvalidURI(t *testing.T) {
	_, err := connection.Connect("mongodb://invalidhost:27017", "merhongo_test")
	assert.Error(t, err, "Expected error for invalid URI")
}

func TestExecuteTransaction_Success(t *testing.T) {
	client, err := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Disconnect()

	err = client.ExecuteTransaction(context.Background(), func(sc mongo.SessionContext) error {
		return nil
	})

	assert.NoError(t, err, "Expected successful transaction")
}

func TestExecuteTransaction_FnError(t *testing.T) {
	client, err := connection.Connect("mongodb://localhost:27017", "merhongo_test")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Disconnect()

	expectedErr := errors.New("fake failure in fn")
	err = client.ExecuteTransaction(context.Background(), func(sc mongo.SessionContext) error {
		return expectedErr
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fake failure")
}

// --- Tests using testify mocks ---

func TestClient_GetDatabase(t *testing.T) {
	client, err := connection.Connect("mongodb://localhost:27017", "merhongo_test_getdb")
	if err != nil {
		t.Skipf("Skipping test; could not connect to MongoDB: %v", err)
		return
	}
	defer client.Disconnect()

	// Test with empty name
	defaultDB := client.GetDatabase("")
	assert.Equal(t, "merhongo_test_getdb", defaultDB.Name(), "Expected default database name")

	// Test with custom name
	customDBName := "merhongo_test_custom"
	customDB := client.GetDatabase(customDBName)
	assert.Equal(t, customDBName, customDB.Name(), "Expected custom database name")
}

func TestClient_RegisterModel_GetModel(t *testing.T) {
	// Create client directly - no need for real MongoDB connection
	client := &connection.Client{
		Models: make(map[string]interface{}),
	}

	// Test model
	type TestModel struct {
		Name string
	}
	model := &TestModel{Name: "TestModel"}
	modelName := "testModel"

	// Register and get back
	client.RegisterModel(modelName, model)
	retrievedModel := client.GetModel(modelName)

	assert.Equal(t, model, retrievedModel, "GetModel should return the same model instance")
	assert.Nil(t, client.GetModel("nonExistentModel"), "Nonexistent model should return nil")
}

func TestClient_Disconnect_NilClient(t *testing.T) {
	client := &connection.Client{
		MongoClient: nil,
		Models:      make(map[string]interface{}),
	}

	err := client.Disconnect()
	assert.NoError(t, err, "Disconnect with nil MongoClient should not error")
}

// --- Mock tests for transaction error cases ---

type testMongoClient struct {
	executeTransaction func(ctx context.Context, fn func(mongo.SessionContext) error) error
}

func (c *testMongoClient) UseSession(ctx context.Context, fn func(mongo.SessionContext) error) error {
	return c.executeTransaction(ctx, fn)
}
