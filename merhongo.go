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

// ModelNew is a convenience function to create a new model.
// It's a simple wrapper around model.New.
func ModelNew(name string, schema *schema.Schema, db *mongo.Database) *model.Model {
	return model.New(name, schema, db)
}

// QueryNew is a convenience function to create a new query builder.
// It's a simple wrapper around query.New.
func QueryNew() *query.Builder {
	return query.New()
}

// Version returns the current version of the merhongo package.
func Version() string {
	return "0.2.0"
}
