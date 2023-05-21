/*
Package bbolt provides a Hord database driver for BoltDB.

BoltDB is an embedded key-value database that persists data on disk. To use this driver, import it as follows:

	import (
	    "github.com/madflojo/hord"
	    "github.com/madflojo/hord/bbolt"
	)

# Connecting to the Database

Use the Dial() function to create a new client for interacting with BoltDB.

	var db hord.Database
	db, err := bbolt.Dial(bbolt.Config{})
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

Hord provides a simple abstraction for working with BoltDB, with easy-to-use methods such as Get() and Set() to read and write values.

	// Connect to the BoltDB database
	db, err := bbolt.Dial(bbolt.Config{})
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
package bbolt

import (
	"fmt"
	"os"
	"time"

	"github.com/madflojo/hord"
	"go.etcd.io/bbolt"
)

// Config represents the configuration for the bbolt database.
type Config struct {
	// Bucketname specifies the bucket to store and retrieve data from.
	Bucketname string

	// Filename specifies the file path of the bbolt file.
	Filename string

	// Permissions specifies the file permissions for the bbolt database file.
	Permissions os.FileMode

	// Timeout specifies the timeout duration for opening obtaining a file lock on the database file.
	// Default value is 5 Seconds, a value of 0 is invalid.
	Timeout time.Duration
}

// Database is an bbolt implementation of the hord.Database interface.
type Database struct {
	// cfg provides a reference to the dial configuration.
	cfg Config

	// db is the underlying database.
	db *bbolt.DB
}

// Dial initializes and returns a new bbolt database instance.
func Dial(cfg Config) (*Database, error) {
	var err error
	db := &Database{cfg: cfg}

	// Verify Bucket is set
	if cfg.Bucketname == "" {
		return db, fmt.Errorf("bucketname cannot be empty")
	}

	// Verify Filename is set
	if cfg.Filename == "" {
		return db, fmt.Errorf("filename must not be empty")
	}

	// Set Default Permissions
	if cfg.Permissions == 0 {
		cfg.Permissions = 0600
	}

	// Set Default Timeout
	if cfg.Timeout == time.Duration(0) {
		cfg.Timeout = time.Duration(5 * time.Second)
	}

	// Open database
	db.db, err = bbolt.Open(cfg.Filename, cfg.Permissions, &bbolt.Options{Timeout: cfg.Timeout})
	if err != nil {
		return db, fmt.Errorf("unable to open database - %s", err)
	}

	return db, nil
}

// Setup initializes the database by creating the necessary bucket if it doesn't exist.
// Returns an error if the database is not connected or if there is an error creating the bucket.
func (db *Database) Setup() error {
	// Verify DB is connected
	if db == nil || db.db == nil {
		return hord.ErrNoDial
	}

	// Open Bucket
	err := db.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(db.cfg.Bucketname))
		if err != nil {
			return fmt.Errorf("unable to open bucket - %s", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// Get retrieves data from the bbolt database based on the provided key.
// It returns the data associated with the key or an error if the key is invalid or the data does not exist.
func (db *Database) Get(key string) ([]byte, error) {
	if err := hord.ValidKey(key); err != nil {
		return nil, err
	}

	// Verify DB is connected
	if db == nil || db.db == nil {
		return nil, hord.ErrNoDial
	}

	var data []byte
	err := db.db.View(func(tx *bbolt.Tx) error {
		// Open Bucket for this Tx
		bucket := tx.Bucket([]byte(db.cfg.Bucketname))
		if bucket == nil {
			return fmt.Errorf("bucket does not exist")
		}

		// Fetch Data from Bucket
		d := bucket.Get([]byte(key))
		if d != nil {
			// Copy results into data as d will only be valid for the lifetime of this Tx
			data = append(data, d...)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error while executing Get - %s", err)
	}

	// If no data returned, return ErrNil
	if len(data) == 0 {
		return nil, hord.ErrNil
	}
	return data, nil
}

// Set inserts or updates data in the bbolt database based on the provided key.
// It returns an error if the key or data is invalid.
func (db *Database) Set(key string, data []byte) error {
	if err := hord.ValidKey(key); err != nil {
		return err
	}

	if err := hord.ValidData(data); err != nil {
		return err
	}

	// Verify DB is connected
	if db == nil || db.db == nil {
		return hord.ErrNoDial
	}

	err := db.db.Update(func(tx *bbolt.Tx) error {
		// Open Bucket for this Tx
		bucket := tx.Bucket([]byte(db.cfg.Bucketname))
		if bucket == nil {
			return fmt.Errorf("bucket does not exist")
		}

		// Store Data into Bucket
		err := bucket.Put([]byte(key), data)
		if err != nil {
			return fmt.Errorf("error while executing Set - %s", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error while executing Set transaction - %s", err)
	}

	return nil
}

// Delete removes data from the bbolt database based on the provided key.
// It returns an error if the key is invalid.
func (db *Database) Delete(key string) error {
	if err := hord.ValidKey(key); err != nil {
		return err
	}

	// Verify DB is connected
	if db == nil || db.db == nil {
		return hord.ErrNoDial
	}

	err := db.db.Update(func(tx *bbolt.Tx) error {
		// Open Bucket for this Tx
		bucket := tx.Bucket([]byte(db.cfg.Bucketname))
		if bucket == nil {
			return fmt.Errorf("bucket does not exist")
		}

		// Delete Key
		err := bucket.Delete([]byte(key))
		if err != nil {
			return fmt.Errorf("error while executing Delete - %s", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error while executing Delete transaction - %s", err)
	}

	return nil
}

// Keys retrieves a list of keys stored in the bbolt database.
func (db *Database) Keys() ([]string, error) {
	// Verify DB is connected
	if db == nil || db.db == nil {
		return nil, hord.ErrNoDial
	}

	var keys []string
	err := db.db.View(func(tx *bbolt.Tx) error {
		// Open Bucket for this Tx
		bucket := tx.Bucket([]byte(db.cfg.Bucketname))
		if bucket == nil {
			return fmt.Errorf("bucket does not exist")
		}

		// Loop through keys in bucket and return a list of them
		err := bucket.ForEach(func(k, _ []byte) error {
			keys = append(keys, string(k))
			return nil
		})
		if err != nil {
			return fmt.Errorf("error while executing Keys - %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error while executing Keys transaction - %s", err)
	}

	return keys, nil
}

// HealthCheck performs a health check on the bbolt database.
func (db *Database) HealthCheck() error {
	// Verify DB is connected
	if db == nil || db.db == nil {
		return hord.ErrNoDial
	}

	err := db.db.View(func(tx *bbolt.Tx) error {
		// Open Bucket for this Tx
		bucket := tx.Bucket([]byte(db.cfg.Bucketname))
		if bucket == nil {
			return fmt.Errorf("bucket does not exist")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error while checking database health - %s", err)
	}

	return nil
}

// Close closes the bbolt database connection and clears all stored data.
func (db *Database) Close() {
	// Verify DB is connected
	if db == nil || db.db == nil {
		return
	}

	// Close DB
	err := db.db.Close()
	if err != nil {
		return
	}
}
