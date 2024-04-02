package cache

import (
	"errors"
	"testing"

	"github.com/madflojo/hord"
	"github.com/madflojo/hord/drivers/mock"
)

func TestDial(t *testing.T) {
	unitTests := map[string]struct {
		config        Config
		expectedError error
	}{
		"No Config": {
			config:        Config{},
			expectedError: hord.ErrInvalidDatabase,
		},
		"No Database": {
			config: Config{
				Cache: &mock.Database{},
			},
			expectedError: hord.ErrInvalidDatabase,
		},
		"No Cache": {
			config: Config{
				Database: &mock.Database{},
			},
			expectedError: hord.ErrInvalidDatabase,
		},
		"Invalid Type": {
			config: Config{
				Type:     "invalid",
				Database: &mock.Database{},
				Cache:    &mock.Database{},
			},
			expectedError: ErrNoType,
		},
		"Type: Lookaside": {
			config: Config{
				Type:     Lookaside,
				Database: &mock.Database{},
				Cache:    &mock.Database{},
			},
			expectedError: nil,
		},
		"Type: None": {
			config: Config{
				Type:     None,
				Database: &mock.Database{},
				Cache:    &mock.Database{},
			},
			expectedError: nil,
		},
	}

	for name, test := range unitTests {
		t.Run(name, func(t *testing.T) {
			_, err := Dial(test.config)
			if !errors.Is(err, test.expectedError) {
				t.Errorf("Dial(%v) returned error: %s, expected %s", test.config, err, test.expectedError)
			}
		})
	}
}

func TestNilCache(t *testing.T) {
	nc := &NilCache{}
	if err := nc.Setup(); !errors.Is(err, hord.ErrNoDial) {
		t.Errorf("NilCache.Setup() returned error: %s, expected %s", err, hord.ErrNoDial)
	}
	if err := nc.HealthCheck(); !errors.Is(err, hord.ErrNoDial) {
		t.Errorf("NilCache.HealthCheck() returned error: %s, expected %s", err, hord.ErrNoDial)
	}
	if _, err := nc.Get(""); !errors.Is(err, hord.ErrNoDial) {
		t.Errorf("NilCache.Get() returned error: %s, expected %s", err, hord.ErrNoDial)
	}
	if err := nc.Set("", nil); !errors.Is(err, hord.ErrNoDial) {
		t.Errorf("NilCache.Set() returned error: %s, expected %s", err, hord.ErrNoDial)
	}
	if err := nc.Delete(""); !errors.Is(err, hord.ErrNoDial) {
		t.Errorf("NilCache.Delete() returned error: %s, expected %s", err, hord.ErrNoDial)
	}
	if _, err := nc.Keys(); !errors.Is(err, hord.ErrNoDial) {
		t.Errorf("NilCache.Keys() returned error: %s, expected %s", err, hord.ErrNoDial)
	}
	nc.Close()
}
