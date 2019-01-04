package cassandra

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/madflojo/hord/databases"
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

// Dial will establish a session to a Cassandra cluster and provide a Database interface that can be used to interact
// with Cassandra.
func Dial(conf *Config) (*Database, error) {
	var db Database

	// Inject the Database interface with provided configuration
	db.config = conf

	// Setup cluster hosts
	if len(db.config.Hosts) < 1 {
		return nil, fmt.Errorf("Must provide at least one Cassandra host to connect to")
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
			return nil, fmt.Errorf("If ReplicationStrategy is set, Replicas is a required value")
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

// Initialize will setup the basic keyspace and tables required for a Hord database. If the database has
// already been initialized this function will not execute but return with a nil error. If any issues occur
// while initializing an error will be returned.
func (db *Database) Initialize() error {
	ksMeta, err := db.conn.KeyspaceMetadata(db.config.Keyspace)

	// If keyspace exists and there was an error dip out with an err
	if err != nil && err != gocql.ErrNoKeyspace {
		return fmt.Errorf("Unable to initialize database, failed keystore validation - %s", err)
	}

	// If keyspace doesn't exist, let's get creating
	if err == gocql.ErrNoKeyspace {
		qry := fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': '%s', 'replication_factor' : %d};",
			db.config.Keyspace,
			db.config.ReplicationStrategy,
			db.config.Replicas)
		err := db.conn.Query(qry).Exec()
		if err != nil {
			return fmt.Errorf("Unable to initialize database, failed to create keystore - %s", err)
		}
	}

	// Check if table already exists, if not create it
	if _, ok := ksMeta.Tables["hord"]; ok {
		return nil
	}
	qry := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.hord ( key text, data blob, last_updated bigint, PRIMARY KEY (key));",
		db.config.Keyspace)
	err = db.conn.Query(qry).Exec()
	if err != nil {
		return fmt.Errorf("Unable to initialize database, failed to create table - %s", err)
	}

	return nil
}

// Read is called to retrieve data from the database. This function will take in a key and return
// the database.Data structure stored within the database.
func (db *Database) Read(key string) (*databases.Data, error) {
	var data databases.Data

	err := db.conn.Query(`SELECT data, last_updated FROM hord WHERE key = ?;`, key).Scan(&data.Data, &data.LastUpdated)
	if err != nil {
		return &databases.Data{}, err
	}

	return &data, nil
}

// Set is called when data within the database needs to be updated or inserted. This function will
// take the data provided and create an entry within the database using the key as a lookup value.
func (db *Database) Set(key string, data *databases.Data) error {
	err := db.conn.Query(`UPDATE hord SET data = ?, last_updated = ? WHERE key = ?`, data.Data, data.LastUpdated, key).Exec()
	return err
}

func (db *Database) Delete(key string) error {
	return nil
}

func (db *Database) HealthCheck() error {
	return nil
}
