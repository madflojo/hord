// Package hashmap is a Hord database driver that creates a hashmap-based in-memory key-value store.
//
//	// Connect to Hashmap
//	db, err := hashmap.Dial(&hashmap.Config{})
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
// The hashmap driver in the Hord package allows you to quickly work with an embedded in-memory hashmap in your Go
// applications. It provides methods to store and retrieve key-value pairs, enabling efficient goroutine safe data
// management.
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
