// Package hord provides a modular interface for interacting with key-value datastores. This interface is designed
// to work with any database. This means some drivers may implement more functionality than the generic Hord
// interface provides.
//
// The goal of this package is to allow users to quickly switch out the underlying database without having to re-write
// significant application code.
package hord

// Database is an interface that is used to create a unified database access object
type Database interface {
	Setup() error
	HealthCheck() error
	Get(string) ([]byte, error)
	Set(string, []byte) error
	Delete(string) error
	Keys() ([]string, error)
}
