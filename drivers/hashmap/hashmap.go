/*
Package hashmap provides a Hord database driver for an in-memory hashmap.

The Hashmap driver is a simple, in-memory key-value store that stores data in a hashmap structure. To use this driver, import it as follows:

	import (
	    "github.com/madflojo/hord"
	    "github.com/madflojo/hord/hashmap"
	)

# Connecting to the Database

Use the Dial() function to create a new client for interacting with the hashmap driver.

	var db hord.Database
	db, err := hashmap.Dial(hashmap.Config{})
	if err != nil {
	    // Handle connection error
	}

# Initialize database

Hord provides a Setup() function for preparing a database. This function is safe to execute after every Dial().

	err := db.Setup()
	if err != nil {
	    // Handle setup error
	}

# Database Operations

Hord provides a simple abstraction for working with the hashmap driver, with easy-to-use methods such as Get() and Set() to read and write values.

	// Connect to the hashmap database
	db, err := hashmap.Dial(hashmap.Config{})
	if err != nil {
	    // Handle connection error
	}

	err := db.Setup()
	if err != nil {
	    // Handle setup error
	}

	// Set a value
	err = db.Set("key", []byte("value"))
	if err != nil {
	    // Handle error
	}

	// Retrieve a value
	value, err := db.Get("key")
	if err != nil {
	    // Handle error
	}
*/
package hashmap

import (
	"sync"

	"github.com/madflojo/hord"
)

// Config represents the configuration for the hashmap database.
type Config struct{}

// Database is an in-memory hashmap implementation of the hord.Database interface.
type Database struct {
	sync.RWMutex

	// data is used to store data in a simple map
	data map[string][]byte
}

// Dial initializes and returns a new hashmap database instance.
func Dial(_ Config) (*Database, error) {
	db := &Database{}
	db.data = make(map[string][]byte)
	return db, nil
}

// Setup sets up the hashmap database. This function does nothing for the hashmap driver.
func (db *Database) Setup() error {
	return nil
}

// Get retrieves data from the hashmap database based on the provided key.
// It returns the data associated with the key or an error if the key is invalid or the data does not exist.
func (db *Database) Get(key string) ([]byte, error) {
	if err := hord.ValidKey(key); err != nil {
		return []byte(""), err
	}

	db.RLock()
	defer db.RUnlock()
	if db.data == nil {
		return []byte(""), hord.ErrNoDial
	}

	v, ok := db.data[key]
	if ok {
		return v, nil
	}
	return []byte(""), hord.ErrNil
}

// Set inserts or updates data in the hashmap database based on the provided key.
// It returns an error if the key or data is invalid.
func (db *Database) Set(key string, data []byte) error {
	if err := hord.ValidKey(key); err != nil {
		return err
	}

	if err := hord.ValidData(data); err != nil {
		return err
	}

	db.Lock()
	defer db.Unlock()
	if db.data == nil {
		return hord.ErrNoDial
	}

	db.data[key] = data
	return nil
}

// Delete removes data from the hashmap database based on the provided key.
// It returns an error if the key is invalid.
func (db *Database) Delete(key string) error {
	if err := hord.ValidKey(key); err != nil {
		return err
	}

	db.Lock()
	defer db.Unlock()
	if db.data == nil {
		return hord.ErrNoDial
	}

	delete(db.data, key)
	return nil
}

// Keys retrieves a list of keys stored in the hashmap database.
func (db *Database) Keys() ([]string, error) {
	db.RLock()
	defer db.RUnlock()
	if db.data == nil {
		return []string{}, hord.ErrNoDial
	}

	var keys []string
	for k := range db.data {
		keys = append(keys, k)
	}
	return keys, nil
}

// HealthCheck performs a health check on the hashmap database.
// Since the hashmap database is an in-memory implementation, it always returns nil.
func (db *Database) HealthCheck() error {
	db.RLock()
	defer db.RUnlock()
	if db.data == nil {
		return hord.ErrNoDial
	}

	return nil
}

// Close closes the hashmap database connection and clears all stored data.
func (db *Database) Close() {
	db.Lock()
	defer db.Unlock()
	db.data = nil
}
