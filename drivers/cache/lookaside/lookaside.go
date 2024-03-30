/*
Package lookaside provides a Hord database driver for a look-aside cache. To use this driver, import it as follows:

	import (
	    "github.com/madflojo/hord"
	    "github.com/madflojo/hord/cache/lookaside"
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
	db, err := lookaside.Dial(lookaside.Config{
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
	db, err := lookaside.Dial(lookaside.Config{
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
package lookaside

import (
	"errors"
	"fmt"

	"github.com/madflojo/hord"
)

// Config provides the configuration options for the Lookaside driver.
type Config struct {
	Database hord.Database
	Cache    hord.Database
}

// Lookaside is used to store data in a look-aside caching pattern. It also satisfies the Hord database interface.
type Lookaside struct {
	data  hord.Database
	cache hord.Database
}

func Dial(cfg Config) (hord.Database, error) {
	if (cfg.Database == nil) || (cfg.Cache == nil) {
		return nil, hord.ErrInvalidDatabase
	}

	return &Lookaside{
		data:  cfg.Database,
		cache: cfg.Cache,
	}, nil
}

// Setup will run the Setup function for both the database and the cache.
func (db *Lookaside) Setup() error {
	if db == nil || db.data == nil || db.cache == nil {
		return hord.ErrNoDial
	}

	if err := db.data.Setup(); err != nil {
		return err
	}

	if err := db.cache.Setup(); err != nil {
		return err
	}

	return nil
}

// HealthCheck will run the HealthCheck function for both the database and the cache.
func (db *Lookaside) HealthCheck() error {
	if db == nil || db.data == nil || db.cache == nil {
		return hord.ErrNoDial
	}

	dataErr := db.data.HealthCheck()
	cacheErr := db.cache.HealthCheck()

	if dataErr != nil {
		return dataErr
	} else if cacheErr != nil {
		return cacheErr
	}

	return nil
}

// Get will get the data from the cache database. If not found, it uses a look-aside pattern to fetch from the data database and store the data in the cache.
func (db *Lookaside) Get(key string) ([]byte, error) {
	if db == nil || db.data == nil || db.cache == nil {
		return nil, hord.ErrNoDial
	}

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

	// Update the cache
	err = db.cache.Set(key, data)
	if err != nil {
		return data, fmt.Errorf("%w: %w", hord.ErrCacheError, err)
	}

	return data, nil
}

// Set will set the data in both the data and cache databases.
func (db *Lookaside) Set(key string, data []byte) error {
	if db == nil || db.data == nil || db.cache == nil {
		return hord.ErrNoDial
	}

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
func (db *Lookaside) Delete(key string) error {
	if db == nil || db.data == nil || db.cache == nil {
		return hord.ErrNoDial
	}

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
func (db *Lookaside) Keys() ([]string, error) {
	if db == nil || db.data == nil || db.cache == nil {
		return nil, hord.ErrNoDial
	}

	return db.data.Keys()
}

// CacheKeys will return the keys from the cache database.
func (db *Lookaside) CacheKeys() ([]string, error) {
	if db == nil || db.data == nil || db.cache == nil {
		return nil, hord.ErrNoDial
	}

	return db.cache.Keys()
}

// GetCache will return the cache database.
func (db *Lookaside) GetCache() hord.Database {
	return db.cache
}

// GetDatabase will return the data database.
func (db *Lookaside) GetDatabase() hord.Database {
	return db.data
}

// Close will close the connections to both the database and the cache.
func (db *Lookaside) Close() {
	if db != nil && db.data != nil && db.cache != nil {
		db.data.Close()
		db.cache.Close()
	}
}
