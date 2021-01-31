package mock

import (
	"fmt"
	"github.com/madflojo/hord"
	"testing"
)

func TestDefaults(t *testing.T) {
	var db hord.Database
	db, err := Dial(Config{})
	if err != nil {
		t.Errorf("Unexepected error when creating Mock interface - %s", err)
	}
	defer db.Close()

	t.Run("Validate Setup", func(t *testing.T) {
		err := db.Setup()
		if err != nil {
			t.Errorf("Setup mocked function did not work as expected err returned - %s", err)
		}
	})

	t.Run("Validate HealthCheck", func(t *testing.T) {
		err := db.HealthCheck()
		if err != nil {
			t.Errorf("HealthCheck mocked function did not work as expected err returned - %s", err)
		}
	})

	t.Run("Validate Get", func(t *testing.T) {
		r, err := db.Get("works")
		if err != nil || len(r) != 0 {
			t.Errorf("Get mocked function did not work as expected err returned - %s", err)
		}
	})

	t.Run("Validate Set", func(t *testing.T) {
		err := db.Set("works", []byte{})
		if err != nil {
			t.Errorf("Set mocked function did not work as expected err returned - %s", err)
		}
	})

	t.Run("Validate Delete", func(t *testing.T) {
		err := db.Delete("works")
		if err != nil {
			t.Errorf("Delete mocked function did not work as expected err returned - %s", err)
		}
	})

	t.Run("Validate Keys", func(t *testing.T) {
		keys, err := db.Keys()
		if err != nil {
			t.Errorf("Keys mocked function did not work as expected err returned - %s", err)
		}
		if len(keys) != 0 {
			t.Errorf("Keys mocked function did not work as expected returned %d values", len(keys))
		}
	})
}

func TestMocking(t *testing.T) {
	var db hord.Database
	cfg := Config{
		// Create a fake Setup function
		SetupFunc: func() error {
			return fmt.Errorf("This is an error")
		},
		// Create a fake HealthCheck function
		HealthCheckFunc: func() error {
			return fmt.Errorf("This is an error")
		},
		// Create a fake GET function
		GetFunc: func(key string) ([]byte, error) {
			if key == "works" {
				return []byte("Yes"), nil
			}
			return []byte{}, hord.ErrNil
		},
		// Create a fake SET function
		SetFunc: func(key string, data []byte) error {
			if key == "works" {
				return nil
			}
			return fmt.Errorf("Error inserting data")
		},
		// Create a fake Delete function
		DeleteFunc: func(key string) error {
			if key == "works" {
				return nil
			}
			return fmt.Errorf("Error deleting data")
		},
		// Create a fake Keys function
		KeysFunc: func() ([]string, error) {
			return []string{"key1", "key2"}, nil
		},
	}

	db, err := Dial(cfg)
	if err != nil {
		t.Errorf("Unexpected error when creating Mock interface - %s", err)
	}
	defer db.Close()

	t.Run("Validate Setup", func(t *testing.T) {
		err := db.Setup()
		if err == nil {
			t.Errorf("Setup mocked function did not work as expected err returned - %s", err)
		}
	})

	t.Run("Validate HealthCheck", func(t *testing.T) {
		err := db.HealthCheck()
		if err == nil {
			t.Errorf("HealthCheck mocked function did not work as expected err returned - %s", err)
		}
	})

	t.Run("Validate Get", func(t *testing.T) {
		_, err := db.Get("works")
		if err != nil {
			t.Errorf("Get mocked function did not work as expected err returned - %s", err)
		}
	})

	t.Run("Validate Get Errors", func(t *testing.T) {
		_, err := db.Get("doesntwork")
		if err != hord.ErrNil {
			t.Errorf("Get mocked function did not work as expected err returned - %s", err)
		}
	})

	t.Run("Validate Set", func(t *testing.T) {
		err := db.Set("works", []byte{})
		if err != nil {
			t.Errorf("Set mocked function did not work as expected err returned - %s", err)
		}
	})

	t.Run("Validate Set Errors", func(t *testing.T) {
		err := db.Set("doesntwork", []byte{})
		if err == nil {
			t.Errorf("Set mocked function did not work as expected err returned - %s", err)
		}
	})

	t.Run("Validate Delete", func(t *testing.T) {
		err := db.Delete("works")
		if err != nil {
			t.Errorf("Delete mocked function did not work as expected err returned - %s", err)
		}
	})

	t.Run("Validate Delete Errors", func(t *testing.T) {
		err := db.Delete("doesntwork")
		if err == nil {
			t.Errorf("Delete mocked function did not work as expected err returned - %s", err)
		}
	})

	t.Run("Validate Keys", func(t *testing.T) {
		keys, err := db.Keys()
		if err != nil {
			t.Errorf("Keys mocked function did not work as expected err returned - %s", err)
		}
		if len(keys) != 2 {
			t.Errorf("Keys mocked function did not work as expected returned %d values", len(keys))
		}
	})

}
