// Package app is the primary application package for Hord.
//
// This package handles all of the primary application responsibilities. These include request handling, logging, and
// basic runtime functionality.
package app

import (
	"errors"
	"fmt"
	"github.com/madflojo/hord/config"
	"github.com/madflojo/hord/databases"
	"github.com/madflojo/hord/databases/cassandra"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

var ErrShutdown = errors.New("System was shutdown")

var Config *config.Config
var db databases.Database
var log *logrus.Logger

func Run(cfg *config.Config) error {
	Config = cfg

	// Setup Logrus Logger instance
	log = logrus.New()

	if Config.Debug {
		log.Level = logrus.DebugLevel
		log.Debug("Enabling Debug logging mode")
	}

	// Dumping configuration for troubleshooting reasons
	log.Debugf("Dumping Config: %+v", Config)

	// Setup DB connection
	switch strings.ToLower(Config.DatabaseType) {
	case "cassandra":
		d, err := cassandra.Dial(Config.Databases.Cassandra)
		if err != nil {
			return fmt.Errorf("Unable to connect to cassandra database - %s", err)
		}

		// Assign d to db global
		db = d

		// Start Health Checker
		go func() {
			for {
				err = db.HealthCheck()
				if err != nil {
					log.Errorf("Database healthcheck failed - %s", err)
				}
				if Config.Debug {
					log.Debug("Databases healthceck success")
				}
				time.Sleep(5 * time.Second)
			}
		}()
	default:
		return fmt.Errorf("%s is not a known Database type", Config.DatabaseType)
	}

	time.Sleep(1 * time.Minute)
	return ErrShutdown
}
