# API Reference

This document provides a reference of the main APIs available in Merhongo. For more detailed information, refer to the GoDoc documentation.

## Package: merhongo

The main package provides functions for connection management and convenience wrappers for the sub-packages.

### Connection Management

```go
// Connect creates a new MongoDB connection and stores it as the default connection
func Connect(uri, dbName string) (*connection.Client, error)

// ConnectWithName creates a named MongoDB connection
func ConnectWithName(name, uri, dbName string) (*connection.Client, error)

// GetConnection returns the default connection
func GetConnection() *connection.Client

// GetConnectionByName returns a specific named connection
func GetConnectionByName(name string) *connection.Client

// Disconnect closes the default connection
func Disconnect() error

// DisconnectByName closes a specific named connection
func DisconnectByName(name string) error

// DisconnectAll closes all connections
func DisconnectAll() error
```

### Helper Functions

```go
// SchemaNew creates a new schema
func SchemaNew(fields map[string]schema.Field, options ...schema.Option) *schema.Schema

// ModelNew creates a new model with type safety
func ModelNew[T any](name string, schema *schema.Schema, options ...ModelOptions) *model.GenericModel[T]

// QueryNew creates a new query builder
func QueryNew() *query.Builder
```

### ModelOptions

```go
// ModelOptions contains optional settings for model creation
type ModelOptions struct {
// Database is the MongoDB database to use
Database *mongo.Database

// ConnectionName specifies a named connection to use if Database is nil
ConnectionName string

// AutoCreateIndexes determines if indexes should be created automatically
AutoCreateIndexes bool

// CustomValidator can override the default document validator
CustomValidator func(interface{}) error
}
```

## Package: schema

The schema package provides functionality for defining MongoDB document structures and validation rules.

### Schema Creation

```go
// New creates a new Schema with the specified fields and options
func New(fields map[string]Field, options ...Option) *Schema

// GenerateFromStruct automatically generates a Schema from a struct type
func GenerateFromStruct(structType interface{}, options ...Option) *Schema

// Field represents a schema field definition with validation rules
type Field struct {
Type         interface{}
Required     bool
Default      interface{}
Unique       bool
Min          int
Max          int
Enum         []interface{}
ValidateFunc func(interface{}) bool
}

// Schema defines the structure and validation rules for a MongoDB collection
type Schema struct {
Fields          map[string]Field
Timestamps      bool
Collection      string
Middlewares     map[string][]func(interface{}) error
CustomValidator func(doc interface{}) error
}
```

### Schema Options

```go
// WithCollection sets the collection name for the schema
func WithCollection(name string) Option

// WithTimestamps enables or disables automatic timestamps
func WithTimestamps(enable bool) Option
```

### Schema Methods

```go
// Pre adds a middleware function to be executed before the specified event
func (s *Schema) Pre(event string, fn func(interface{}) error)

// ValidateDocument validates a document against the schema
func (s *Schema) ValidateDocument(doc interface{}) error
```

## Package: model

The model package provides MongoDB collection model operations.

### Model Creation

```go
// New creates a new model for a collection
func New(name string, schema *schema.Schema, db *mongo.Database) *Model

// NewGeneric creates a new generic model with type-safe operations
func NewGeneric[T any](name string, schema *schema.Schema, db *mongo.Database) *GenericModel[T]
```

### Model Operations

```go
// Create inserts a new document into the collection
func (m *Model) Create(ctx context.Context, doc interface{}) error

// FindById finds a document by its ID
func (m *Model) FindById(ctx context.Context, id string, result interface{}) error

// Find finds documents matching the filter
func (m *Model) Find(ctx context.Context, filter interface{}, results interface{}) error

// FindOne finds a single document matching the filter
func (m *Model) FindOne(ctx context.Context, filter interface{}, result interface{}) error

// UpdateById updates a document by its ID
func (m *Model) UpdateById(ctx context.Context, id string, update interface{}) error

// DeleteById deletes a document by its ID
func (m *Model) DeleteById(ctx context.Context, id string) error

// Count returns the number of documents matching the filter
func (m *Model) Count(ctx context.Context, filter interface{}) (int64, error)
```

### GenericModel Operations

```go
// Create inserts a new document with type safety
func (m *GenericModel[T]) Create(ctx context.Context, doc *T) error

// FindById finds a document by its ID with type safety
func (m *GenericModel[T]) FindById(ctx context.Context, id string) (*T, error)

// Find finds documents matching the filter with type safety
func (m *GenericModel[T]) Find(ctx context.Context, filter interface{}) ([]T, error)

// FindOne finds a single document matching the filter with type safety
func (m *GenericModel[T]) FindOne(ctx context.Context, filter interface{}) (*T, error)

// UpdateById updates a document by its ID with type safety
func (m *GenericModel[T]) UpdateById(ctx context.Context, id string, update interface{}) error

// DeleteById deletes a document by its ID with type safety
func (m *GenericModel[T]) DeleteById(ctx context.Context, id string) error

// Count returns the number of documents matching the filter with type safety
func (m *GenericModel[T]) Count(ctx context.Context, filter interface{}) (int64, error)
```

### Query Methods

```go
// FindWithQuery finds documents using a query builder
func (m *Model) FindWithQuery(ctx context.Context, queryBuilder *query.Builder, results interface{}) error

// FindOneWithQuery finds a single document using a query builder
func (m *Model) FindOneWithQuery(ctx context.Context, queryBuilder *query.Builder, result interface{}) error

// CountWithQuery counts documents using a query builder
func (m *Model) CountWithQuery(ctx context.Context, queryBuilder *query.Builder) (int64, error)

// UpdateWithQuery updates documents using a query builder
func (m *Model) UpdateWithQuery(ctx context.Context, queryBuilder *query.Builder, update interface{}) (int64, error)

// DeleteWithQuery deletes documents using a query builder
func (m *Model) DeleteWithQuery(ctx context.Context, queryBuilder *query.Builder) (int64, error)
```

## Package: query

The query package provides a fluent API for building MongoDB queries.

### Query Builder Creation

```go
// New creates a new query builder
func New() *Builder

// WithError creates a new query builder that starts with an error
func WithError(err error) *Builder
```

### Query Conditions

```go
// Where adds filter conditions
func (b *Builder) Where(key string, value interface{}) *Builder

// WhereOperator adds a filter with a specific operator
func (b *Builder) WhereOperator(key string, operator string, value interface{}) *Builder

// Equals adds an equals condition
func (b *Builder) Equals(key string, value interface{}) *Builder

// NotEquals adds a not equals condition
func (b *Builder) NotEquals(key string, value interface{}) *Builder

// GreaterThan adds a greater than condition
func (b *Builder) GreaterThan(key string, value interface{}) *Builder

// GreaterThanOrEqual adds a greater than or equal condition
func (b *Builder) GreaterThanOrEqual(key string, value interface{}) *Builder

// LessThan adds a less than condition
func (b *Builder) LessThan(key string, value interface{}) *Builder

// LessThanOrEqual adds a less than or equal condition
func (b *Builder) LessThanOrEqual(key string, value interface{}) *Builder

// In adds an in condition
func (b *Builder) In(key string, values interface{}) *Builder

// NotIn adds a not in condition
func (b *Builder) NotIn(key string, values interface{}) *Builder

// Exists adds an exists condition
func (b *Builder) Exists(key string, exists bool) *Builder

// Regex adds a regular expression condition
func (b *Builder) Regex(key string, pattern string, options ...string) *Builder

// MergeFilter merges another filter into this builder
func (b *Builder) MergeFilter(filter bson.M) *Builder
```

### Query Options

```go
// SortBy adds sort criteria
func (b *Builder) SortBy(key string, ascending bool) *Builder

// Limit sets the maximum number of results
func (b *Builder) Limit(limit int64) *Builder

// Skip sets the number of results to skip
func (b *Builder) Skip(skip int64) *Builder
```

### Query Building

```go
// GetFilter returns the filter
func (b *Builder) GetFilter() (bson.M, error)

// GetOptions returns the query options
func (b *Builder) GetOptions() (*options.FindOptions, error)

// Build returns both the filter and options
func (b *Builder) Build() (bson.M, *options.FindOptions, error)

// Error returns any error that occurred during query building
func (b *Builder) Error() error
```

## Package: errors

The errors package provides standardized error handling for Merhongo.

### Standard Errors

```go
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
```

### Error Wrapping

```go
// WithDetails adds detailed information to a standard error
func WithDetails(err error, details string) error

// Wrap wraps an error with additional context message
func Wrap(err error, message string) error

// WrapWithID wraps an error and includes the document ID in the message
func WrapWithID(err error, message string, id string) error
```

### Error Checking

```go
// IsNotFound checks if an error is or wraps ErrNotFound
func IsNotFound(err error) bool

// IsInvalidObjectID checks if an error is or wraps ErrInvalidObjectID
func IsInvalidObjectID(err error) bool

// IsValidationError checks if an error is or wraps ErrValidation
func IsValidationError(err error) bool

// IsMiddlewareError checks if an error is or wraps ErrMiddleware
func IsMiddlewareError(err error) bool

// IsNilCollectionError checks if an error is or wraps ErrNilCollection
func IsNilCollectionError(err error) bool

// IsDatabaseError checks if an error is or wraps ErrDatabase
func IsDatabaseError(err error) bool

// IsConnectionError checks if an error is or wraps ErrConnection
func IsConnectionError(err error) bool

// IsDecodingError checks if an error is or wraps ErrDecoding
func IsDecodingError(err error) bool
```

### Error Utilities

```go
// GetErrorDetails returns the detailed message from the error if available
func GetErrorDetails(err error) string

// FormatError formats an error for logging or display
func FormatError(err error) string

// ErrorResponse represents a structured error response
type ErrorResponse struct {
Code    string `json:"code"`
Message string `json:"message"`
Details string `json:"details,omitempty"`
}

// ToErrorResponse converts an error to a structured response
func ToErrorResponse(err error) ErrorResponse
```

## Package: connection

The connection package handles MongoDB client connection management.

### Client Operations

```go
// Connect creates a new MongoDB client instance and connects to the database
func Connect(uri, dbName string) (*Client, error)

// Disconnect closes the MongoDB connection
func (c *Client) Disconnect() error

// ExecuteTransaction runs operations in a transaction
func (c *Client) ExecuteTransaction(ctx context.Context, fn func(mongo.SessionContext) error) error

// GetDatabase returns the database with the specified name
func (c *Client) GetDatabase(name string) *mongo.Database

// RegisterModel registers a model with this connection
func (c *Client) RegisterModel(name string, model interface{})

// GetModel retrieves a registered model by name
func (c *Client) GetModel(name string) interface{}
```