package database

import (
	gomongo "github.com/decisionbox-io/decisionbox/libs/go-common/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

// Collection names used by the decisionbox-agent.
const (
	CollectionDiscoveries = "ai_discoveries"
	CollectionAppContext  = "ai_app_context"
	CollectionApps        = "apps"
	CollectionDebugLogs   = "ai_discovery_debug_logs"
)

// DB wraps go-common's MongoDB client for decisionbox-agent.
type DB struct {
	client *gomongo.Client
}

// New creates a DB wrapper.
func New(client *gomongo.Client) *DB {
	return &DB{client: client}
}

// Client returns the underlying go-common MongoDB client.
func (db *DB) Client() *gomongo.Client {
	return db.client
}

// Collection returns a MongoDB collection by name.
func (db *DB) Collection(name string) *mongo.Collection {
	return db.client.Collection(name)
}

// Database returns the underlying mongo.Database for packages that need it.
func (db *DB) Database() *mongo.Database {
	return db.client.Database()
}
