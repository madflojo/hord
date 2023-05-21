// Package nats is a Hord database driver that interacts with a nats key-value store.
//
//	// Connect to NATS
//	db, err := nats.Dial(&nats.Config{})
//	if err != nil {
//	  // do stuff
//	}
//
//	// Setup and Initialize the Keyspace if necessary
//	err = db.Setup()
//	if err != nil {
//	  // do stuff
//	}
//
//	// Write data to the cluster
//	err = db.Set("mykey", []byte("My Data"))
//	if err != nil {
//	  // do stuff
//	}
//
//	// Fetch the same data
//	d, err := db.Get("mykey")
//	if err != nil {
//	  // do stuff
//	}
//
// The nats driver in the Hord package allows you to quickly use nats to store data with your Go
// applications. It provides methods to store and retrieve key-value pairs, enabling efficient goroutine safe data
// management.
package nats

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/madflojo/hord"
	"github.com/nats-io/nats.go"
)

// Config represents the configuration for the NATS database connection.
type Config struct {
	// URL of the NATS server
	URL string

	// Bucket name for the key-value store bucket
	Bucket string
}

// Database is an in-memory NATS implementation of the hord.Database interface.
type Database struct {
	sync.RWMutex

	// bucket name for the key-value store bucket
	bucket string

	// conn provides a NATS connection
	conn *nats.Conn

	// kv provides a NATS key-value store
	kv nats.KeyValue
}

// reBucket is used to validate bucket names
var reBucket = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Dial initializes and returns a new NATS database instance.
func Dial(cfg Config) (*Database, error) {
	var err error
	db := &Database{bucket: cfg.Bucket}

	// Validate Config
	if cfg.URL == "" {
		return db, fmt.Errorf("URL cannot be empty")
	}

	// Validate Bucket
	if cfg.Bucket == "" || !reBucket.MatchString(cfg.Bucket) {
		return db, fmt.Errorf("Bucket name is invalid")
	}

	// Connect to the NATS server
	db.conn, err = nats.Connect(cfg.URL)
	if err != nil {
		return db, fmt.Errorf("unable to connect to NATS server - %s", err)
	}

	// Create a JetStream context
	js, err := db.conn.JetStream()
	if err != nil {
		return db, fmt.Errorf("unable to open JetStream - %s", err)
	}

	// Create a key-value store within JetStream
	db.kv, err = js.CreateKeyValue(&nats.KeyValueConfig{Bucket: cfg.Bucket})
	if err != nil {
		return db, fmt.Errorf("unable to open key-value store - %s", err)
	}

	return db, nil
}

// Setup sets up the nats database. This function does nothing for the nats driver.
func (db *Database) Setup() error {
	err := db.HealthCheck()
	if err != nil {
		return fmt.Errorf("could not setup database, unhealthy - %s", err)
	}
	return nil
}

// Get retrieves data from the NATS database based on the provided key.
// It returns the data associated with the key or an error if the key is invalid or the data does not exist.
func (db *Database) Get(key string) ([]byte, error) {
	// Validate the key
	if err := hord.ValidKey(key); err != nil {
		return []byte(""), err
	}

	// Acquire a read lock to ensure data consistency during retrieval
	db.RLock()
	defer db.RUnlock()

	// Check if the NATS key-value store is initialized
	if db.kv == nil {
		return []byte(""), hord.ErrNoDial
	}

	// Retrieve the value from the NATS key-value store
	r, err := db.kv.Get(key)
	if err != nil {
		if err == nats.ErrKeyNotFound {
			// Return an error if the value is nil
			return []byte(""), hord.ErrNil
		}
		return []byte(""), fmt.Errorf("unable to fetch key - %s", err)
	}

	return r.Value(), nil
}

// Set inserts or updates data in the NATS database based on the provided key.
// It returns an error if the key or data is invalid.
func (db *Database) Set(key string, data []byte) error {
	// Validate the key
	if err := hord.ValidKey(key); err != nil {
		return err
	}

	// Validate the data
	if err := hord.ValidData(data); err != nil {
		return err
	}

	// Acquire a write lock to ensure data consistency during insertion/update
	db.Lock()
	defer db.Unlock()

	// Check if the NATS key-value store is initialized
	if db.kv == nil {
		return hord.ErrNoDial
	}

	// Insert or update the key-value pair in the NATS key-value store
	_, err := db.kv.Put(key, data)
	if err != nil {
		return fmt.Errorf("unable to set key - %s", err)
	}

	return nil
}

// Delete removes data from the NATS database based on the provided key.
// It returns an error if the key is invalid.
func (db *Database) Delete(key string) error {
	// Validate the key
	if err := hord.ValidKey(key); err != nil {
		return err
	}

	// Acquire a write lock to ensure data consistency during deletion
	db.Lock()
	defer db.Unlock()

	// Check if the NATS key-value store is initialized
	if db.kv == nil {
		return hord.ErrNoDial
	}

	// Delete the key from the NATS key-value store
	err := db.kv.Delete(key)
	if err != nil {
		return fmt.Errorf("unable to remove key - %s", err)
	}

	return nil
}

// Keys retrieves a list of keys stored in the NATS database.
func (db *Database) Keys() ([]string, error) {
	// Acquire a read lock to ensure data consistency during key retrieval
	db.RLock()
	defer db.RUnlock()

	// Check if the NATS key-value store is initialized
	if db.kv == nil {
		return []string{}, hord.ErrNoDial
	}

	// Retrieve the keys from the NATS key-value store
	keys, err := db.kv.Keys()
	if err != nil {
		// If no keys, return empty list
		if err == nats.ErrNoKeysFound {
			return []string{}, nil
		}
		return []string{}, fmt.Errorf("unable to fetch keys - %s", err)
	}

	return keys, nil
}

// HealthCheck performs a health check on the NATS database.
// Since the NATS database is an in-memory implementation, it always returns nil.
func (db *Database) HealthCheck() error {
	// Acquire a read lock to ensure data consistency during health check
	db.RLock()
	defer db.RUnlock()

	// Check if the NATS key-value store is initialized
	if db.kv == nil {
		return hord.ErrNoDial
	}

	// Check the status of the NATS key-value store
	status, err := db.kv.Status()
	if err != nil {
		return fmt.Errorf("kv store unhealthy - %s", err)
	}

	if status.Bucket() != db.bucket {
		return fmt.Errorf("kv store returned an unhealthy response")
	}

	return nil
}

// Close closes the NATS database connection and clears all stored data.
func (db *Database) Close() {
	// Acquire a write lock to ensure proper closing of the connection and clearing of data
	db.Lock()
	defer db.Unlock()

	// Check if the NATS server is connected
	if db.conn == nil {
		return
	}

	// Drain the NATS connection to close it gracefully
	err := db.conn.Drain()
	if err != nil {
		db.conn.Close()
	}
}
