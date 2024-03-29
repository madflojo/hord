package cache

import (
	"bytes"
	"errors"
	"testing"

	"github.com/madflojo/hord"
	"github.com/madflojo/hord/drivers/mock"
)

// Test Errors used for testing purposes
var (
	ErrDatabaseTest = errors.New("database error")
	ErrCacheTest    = errors.New("cache error")
)

// setupCache is a helper function to create a new Cache driver using the provided database and cache Config.
func setupCache(cacheConfig mock.Config, databaseConfig mock.Config) (*Cache, error) {
	database, err := mock.Dial(databaseConfig)
	if err != nil {
		return nil, err
	}

	cache, err := mock.Dial(cacheConfig)
	if err != nil {
		return nil, err
	}

	return Dial(Config{
		Database: database,
		Cache:    cache,
	})
}

func TestDial(t *testing.T) {
	unitTests := map[string]struct {
		config        Config
		expectedError error
	}{
		"No Config": {
			config:        Config{},
			expectedError: ErrInvalidDatabase,
		},
		"No Database": {
			config: Config{
				Cache: &mock.Database{},
			},
			expectedError: ErrInvalidDatabase,
		},
		"No Cache": {
			config: Config{
				Database: &mock.Database{},
			},
			expectedError: ErrInvalidCache,
		},
		"Happy Path": {
			config: Config{
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

func TestSetup(t *testing.T) {
	unitTests := map[string]struct {
		databaseError error
		cacheError    error
		expectedError error
	}{
		"Database Error": {
			databaseError: ErrDatabaseTest,
			cacheError:    nil,
			expectedError: ErrDatabaseTest,
		},
		"Cache Error": {
			databaseError: nil,
			cacheError:    ErrCacheTest,
			expectedError: ErrCacheTest,
		},
		"Both Errors": {
			databaseError: ErrDatabaseTest,
			cacheError:    ErrCacheTest,
			expectedError: ErrDatabaseTest,
		},
		"Happy Path": {
			databaseError: nil,
			cacheError:    nil,
			expectedError: nil,
		},
	}

	for name, test := range unitTests {
		t.Run(name, func(t *testing.T) {
			databaseConfig := mock.Config{
				SetupFunc: func() error {
					return test.databaseError
				},
			}
			cacheConfig := mock.Config{
				SetupFunc: func() error {
					return test.cacheError
				},
			}

			db, err := setupCache(cacheConfig, databaseConfig)
			if err != nil {
				t.Fatalf("Failed to connect to database - %s", err)
			}

			err = db.Setup()
			if !errors.Is(err, test.expectedError) {
				t.Errorf("Setup() returned error: %s, expected %s", err, test.expectedError)
			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	unitTests := map[string]struct {
		databaseError error
		cacheError    error
		expectedError error
	}{
		"Database Error": {
			databaseError: ErrDatabaseTest,
			cacheError:    nil,
			expectedError: ErrDatabaseTest,
		},
		"Cache Error": {
			databaseError: nil,
			cacheError:    ErrCacheTest,
			expectedError: ErrCacheTest,
		},
		"Both Errors": {
			databaseError: ErrDatabaseTest,
			cacheError:    ErrCacheTest,
			expectedError: ErrDatabaseTest,
		},
		"Happy Path": {
			databaseError: nil,
			cacheError:    nil,
			expectedError: nil,
		},
	}

	for name, test := range unitTests {
		t.Run(name, func(t *testing.T) {
			databaseConfig := mock.Config{
				HealthCheckFunc: func() error {
					return test.databaseError
				},
			}
			cacheConfig := mock.Config{
				HealthCheckFunc: func() error {
					return test.cacheError
				},
			}

			db, err := setupCache(cacheConfig, databaseConfig)
			if err != nil {
				t.Fatalf("Failed to connect to database - %s", err)
			}

			err = db.HealthCheck()
			if !errors.Is(err, test.expectedError) {
				t.Errorf("HealthCheck() returned error: %s, expected %s", err, test.expectedError)
			}
		})
	}
}

func TestGet(t *testing.T) {
	cacheConfig := mock.Config{
		GetFunc: func(key string) ([]byte, error) {
			switch key {
			case "cache-hit":
				return []byte("cache-data"), nil
			case "cache-miss":
				return nil, hord.ErrNil
			case "cache-error":
				return nil, ErrCacheTest
			case "cache-write-error":
				return nil, hord.ErrNil
			case "database-error":
				return nil, hord.ErrNil
			}
			return nil, errors.New("Unexpected Cache Error")
		},
		SetFunc: func(key string, _ []byte) error {
			if key == "cache-write-error" {
				return ErrCacheTest
			}
			return nil
		},
	}
	databaseConfig := mock.Config{
		GetFunc: func(key string) ([]byte, error) {
			switch key {
			case "cache-hit":
				return nil, errors.New("Expected Cache Hit")
			case "cache-miss":
				return []byte("database-data"), nil
			case "cache-error":
				return nil, errors.New("Expected Cache Error")
			case "cache-write-error":
				return nil, nil
			case "database-error":
				return nil, ErrDatabaseTest
			}
			return nil, errors.New("Unexpected Database Error")
		},
	}

	unitTests := map[string]struct {
		key           string
		expectedError error
		expectedData  []byte
	}{
		"Cache Hit": {
			key:           "cache-hit",
			expectedError: nil,
			expectedData:  []byte("cache-data"),
		},
		"Cache Miss": {
			key:           "cache-miss",
			expectedError: nil,
			expectedData:  []byte("database-data"),
		},
		"Cache Error": {
			key:           "cache-error",
			expectedError: ErrCacheTest,
			expectedData:  nil,
		},
		"Cache Write Error": {
			key:           "cache-write-error",
			expectedError: ErrCacheTest,
			expectedData:  nil,
		},
		"Database Error": {
			key:           "database-error",
			expectedError: ErrDatabaseTest,
			expectedData:  nil,
		},
	}

	for name, test := range unitTests {
		t.Run(name, func(t *testing.T) {
			db, err := setupCache(cacheConfig, databaseConfig)
			if err != nil {
				t.Fatalf("Failed to connect to database - %s", err)
			}

			data, err := db.Get(test.key)
			if !errors.Is(err, test.expectedError) {
				t.Errorf("Get(%s) returned error: %s, expected %s", test.key, err, test.expectedError)
			}
			if string(data) != string(test.expectedData) {
				t.Errorf("Get(%s) returned data: %s, expected %s", test.key, data, test.expectedData)
			}
		})
	}
}

func TestSet(t *testing.T) {
	cacheValue := []byte("")
	databaseValue := []byte("")

	cacheConfig := mock.Config{
		SetFunc: func(key string, data []byte) error {
			cacheValue = nil
			switch key {
			case "happy-path":
				cacheValue = data
				return nil
			case "cache-error":
				return ErrCacheTest
			case "database-error":
				cacheValue = data
				return nil
			}
			return errors.New("Unexpected Cache Error")
		},
	}
	databaseConfig := mock.Config{
		SetFunc: func(key string, data []byte) error {
			databaseValue = nil
			switch key {
			case "happy-path":
				databaseValue = data
				return nil
			case "cache-error":
				databaseValue = data
				return nil
			case "database-error":
				return ErrDatabaseTest
			}
			return errors.New("Unexpected Database Error")
		},
	}

	unitTests := map[string]struct {
		key                   string
		data                  []byte
		expectedError         error
		expectedCacheValue    []byte
		expectedDatabaseValue []byte
	}{
		"Cache Error": {
			key:                   "cache-error",
			data:                  []byte("cache-error-data"),
			expectedError:         ErrCacheTest,
			expectedCacheValue:    nil,
			expectedDatabaseValue: []byte("cache-error-data"),
		},
		"Database Error": {
			key:                   "database-error",
			data:                  []byte("database-error-data"),
			expectedError:         ErrDatabaseTest,
			expectedCacheValue:    nil,
			expectedDatabaseValue: nil,
		},
		"Happy Path": {
			key:                   "happy-path",
			data:                  []byte("happy-path-data"),
			expectedError:         nil,
			expectedCacheValue:    []byte("happy-path-data"),
			expectedDatabaseValue: []byte("happy-path-data"),
		},
	}

	for name, test := range unitTests {
		t.Run(name, func(t *testing.T) {
			db, err := setupCache(cacheConfig, databaseConfig)
			if err != nil {
				t.Fatalf("Failed to connect to database - %s", err)
			}

			err = db.Set(test.key, test.data)
			if !errors.Is(err, test.expectedError) {
				t.Errorf("Get(%s) returned error: %s, expected %s", test.key, err, test.expectedError)
			}
			if !bytes.Equal(cacheValue, test.expectedCacheValue) {
				t.Errorf("Get(%s) returned data: %s, expected %s", test.key, cacheValue, test.expectedCacheValue)
			}
			if !bytes.Equal(databaseValue, test.expectedDatabaseValue) {
				t.Errorf("Get(%s) returned data: %s, expected %s", test.key, databaseValue, test.expectedDatabaseValue)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	cacheConfig := mock.Config{
		DeleteFunc: func(key string) error {
			switch key {
			case "happy-path":
				return nil
			case "cache-error":
				return ErrCacheTest
			case "database-error":
				return nil
			}
			return errors.New("Unexpected Cache Error")
		},
	}
	databaseConfig := mock.Config{
		DeleteFunc: func(key string) error {
			switch key {
			case "happy-path":
				return nil
			case "cache-error":
				return nil
			case "database-error":
				return ErrDatabaseTest
			}
			return errors.New("Unexpected Database Error")
		},
	}

	unitTests := map[string]struct {
		key           string
		expectedError error
	}{
		"Cache Error": {
			key:           "cache-error",
			expectedError: ErrCacheTest,
		},
		"Database Error": {
			key:           "database-error",
			expectedError: ErrDatabaseTest,
		},
		"Happy Path": {
			key:           "happy-path",
			expectedError: nil,
		},
	}

	for name, test := range unitTests {
		t.Run(name, func(t *testing.T) {
			db, err := setupCache(cacheConfig, databaseConfig)
			if err != nil {
				t.Fatalf("Failed to connect to database - %s", err)
			}

			err = db.Delete(test.key)
			if !errors.Is(err, test.expectedError) {
				t.Errorf("Get(%s) returned error: %s, expected %s", test.key, err, test.expectedError)
			}
		})
	}
}

func TestKeys(t *testing.T) {
	databaseConfig := mock.Config{
		KeysFunc: func() ([]string, error) {
			return []string{"database-key"}, nil
		},
	}
	cacheConfig := mock.Config{
		KeysFunc: func() ([]string, error) {
			return []string{"cache-key"}, nil
		},
	}

	db, err := setupCache(cacheConfig, databaseConfig)
	if err != nil {
		t.Fatalf("Failed to connect to database - %s", err)
	}

	keys, err := db.Keys()
	if err != nil {
		t.Errorf("Keys() returned error: %s", err)
	}
	if keys[0] != "database-key" {
		t.Errorf("Keys() returned data: %v, expected %v", keys, []string{"database-key"})
	}
}

func TestClose(t *testing.T) {
	databaseConfig := mock.Config{}
	cacheConfig := mock.Config{}

	db, err := setupCache(cacheConfig, databaseConfig)
	if err != nil {
		t.Fatalf("Failed to connect to database - %s", err)
	}

	db.Close()
}
