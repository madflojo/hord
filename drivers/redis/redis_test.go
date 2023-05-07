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
