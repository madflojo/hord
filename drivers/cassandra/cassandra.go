// Package cassandra is a Hord database driver for Cassandra. This package implements the Hord interface
// and can be used to interact with Cassandra database clusters.
//
//  // Connect to a Cassandra Cluster
//  db, err := cassandra.Dial(&cassandra.Config{})
//  if err != nil {
//    // do stuff
//  }
//
//  // Setup and Initialize the Keyspace if necessary
//  err = db.Setup()
//  if err != nil {
//    // do stuff
//  }
//
//  // Write data to the cluster
//  err = db.Set("mykey", []byte("My Data"))
//  if err != nil {
//    // do stuff
//  }
//
//  // Fetch the same data
//  d, err := db.Get("mykey")
//  if err != nil {
//    // do stuff
//  }
//
package cassandra

import (
	"fmt"
	"github.com/gocql/gocql"
)

// Config is a generic configuration type that users can use to pass in configuration when Dialing the Cassandra
// database.
type Config struct {
	Hosts                      []string
	User                       string
	Password                   string
	Port                       int
	Keyspace                   string
	Consistency                string
	EnableHostnameVerification bool
	ReplicationStrategy        string
	Replicas                   int
}

// Database is a interface struct that enables database functionality and stores configuration.
type Database struct {
	conn   *gocql.Session
	config *Config
}

// this is for Errors and Loggings
var (
	dbNill = "database has not been configured"
)

// Dial will establish a session to a Cassandra cluster and provide a Database interface that can be used to interact
// with Cassandra.
func Dial(conf *Config) (*Database, error) {
	var db Database

	// Inject the Database interface with provided configuration
	db.config = conf

	// Setup cluster hosts
	if len(db.config.Hosts) < 1 {
		return nil, fmt.Errorf("must provide at least one Cassandra host to connect to")
	}
	// For Cassandra, only one host needs to be specified, the client will identify peers from the cluster
	cluster := gocql.NewCluster(db.config.Hosts[0])
	cluster.ProtoVersion = 4

	// Define default consistency
	switch db.config.Consistency {
	case "Quorum":
		cluster.Consistency = gocql.Quorum
	default:
		cluster.Consistency = gocql.Quorum
	}

	// Define port if non-default
	if db.config.Port > 0 {
		cluster.Port = db.config.Port
	}

	// Define keyspace if provided
	cluster.Keyspace = "example"
	if db.config.Keyspace != "" {
		cluster.Keyspace = db.config.Keyspace
	}

	// Define replication strategy
	// TODO: Add network topology replication strategy settings
	switch db.config.ReplicationStrategy {
	case "SimpleStrategy":
		if db.config.Replicas < 1 {
			return nil, fmt.Errorf("if ReplicationStrategy is set, Replicas is a required value")
		}
	default:
		db.config.ReplicationStrategy = "SimpleStrategy"
		db.config.Replicas = 1
	}

	// Setup new session
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	db.conn = session

	return &db, nil
}

// Setup will setup the basic keyspace and tables required for a Hord database. If the database has
// already been initialized this function will not execute but return with a nil error. If any issues occur
// while initializing an error will be returned.
func (db *Database) Setup() error {
	//
	if db == nil {
		return fmt.Errorf("%s , db = %v", dbNill, db)
	}
	ksMeta, err := db.conn.KeyspaceMetadata(db.config.Keyspace)

	// If keyspace exists and there was an error dip out with an err
	if err != nil && err != gocql.ErrNoKeyspace {
		return fmt.Errorf("unable to initialize database, failed keystore validation - %s", err)
	}

	// If keyspace doesn't exist, let's get creating
	if err == gocql.ErrNoKeyspace {
		qry := fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': '%s', 'replication_factor' : %d};",
			db.config.Keyspace,
			db.config.ReplicationStrategy,
			db.config.Replicas)
		err := db.conn.Query(qry).Exec()
		if err != nil {
			return fmt.Errorf("unable to initialize database, failed to create keystore - %s", err)
		}
	}

	// Check if table already exists, if not create it
	if _, ok := ksMeta.Tables["hord"]; ok {
		return nil
	}
	qry := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.hord ( key text, data blob PRIMARY KEY (key));",
		db.config.Keyspace)
	err = db.conn.Query(qry).Exec()
	if err != nil {
		return fmt.Errorf("unable to initialize database, failed to create table - %s", err)
	}

	return nil
}

// Get is called to retrieve data from the database. This function will take in a key and return
// the data or any errors received from querying the database.
func (db *Database) Get(key string) ([]byte, error) {
	var data []byte

	if db == nil {
		return data, fmt.Errorf("%s , db = %v", dbNill, db)
	}

	err := db.conn.Query(`SELECT data FROM hord WHERE key = ?;`, key).Scan(&data)
	if err != nil {
		return data, err
	}

	return data, nil
}

// Set is called when data within the database needs to be updated or inserted. This function will
// take the data provided and create an entry within the database using the key as a lookup value.
func (db *Database) Set(key string, data []byte) error {
	if db == nil {
		return fmt.Errorf("%s , db = %v", dbNill, db)
	}
  if len(data) == 0 {
    return fmt.Errorf("data cannot be empty")
  }
	err := db.conn.Query(`UPDATE hord SET data = ? WHERE key = ?`, data, key).Exec()
	return err
}

// Delete is called when data within the database needs to be deleted. This function will delete
// the data stored within the database for the specified Primary Key.
func (db *Database) Delete(key string) error {
	if db == nil {
		return fmt.Errorf("%s , db = %v", dbNill, db)
	}
	err := db.conn.Query(`DELETE FROM hord WHERE key = ?;`, key).Exec()
	if err != nil {
		return err
	}
	return nil
}

// Keys is called to retrieve a list of keys stored within the database. This function will query
// the Cassandra cluster returning all Primary Keys used within the hord table.
func (db *Database) Keys() ([]string, error) {
	var keys []string
	var key string

	if db == nil {
		return keys, fmt.Errorf("%s , db = %v", dbNill, db)
	}

	l := db.conn.Query("SELECT key from hord;").Iter()
	for l.Scan(&key) {
		keys = append(keys, key)
	}

	err := l.Close()
	if err != nil {
		return keys, err
	}

	return keys, nil
}

// HealthCheck is used to verify connectivity and health of the Cassandra cluster. This function
// simply runs a generic query against Cassandra. If the query errors in any fashion this function
// will also return an error.
func (db *Database) HealthCheck() error {
	if db == nil {
		return fmt.Errorf("%s , db = %v", dbNill, db)
	}
	err := db.conn.Query("SELECT now() FROM system.local;").Exec()
	if err != nil {
		return fmt.Errorf("health check of Cassandra cluster failed")
	}
	return nil
}
