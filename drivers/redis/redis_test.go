package redis

import (
	"crypto/tls"
	"fmt"
	"testing"
	"time"
)

func TestConnectivity(t *testing.T) {
	t.Run("No Config", func(t *testing.T) {
		_, err := Dial(Config{})
		if err == nil {
			t.Errorf("Expected error when Dialing with no config set, got nil")
		}
	})

	t.Run("Just Redis", func(t *testing.T) {
		db, err := Dial(Config{
			ConnectTimeout: time.Duration(5) * time.Second,
			Server:         "redis:6379",
		})
		if err != nil {
			t.Fatalf("Failed to connect to Redis - %s", err)
		}
		defer db.Close()

		// Test a connection
		c := db.pool.Get()
		defer c.Close()

		_, err = c.Do("PING")
		if err != nil {
			t.Errorf("Failed to ping Redis server - %s", err)
		}
	})

	t.Run("Fake TLS", func(t *testing.T) {
		_, _ = Dial(Config{
			ConnectTimeout: time.Duration(5) * time.Second,
			Server:         "redis:6379",
			TLSConfig:      &tls.Config{},
		})
	})

	t.Run("Sentinel Connection No Master", func(t *testing.T) {
		db, err := Dial(Config{
			ConnectTimeout: time.Duration(5) * time.Second,
			SentinelConfig: SentinelConfig{
				Servers: []string{"redis-sentinel:26379"},
			},
		})
		if err == nil {
			defer db.Close()
			t.Fatalf("Failed to connect to Redis via Sentinel - %s", err)
		}
	})

	t.Run("Sentinel Connection", func(t *testing.T) {
		db, err := Dial(Config{
			ConnectTimeout: time.Duration(5) * time.Second,
			SentinelConfig: SentinelConfig{
				Servers: []string{"redis-sentinel:26379"},
				Master:  "mymaster",
			},
		})
		if err != nil {
			t.Fatalf("Failed to connect to Redis via Sentinel - %s", err)
		}
		defer db.Close()

		// Test a connection
		c := db.pool.Get()
		defer c.Close()

		_, err = c.Do("PING")
		if err != nil {
			t.Errorf("Failed to ping Redis server - %s", err)
		}

		// Check TestOnBorrow
		err = db.pool.TestOnBorrow(c, time.Now())
		if err != nil {
			t.Errorf("Error returned when testing pool connection - %s", err)
		}
	})
}

func TestUsage(t *testing.T) {
	// Setup Environment
	db, err := Dial(Config{
		ConnectTimeout: time.Duration(5) * time.Second,
		Server:         "redis:6379",
	})
	if err != nil {
		t.Fatalf("Failed to connect to Redis - %s", err)
	}
	defer db.Close()

	t.Run("Setup", func(t *testing.T) {
		err := db.Setup()
		if err != nil {
			t.Errorf("Failed to execute Setup - %s", err)
		}
	})

	t.Run("Writing data", func(t *testing.T) {
		data := []byte("Testing")
		err := db.Set("test_happypath", data)
		if err != nil {
			t.Fatalf("Unexpected error when writing data - %s", err)
		}
	})

	t.Run("Writing Empty data", func(t *testing.T) {
		err := db.Set("test_emptydata", []byte(""))
		if err == nil {
			t.Errorf("Expected Error when writing with an empty byte slice, got nil")
		}
	})

	t.Run("Reading data", func(t *testing.T) {
		data, err := db.Get("test_happypath")
		if err != nil {
			t.Fatalf("Unexpected error when reading data - %s", err)
		}

		for i, v := range []byte("Testing") {
			if v != data[i] {
				t.Fatalf("Data mismatch from previously set data and data just read, got %+v expected %+v", data[i], v)
			}
		}
	})

	t.Run("Deleting data", func(t *testing.T) {
		err := db.Delete("test_happypath")
		if err != nil {
			t.Fatalf("Unexpected error when deleting data - %s", err)
		}

		data, err := db.Get("test_happypath")
		if err == nil {
			t.Fatalf("It does not appear data was completely deleted from table found - %+v", data)
		}
	})
}

func TestHealthCheck(t *testing.T) {
	// Setup Environment
	db, err := Dial(Config{
		ConnectTimeout: time.Duration(5) * time.Second,
		Server:         "redis:6379",
	})
	if err != nil {
		t.Fatalf("Failed to connect to Redis - %s", err)
	}
	defer db.Close()

	err = db.Setup()
	if err != nil {
		t.Fatalf("Got unexpected error when initializing Redis - %s", err)
	}
	time.Sleep(1 * time.Second)

	err = db.HealthCheck()
	if err != nil {
		t.Fatalf("Unexpected error when performing health check against Redis - %s", err)
	}
}

func TestKeys(t *testing.T) {
	// Setup Environment
	db, err := Dial(Config{
		ConnectTimeout: time.Duration(5) * time.Second,
		Server:         "redis:6379",
	})
	if err != nil {
		t.Fatalf("Failed to connect to Redis - %s", err)
	}
	defer db.Close()

	err = db.Setup()
	if err != nil {
		t.Fatalf("Got unexpected error when initializing Redis - %s", err)
	}
	time.Sleep(1 * time.Second)

	t.Run("Clean up Keys with Keys", func(t *testing.T) {
		keys, err := db.Keys()
		if err != nil {
			t.Fatalf("Unexecpted error when obtaining a list of keys from the Redis - %s", err)
		}

		for _, k := range keys {
			_ = db.Delete(k)
		}
	})

	t.Run("No Keys", func(t *testing.T) {
		keys, err := db.Keys()
		if err != nil {
			t.Fatalf("Unexecpted error when obtaining a list of keys from the Redis - %s", err)
		}

		if len(keys) > 0 {
			t.Fatalf("Unexpected keys found in key list got - %+v", keys)
		}
	})

	t.Run("5 keys", func(t *testing.T) {
		// Setup
		data := []byte("Testing")
		for i := 0; i < 5; i++ {
			err := db.Set(fmt.Sprintf("Testing Keys %d", i), data)
			if err != nil {
				t.Fatalf("Error setting up test keys for testcase - %s", err)
			}
		}
		time.Sleep(5 * time.Second)

		keys, err := db.Keys()
		if err != nil {
			t.Fatalf("Unexecpted error when obtaining a list of keys from the Redis - %s", err)
		}

		if len(keys) != 5 {
			t.Fatalf("Unexpected number of keys found in key list got - %+v", keys)
		}

		// Tear down
		for i := 0; i < 5; i++ {
			_ = db.Delete(fmt.Sprintf("Testing Keys %d", i))
		}
	})
}

func TestBlanks(t *testing.T) {
	// Setup Environment
	db, err := Dial(Config{
		ConnectTimeout: time.Duration(5) * time.Second,
		Server:         "redis:6379",
	})
	if err != nil {
		t.Fatalf("Failed to connect to Redis - %s", err)
	}
	defer db.Close()

	err = db.Setup()
	if err != nil {
		t.Fatalf("Got unexpected error when initializing Redis - %s", err)
	}

	t.Run("GET No Key", func(t *testing.T) {
		_, err := db.Get("")
		if err == nil {
			t.Errorf("Expected error when using blank key")
		}
	})

	t.Run("SET No Key", func(t *testing.T) {
		err := db.Set("", []byte("Testing"))
		if err == nil {
			t.Errorf("Expected error when using blank key")
		}
	})

	t.Run("SET No Data", func(t *testing.T) {
		err := db.Set("Testing", []byte(""))
		if err == nil {
			t.Errorf("Expected error when using blank data")
		}
	})

	t.Run("DELETE No Key", func(t *testing.T) {
		err := db.Delete("")
		if err == nil {
			t.Errorf("Expected error when using blank key")
		}
	})
}
