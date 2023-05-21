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
package nats

import (
	"crypto/tls"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/madflojo/hord"
	"github.com/nats-io/nats.go"
)

// Config represents the configuration for the NATS database connection.
type Config struct {
	// URL specifies the URL to connect to the NATS server. This URL follows the format
	// of `nats://user:pass@example:8222` with supported protocols being `nats`,  `tls`, or `ws` for web sockets.
	URL string

	// Bucket name for the key-value store. If Bucket does not exist on the NATS server side,
	// NATS will automatically create the bucket with the first key creation. Bucket names must adhere
	// to the `^[a-zA-Z0-9_-]+$` regex.
	Bucket string

	// Servers enables connectivity to a cluster of NATS servers. Each entry must follow the NATS URL format.
	Servers []string

	// SkipTLSVerify will disable the TLS hostname checking. Warning, using this setting opens the risk of
	// man-in-the-middle attacks.
	SkipTLSVerify bool

	// TLSConfig allows users to specify TLS settings for connecting to NATS. This is a standard TLS configuration
	// and can be used to configure 2-way TLS for NATS.
	TLSConfig *tls.Config

	// Options extend the connection options available within NATS. NATS has many advanced configuration options;
	// use Options to modify those options.
	Options nats.Options
}

// Database is an in-memory NATS implementation of the hord.Database interface.
type Database struct {
	sync.RWMutex

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
	db := &Database{}

	// Validate Bucket
	if cfg.Bucket == "" || !reBucket.MatchString(cfg.Bucket) {
		return db, fmt.Errorf("Bucket name is invalid")
	}

	// Build URL for cluster of servers
	if cfg.URL == "" && len(cfg.Servers) < 1 {
		return db, fmt.Errorf("URL and Servers cannot be empty")
	}
	cfg.Options.Url = cfg.URL
	cfg.Options.Servers = cfg.Servers

	// Set TLS Config
	if cfg.TLSConfig != nil {
		cfg.Options.TLSConfig = cfg.TLSConfig
		cfg.Options.Secure = cfg.SkipTLSVerify
	}

	// Connect to the NATS server
	db.conn, err = cfg.Options.Connect()
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
	_, err := db.kv.Status()
	if err != nil {
		return fmt.Errorf("kv store unhealthy - %s", err)
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
