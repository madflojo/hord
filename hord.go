// Package hord provides a modular key-value interface for interacting with databases. The goal is to provide a
// consistent interface regardless, of the underlying database.
//
// With this package, users can switch out the underlying database without major refactoring.
//
// The below example shows using Hord to connect and interact with Cassandra.
//
//	import "github.com/madflojo/hord"
//	import "github.com/madflojo/hord/driver/cassandra"
//
//	func main() {
//	  // Define our DB Interface
//	  var db hord.Database
//
//	  // Connect to a Cassandra Cluster
//	  db, err := cassandra.Dial(&cassandra.Config{})
//	  if err != nil {
//	    // do stuff
//	  }
//
//	  // Setup and Initialize the Keyspace if necessary
//	  err = db.Setup()
//	  if err != nil {
//	    // do stuff
//	  }
//
//	  // Write data to the cluster
//	  err = db.Set("mykey", []byte("My Data"))
//	  if err != nil {
//	    // do stuff
//	  }
//
//	  // Fetch the same data
//	  d, err := db.Get("mykey")
//	  if err != nil {
//	    // do stuff
//	  }
//	}
package hord

import "fmt"

// Database is an interface that is used to create a unified database access object.
type Database interface {
	// Setup is used to setup and configure the underlying database.
	// This can include setting optimal cluster settings, creating a database or tablespace,
	// or even creating the database structure.
	// Setup is meant to allow users to start with a fresh database service and turn it into a production-ready datastore.
	Setup() error

	// HealthCheck performs a check against the underlying database.
	// If any errors are returned, this health check will return an error.
	// An error returned from HealthCheck should be treated as the database service being untrustworthy.
	HealthCheck() error

	// Get is used to fetch data with the provided key.
	Get(key string) ([]byte, error)

	// Set is used to insert and update the specified key.
	// This function can be used on existing keys, with the new data overwriting existing data.
	Set(key string, data []byte) error

	// Delete will delete the data for the specified key.
	Delete(key string) error

	// Keys will return a list of keys for the entire database.
	// This operation can be expensive, use with caution.
	Keys() ([]string, error)

	// Close will close the database connection.
	// After executing close, all other functions should return an error.
	Close()
}

// Common Errors Used by Hord Drivers
var (
	ErrInvalidKey  = fmt.Errorf("Key cannot be nil")
	ErrInvalidData = fmt.Errorf("Data cannot be empty")
	ErrNil         = fmt.Errorf("Nil value returned from database")
	ErrNoDial      = fmt.Errorf("No database connection defined, did you dial?")
)

// ValidKey checks if a key is valid.
// A valid key should have a length greater than 0.
// Returns nil if the key is valid, otherwise returns ErrInvalidKey.
func ValidKey(key string) error {
	if len(key) > 0 {
		return nil
	}
	return ErrInvalidKey
}

// ValidData checks if data is valid.
// Valid data should have a length greater than 0.
// Returns nil if the data is valid, otherwise returns ErrInvalidData.
func ValidData(data []byte) error {
	if len(data) > 0 {
		return nil
	}
	return ErrInvalidData
}
