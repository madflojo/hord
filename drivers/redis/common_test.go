package redis

import (
	"fmt"
	"testing"
	"time"

	"github.com/madflojo/hord"
)

func TestInterfaceHappyPath(t *testing.T) {
	cfgs := make(map[string]Config)
	cfgs["Simple Redis"] = Config{
		ConnectTimeout: time.Duration(5) * time.Second,
		Server:         "redis:6379",
	}
	cfgs["Redis with Sentinel"] = Config{
		ConnectTimeout: time.Duration(5) * time.Second,
		SentinelConfig: SentinelConfig{
			Servers: []string{"redis-sentinel:26379"},
			Master:  "mymaster",
		},
	}
	cfgs["Redis with optimized settings"] = Config{
		ConnectTimeout: time.Duration(5) * time.Second,
		MaxActive:      500,
		MaxIdle:        100,
		IdleTimeout:    time.Duration(5) * time.Second,
		Server:         "redis:6379",
	}
	cfgs["Dragonfly with optimized settings"] = Config{
		ConnectTimeout: time.Duration(5) * time.Second,
		MaxActive:      500,
		MaxIdle:        100,
		IdleTimeout:    time.Duration(5) * time.Second,
		Server:         "dragonfly:6379",
	}
	cfgs["KeyDB with optimized settings"] = Config{
		ConnectTimeout: time.Duration(5) * time.Second,
		MaxActive:      500,
		MaxIdle:        100,
		IdleTimeout:    time.Duration(5) * time.Second,
		Server:         "keydb:6379",
	}

	// Loop through valid Configs and validate the driver adheres to the Hord interface
	for name, cfg := range cfgs {
		t.Run(name, func(t *testing.T) {
			// Establish Connectivity
			db, err := Dial(cfg)
			if err != nil {
				t.Fatalf("Failed to connect to database - %s", err)
			}
			defer db.Close()

			// Setup Database
			t.Run("Setup Database", func(t *testing.T) {
				err := db.Setup()
				if err != nil {
					t.Errorf("Failed to execute Setup - %s", err)
				}
				<-time.After(1 * time.Second)
			})

			// Perform HealthCheck
			t.Run("Validate Database Health", func(t *testing.T) {
				err = db.HealthCheck()
				if err != nil {
					t.Fatalf("Unexpected error when performing health check - %s", err)
				}
			})

			// Single Key Execution
			t.Run("Single Key Execution", func(t *testing.T) {

				// Clear Database when done
				t.Cleanup(func() {
					keys, err := db.Keys()
					if err != nil {
						t.Fatalf("Unexecpted error when obtaining a list of keys from the Redis - %s", err)
					}

					for _, k := range keys {
						_ = db.Delete(k)
					}
				})

				// No Keys
				t.Run("No Keys", func(t *testing.T) {
					keys, err := db.Keys()
					if err != nil {
						t.Fatalf("Unexecpted error when obtaining a list of keys from the Redis - %s", err)
					}

					if len(keys) > 0 {
						t.Fatalf("Unexpected keys found in key list got - %+v", keys)
					}
				})

				// Get a Missing Key
				t.Run("Get Missing Key", func(t *testing.T) {
					_, err := db.Get("404notfound")
					if err == nil && err != hord.ErrNil {
						t.Errorf("Expected ErrNil when looking up nonexistent key - %s", err)
					}
				})

				// Delete a Missing Key
				t.Run("Delete Missing Key", func(t *testing.T) {
					err := db.Delete("404notfound")
					if err != nil {
						t.Errorf("Expected nil when deleting nonexistent key - %s", err)
					}
				})

				// Set a Key
				t.Run("Set a Key", func(t *testing.T) {
					err := db.Set("test_key", []byte("Testing"))
					if err != nil {
						t.Errorf("Unexpected error when writing data - %s", err)
					}
				})

				// Get a Key
				t.Run("Get a Key", func(t *testing.T) {
					data, err := db.Get("test_key")
					if err != nil {
						t.Fatalf("Unexpected error when reading data - %s", err)
					}

					if string(data) != "Testing" {
						t.Errorf("Data mismatch from previously set data and fetched data got %+v expected %+v", data, []byte("Testing"))
					}
				})

				// Get list of Keys
				t.Run("Get a list of Keys", func(t *testing.T) {
					keys, err := db.Keys()
					if err != nil {
						t.Fatalf("Unexpected error when fetching keys - %s", err)
					}

					if len(keys) != 1 {
						t.Errorf("Unexpected number of returned keys - got %d, expected 1", len(keys))
					}
				})

				// Delete a Key
				t.Run("Delete a Key", func(t *testing.T) {
					err := db.Delete("test_key")
					if err != nil {
						t.Fatalf("Unexpected error when deleting data - %s", err)
					}

					data, err := db.Get("test_key")
					if err != hord.ErrNil && len(data) != 0 {
						t.Errorf("It does not appear data was completely deleted - %+v", data)
					}
				})

				// Set a Invalid Key
				t.Run("Set a Invalid Key", func(t *testing.T) {
					err := db.Set("", []byte("Testing"))
					if err == nil || err != hord.ErrInvalidKey {
						t.Errorf("Expected ErrInvalidKey when using blank key")
					}
				})

				// Get a Invalid Key
				t.Run("Get a Invalid Key", func(t *testing.T) {
					_, err := db.Get("")
					if err == nil || err != hord.ErrInvalidKey {
						t.Errorf("Expected ErrInvalidKey when using blank key")
					}
				})

				// Delete a Invalid Key
				t.Run("Delete a Invalid Key", func(t *testing.T) {
					err := db.Delete("")
					if err == nil || err != hord.ErrInvalidKey {
						t.Errorf("Expected ErrInvalidKey when using blank key")
					}
				})

				// Set with Invalid Data
				t.Run("Set with Invalid Data", func(t *testing.T) {
					err := db.Set("test_key", []byte(""))
					if err == nil || err != hord.ErrInvalidData {
						t.Errorf("Expected ErrInvalidData when using blank data")
					}
				})

			})

			// Lots of Keys Execution
			t.Run("Multiple Key Execution", func(t *testing.T) {
				// Clear Database when done
				t.Cleanup(func() {
					keys, err := db.Keys()
					if err != nil {
						t.Fatalf("Unexecpted error when obtaining a list of keys from the Redis - %s", err)
					}

					for _, k := range keys {
						_ = db.Delete(k)
					}
				})

				// Create a ton of keys
				t.Run("Create 10 keys", func(t *testing.T) {
					for i := 0; i < 10; i++ {
						err := db.Set(fmt.Sprintf("Testing 1000 keys with key number %d", i), []byte("Testing"))
						if err != nil {
							t.Fatalf("Error setting up test keys - %s", err)
						}
					}
				})

				// Count Keys
				t.Run("Ensure 10 keys exist", func(t *testing.T) {
					keys, err := db.Keys()
					if err != nil {
						t.Fatalf("Error fetcing keys from database - %s", err)
					}

					if len(keys) != 10 {
						t.Errorf("Invalid Number of Keys returned %d", len(keys))
					}
				})

			})

			t.Run("Closed DB Execution", func(t *testing.T) {

				db.Close()

				// Perform HealthCheck
				t.Run("Validate Database Health", func(t *testing.T) {
					err = db.HealthCheck()
					if err == nil {
						t.Errorf("Unexpected success when performing task on closed database - %s", err)
					}
				})

				// Single Key Execution
				t.Run("Single Key Execution", func(t *testing.T) {
					// Set a Key
					t.Run("Set a Key", func(t *testing.T) {
						err := db.Set("test_key", []byte("Testing"))
						if err == nil {
							t.Errorf("Unexpected success when performing task on closed database - %s", err)
						}
					})

					// Get a Key
					t.Run("Get a Key", func(t *testing.T) {
						_, err := db.Get("test_key")
						if err == nil {
							t.Errorf("Unexpected success when performing task on closed database - %s", err)
						}
					})

					// Get list of Keys
					t.Run("Get a list of Keys", func(t *testing.T) {
						_, err := db.Keys()
						if err == nil {
							t.Errorf("Unexpected success when performing task on closed database - %s", err)
						}
					})

					// Delete a Key
					t.Run("Delete a Key", func(t *testing.T) {
						err := db.Delete("test_key")
						if err == nil {
							t.Errorf("Unexpected success when performing task on closed database - %s", err)
						}
					})

				})
			})

		})
	}
}

func TestInterfaceFail(t *testing.T) {
	cfgs := make(map[string]Config)
	cfgs["Wrong Redis Port"] = Config{
		ConnectTimeout: time.Duration(5) * time.Second,
		Server:         "redis:9000",
	}
	cfgs["Sentinel without Master"] = Config{
		ConnectTimeout: time.Duration(5) * time.Second,
		SentinelConfig: SentinelConfig{
			Servers: []string{"redis-sentinel:26379"},
		}}

	// Loop through invalid Configs and validate the driver reacts appropriately
	for name, cfg := range cfgs {
		t.Run(name, func(t *testing.T) {
			// Establish Connectivity
			db, err := Dial(cfg)
			if err == nil {
				t.Errorf("Expected error when connecting to database but got no error...")
			}
			defer db.Close()

			// Setup Database
			t.Run("Setup Database", func(t *testing.T) {
				err := db.Setup()
				if err == nil {
					t.Errorf("Expected error when attempting to setup database without connection...")
				}
			})

			// Perform HealthCheck
			t.Run("Validate Database Health", func(t *testing.T) {
				err = db.HealthCheck()
				if err == nil {
					t.Errorf("Expected error when attempting to healthcheck database without connection...")
				}
			})

			// Single Key Execution
			t.Run("Single Key Execution", func(t *testing.T) {

				// Clear Database when done
				t.Cleanup(func() {
					keys, _ := db.Keys()
					for _, k := range keys {
						_ = db.Delete(k)
					}
				})

				// Set a Key
				t.Run("Set a Key", func(t *testing.T) {
					err := db.Set("test_key", []byte("Testing"))
					if err == nil {
						t.Errorf("Expected error when using data with no connection...")
					}
				})
				// Get a Key
				t.Run("Get a Key", func(t *testing.T) {
					_, err := db.Get("test_key")
					if err == nil {
						t.Errorf("Expected error when using data with no connection...")
					}
				})

				// Get list of Keys
				t.Run("Get a list of Keys", func(t *testing.T) {
					keys, err := db.Keys()
					if err == nil {
						t.Errorf("Expected error when using data with no connection...")
					}
					if len(keys) != 0 {
						t.Errorf("Unexpected number of returned keys - got %d, expected 0", len(keys))
					}
				})

				// Delete a Key
				t.Run("Delete a Key", func(t *testing.T) {
					err := db.Delete("test_key")
					if err == nil {
						t.Errorf("Expected error when using data with no connection...")
					}
				})
			})
		})
	}
}
