// Package databases is the modular database interface for Hord. This interface is designed to work with any database
// backend. However, each implementation must conform to the defined Key/Value interface.
package databases

// Data is a structure that is returned for Gets and provided for Writes to the database
type Data struct {
	// Data is the actual data in a byte slice
	Data []byte
	// LastUpdated is a Epoch Nano timestamp that reflects the last time this data was updated
	LastUpdated int64
}

// Database is an interface that is used to create a unified database access object
type Database interface {
	Initialize() error
	Get(string) (*Data, error)
	Set(string, *Data) error
	Delete(string) error
	Keys() ([]string, error)
	HealthCheck() error
}
