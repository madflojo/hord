// Package app is the primary application package for Hord.
//
// This package handles all of the primary application responsibilities. These include request handling, logging, and
// basic runtime functionality.
package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/madflojo/hord/config"
	"github.com/madflojo/hord/databases"
	"github.com/madflojo/hord/databases/cassandra"
	"github.com/sirupsen/logrus"
)

// ErrShutdown is returned when a system shutdown was triggered under normal circumstances
var ErrShutdown = errors.New("System was shutdown")

// Config is a package wide global configuration object
var Config *config.Config

// db is a package global that is used to store databases
var db databases.Database

// log is a package global used for logging
var log *logrus.Logger

// This part is just for Errors
var (
	unableConnectDB = "Unable to connect to Database"
	unableInitDB    = "Unable initilize the Database"
)

// Run is the primary runnable function. Call this function from the command line packaging
func Run(cfg *config.Config) error {
	Config = cfg

	// Setup Logrus Logger instance
	log = logrus.New()

	if Config.Debug {
		log.Level = logrus.DebugLevel
		log.Debug("Enabling Debug logging mode")
	}

	if Config.Trace {
		log.Level = logrus.TraceLevel
		log.Trace("Enabling Trace logging mode")
	}

	// Dumping configuration for troubleshooting reasons
	log.Debugf("Dumping Config: %+v", Config)

	// Setup DB connection
	switch strings.ToLower(Config.DatabaseType) {
	case "cassandra":
		var err error
		db, err = cassandra.Dial(Config.Databases.Cassandra)
		if err != nil {
			return fmt.Errorf("%s %s- %s", unableConnectDB, "cassandra", err)
		}
	default:
		return fmt.Errorf("%s is not a known Database type", Config.DatabaseType)
	}

	// Initialize the database
	err := db.Initialize()
	if err != nil {
		return fmt.Errorf("%s - %s", unableInitDB, err)
	}

	// Start Health Checker
	go func() {
		for {
			err := db.HealthCheck()
			if err != nil {
				log.Errorf("Database healthcheck failed - %s", err)
			}
			go log.Trace("Databases healthceck success")
			time.Sleep(5 * time.Second)
		}
	}()

	// Start GRPC Listener
	log.WithFields(logrus.Fields{"listen": Config.Listen, "grpc_port": Config.GRPCPort}).Debugf("Starting GRPC listener")
	err = Listen()
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Errorf("Error returned from GRPC listener - %s", err)
		return err
	}
	return ErrShutdown
}
