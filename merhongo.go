// Package merhongo provides MongoDB utilities inspired by Mongoose
package merhongo

import (
	"sync"

	"github.com/isimtekin/merhongo/connection"
	"github.com/isimtekin/merhongo/errors"
	"github.com/isimtekin/merhongo/model"
	"github.com/isimtekin/merhongo/query"
	"github.com/isimtekin/merhongo/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	// connections is a map of named MongoDB client connections
	connections = make(map[string]*connection.Client)
	// defaultConnectionName is the key used for the default connection
	defaultConnectionName = "default"
	// connectionMutex guards access to the connections map
	connectionMutex sync.RWMutex
)

// Connect creates a new MongoDB connection and stores it as the default connection.
// It returns the connection client or an error if the connection fails.
func Connect(uri, dbName string) (*connection.Client, error) {
	return ConnectWithName(defaultConnectionName, uri, dbName)
}

// ConnectWithName creates a new MongoDB connection with the specified name.
// This allows maintaining multiple connections to different databases.
func ConnectWithName(name, uri, dbName string) (*connection.Client, error) {
	if name == "" {
		return nil, errors.WithDetails(errors.ErrValidation, "connection name cannot be empty")
	}

	client, err := connection.Connect(uri, dbName)
	if err != nil {
		return nil, err
	}

	connectionMutex.Lock()
	connections[name] = client
	connectionMutex.Unlock()

	return client, nil
}

// GetConnection returns the default connection.
// Returns nil if the default connection has not been established.
func GetConnection() *connection.Client {
	return GetConnectionByName(defaultConnectionName)
}

// GetConnectionByName returns the connection with the specified name.
// Returns nil if no connection exists with the given name.
func GetConnectionByName(name string) *connection.Client {
	connectionMutex.RLock()
	defer connectionMutex.RUnlock()

	return connections[name]
}

// DisconnectAll closes all stored connections.
// Returns an error if any connection fails to disconnect.
func DisconnectAll() error {
	connectionMutex.Lock()
	defer connectionMutex.Unlock()

	for name, client := range connections {
		if err := client.Disconnect(); err != nil {
			return err
		}
		delete(connections, name)
	}

	return nil
}

// Disconnect closes the default connection.
// Returns an error if the disconnection fails.
func Disconnect() error {
	return DisconnectByName(defaultConnectionName)
}

// DisconnectByName closes the connection with the specified name.
// Returns an error if the disconnection fails.
// No error is returned if the connection doesn't exist.
func DisconnectByName(name string) error {
	connectionMutex.Lock()
	defer connectionMutex.Unlock()

	client, exists := connections[name]
	if !exists {
		return nil
	}

	if err := client.Disconnect(); err != nil {
		return err
	}

	delete(connections, name)
	return nil
}

// SchemaNew is a convenience function to create a new schema.
// It's a simple wrapper around schema.New.
func SchemaNew(fields map[string]schema.Field, options ...schema.Option) *schema.Schema {
	return schema.New(fields, options...)
}

// ModelOptions contains optional settings for model creation
type ModelOptions struct {
	// Database is the MongoDB database to use, if nil the default connection database is used
	Database *mongo.Database
	// ConnectionName specifies a named connection to use if Database is nil
	ConnectionName string
	// AutoCreateIndexes determines if indexes should be created automatically
	AutoCreateIndexes bool
	// CustomValidator can override the default document validator
	CustomValidator func(interface{}) error
}

// ModelNew is a convenience function to create a new model.
// It accepts the model struct type as a generic parameter.
// If no database is specified in options, it uses the default client's database.
func ModelNew[T any](name string, schema *schema.Schema, options ...ModelOptions) *model.Model {
	// Default options
	opts := ModelOptions{
		AutoCreateIndexes: true,
	}

	// Apply provided options if any
	if len(options) > 0 {
		opts = options[0]
	}

	// Determine which database to use
	var db *mongo.Database
	if opts.Database != nil {
		// Use explicitly provided database
		db = opts.Database
	} else if opts.ConnectionName != "" {
		// Use database from specified connection
		client := GetConnectionByName(opts.ConnectionName)
		if client != nil {
			db = client.Database
		}
	} else {
		// Use default connection's database
		defaultClient := GetConnection()
		if defaultClient != nil {
			db = defaultClient.Database
		}
	}

	// Create the model
	m := model.New(name, schema, db)

	// Apply custom validator if provided
	if opts.CustomValidator != nil && m.Schema != nil {
		m.Schema.CustomValidator = opts.CustomValidator
	}

	// Register the type with the model if we have a valid connection
	var modelType T

	// Find if we have a connection client that implements RegisterModel
	if db != nil {
		// Check if the db belongs to our own connection.Client
		// (We can't directly cast mongo.Client to connection.Client)
		for _, client := range connections {
			if client.Database == db {
				client.RegisterModel(name, &modelType)
				break
			}
		}
	}

	return m
}

// QueryNew is a convenience function to create a new query builder.
// It's a simple wrapper around query.New.
func QueryNew() *query.Builder {
	return query.New()
}
