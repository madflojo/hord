// Package config is used to store the hord configuration struct as well as functions for reading configuration files, environment variables, external sources.
package config

import (
	"github.com/madflojo/hord/databases/cassandra"
)

// Config is the configuration struct which will control the application.
type Config struct {
	// Debug is used to determine if Debug logging should be enabled or not.
	Debug bool

	// Trace is used to determine if Trace logging should be enabled or not.
	Trace bool

	// Peers is a list of Peers identified from cli/configuration file. This list is used to seed the Memberlist which will discover new peers via the SWIM protocol.
	Peers []string

	// Listen is the address to bind to listen for GRPC and HTTP requests
	Listen string

	// GRPC Port is the port used to listen for GRPC requests
	GRPCPort string

	// DatabaseType is used to determine the database type to use for the backend data source.
	DatabaseType string

	// Databases is a struct that contains configuration for various databases that can be used.
	Databases Databases
}

// Databases is a type that aggregates all of the various supported DB configurations this type isn't used directly but as an import within the Config type.
type Databases struct {
	// Cassandra is used to store configuration that is unique to a Cassandra database type. This is only used if Cassandra is the selected database type.
	Cassandra *cassandra.Config
}
