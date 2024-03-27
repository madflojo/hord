/*
Package cache provides a Hord database driver for a look-aside cache. To use this driver, import it as follows:

	import (
	    "github.com/madflojo/hord"
	    "github.com/madflojo/hord/cache"
	)

# Connecting to the Database

Use the Dial() function to create a new client for interacting with the cache.

	// Handle database connection
	var database hord.Database
	...

	// Handle cache connection
	var cache hord.Database
	...

	var db hord.Database
	db, err := cache.Dial(cache.Config{
		Database: database,
		Cache: 	  cache,
	})
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

Hord provides a simple abstraction for working with the cache, with easy-to-use methods such as Get() and Set() to read and write values.

	// Handle database connection
	var database hord.Database
	database, err := cassandra.Dial(cassandra.Config{})
	if err != nil {
		// Handle connection error
	}

	// Handle cache connection
	var cache hord.Database
	cache, err := redis.Dial(redis.Config{})
	if err != nil {
		// Handle connection error
	}

	// Connect to the Cache database
	db, err := cache.Dial(cache.Config{
		Database: database,
		Cache: 	  cache,
	})
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

package cache

import (
	"errors"

	"github.com/madflojo/hord"
)

// Config provides the configuration options for the Cache driver.
type Config struct {
	Database hord.Database
	Cache    hord.Database
}

// Cache is used to store data in a look-aside caching pattern. It also satisfies the Hord database interface.
type Cache struct {
	data  hord.Database
	cache hord.Database
}

var (
	// ErrInvalidDatabase is returned when a database is nil.
	ErrInvalidDatabase = errors.New("database cannot be nil")

	// ErrInvalidCache is returned when a cache is nil.
	ErrInvalidCache = errors.New("cache cannot be nil")
)

// Dial will create a new Cache driver using the provided Config. It will return an error if either the Database or Cache values in Config are nil.
func Dial(cfg Config) (*Cache, error) {
	if cfg.Database == nil {
		return nil, ErrInvalidDatabase
	}

	if cfg.Cache == nil {
		return nil, ErrInvalidCache
	}

	return &Cache{
		data:  cfg.Database,
		cache: cfg.Cache,
	}, nil
}

// Setup will run the Setup function for both the database and the cache.
func (db *Cache) Setup() error {
	if err := db.data.Setup(); err != nil {
		return err
	}

	if err := db.cache.Setup(); err != nil {
		return err
	}

	return nil
}

// HealthCheck will run the HealthCheck function for both the database and the cache.
func (db *Cache) HealthCheck() error {
	dataErr := db.data.HealthCheck()
	cacheErr := db.cache.HealthCheck()

	if dataErr != nil {
		return dataErr
	} else if cacheErr != nil {
		return cacheErr
	}

	return nil
}

// Get will get the data from the database. It uses a look-aside pattern to store the data in the cache if it is not already there.
func (db *Cache) Get(key string) ([]byte, error) {
	// Check the cache first
	data, err := db.cache.Get(key)
	if (err != nil) && !errors.Is(err, hord.ErrNil) {
		return nil, err
	} else if !errors.Is(err, hord.ErrNil) {
		return data, nil
	}

	// Check the data database
	data, err = db.data.Get(key)
	if err != nil {
		return nil, err
	}
	db.cache.Set(key, data)

	return data, nil
}

// Set will set the data in both the data and cache databases.
func (db *Cache) Set(key string, data []byte) error {
	err := db.data.Set(key, data)
	if err != nil {
		return err
	}

	// Update cache only if database Set was successful
	err = db.cache.Set(key, data)
	if err != nil {
		return err
	}

	return nil
}

// Delete will delete the data from both the data and cache databases.
func (db *Cache) Delete(key string) error {
	dataErr := db.data.Delete(key)
	cacheErr := db.cache.Delete(key)

	if dataErr != nil {
		return dataErr
	} else if cacheErr != nil {
		return cacheErr
	}

	return nil
}

// Keys will return the keys from the data database.
func (db *Cache) Keys() ([]string, error) {
	return db.data.Keys()
}

// Close will close the connections to both the database and the cache.
func (db *Cache) Close() {
	db.data.Close()
	db.cache.Close()
}
