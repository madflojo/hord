package databases

// Data is a structure that is returned for Reads and provided for Writes to the database
type Data struct {
	// Data is the actual data in a byte slice
	Data []byte
	// LastUpdated is a Epoch Nano timestamp that reflects the last time this data was updated
	LastUpdated int64
}

// Database is an interface that is used to create a unified database access object
type Database interface {
	Read(string) (*Data, error)
	Set(string, *Data) error
	Delete(string) error
	HealthCheck() error
}
