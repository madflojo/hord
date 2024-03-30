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
				CacheType: "invalide",
				Database:  &mock.Database{},
				Cache:     &mock.Database{},
			},
			expectedError: ErrNoType,
		},
		"Type: Lookaside": {
			config: Config{
				CacheType: Lookaside,
				Database:  &mock.Database{},
				Cache:     &mock.Database{},
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
