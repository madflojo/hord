/*
Package cassandra provides a Hord database driver for Cassandra.

Cassandra is a highly scalable, distributed database designed to handle large amounts of data across many commodity servers. To use this driver, import it as follows:

	import (
	    "github.com/madflojo/hord"
	    "github.com/madflojo/hord/cassandra"
	)

# Connecting to the Database

Use the Dial() function to create a new client for interacting with Cassandra.

	var db hord.Database
	db, err := cassandra.Dial(cassandra.Config{})
	if err != nil {
	    // Handle connection error
	}

# Initialize database

Hord provides a Setup() function for preparing a database. This function is safe to execute after every Dial().

	err := db.Setup()
	if err != nil {
	    // Handle setup error
	}

# Database Operations

Hord provides a simple abstraction for working with Cassandra, with easy-to-use methods such as Get() and Set() to read and write values.

	// Connect to the Cassandra database
	db, err := cassandra.Dial(cassandra.Config{})
	if err != nil {
	    // Handle connection error
	}

	err := db.Setup()
	if err != nil {
	    // Handle setup error
	}

	// Set a value
	err = db.Set("key", []byte("value"))
	if err != nil {
	    // Handle error
	}

	// Retrieve a value
	value, err := db.Get("key")
	if err != nil {
	    // Handle error
	}
*/
package cassandra

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/madflojo/hord"
)

// Config is a generic configuration that is passed when Dialing the Cassandra cluster.
type Config struct {
	// Hosts is used to provide a list of Cassandra nodes. These hosts will be used to establish connectivity.
	Hosts []string

	// User provides the Username to authenticate with.
	User string

	// Password provides the User's password for authentication.
	Password string

	// Port specifies the listener port for Cassandra. This will be used to etablish connectivity.
	Port int

	// Keyspace defines the keyspace to use. The Keyspace will automatially be created when executing the Setup
	// function.
	Keyspace string

	// Consistency defines the desired consistency to use with Cassandra. By default Quorum will be used.
	Consistency string

	// EnableHostnameVerification will validate TLS Certificates with the hostname provided.
	EnableHostnameVerification bool

	// ReplicationStrategy is used to define the Cassandra replication strategy for the specified keyspace. Default
	// is SimpleStrategy.
	ReplicationStrategy string

	// Replicas is used to define the default number of replicas for data. Default is 1.
	Replicas int
}

// Database is used to interface with Cassandra. It also satisfies the Hord Database interface.
type Database struct {
	// conn is the underlying Cassandra connection
	conn *gocql.Session

	// config is a copy of the Config used during initialization
	config Config
}

// Dial will establish a session to a Cassandra cluster and provide a Database interface that can be used to interact
// with Cassandra.
func Dial(conf Config) (*Database, error) {
	db := Database{
		config: conf,
	}

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

// Setup will setup the basic keyspace and tables required for Hord to use Cassandra. If the database has
// already been initialized this function will not execute but return with a nil error. If any issues occur
// while initializing an error will be returned.
func (db *Database) Setup() error {
	if db == nil || db.conn == nil {
		return hord.ErrNoDial
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
	qry := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.hord ( key text, data blob, PRIMARY KEY (key));",
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

	if db == nil || db.conn == nil {
		return data, hord.ErrNoDial
	}

	if err := hord.ValidKey(key); err != nil {
		return data, err
	}

	err := db.conn.Query(`SELECT data FROM hord WHERE key = ?;`, key).Scan(&data)
	if err != nil && err != gocql.ErrNotFound {
		return data, err
	}
	if err == gocql.ErrNotFound {
		return data, hord.ErrNil
	}

	return data, nil
}

// Set is called when data within the database needs to be updated or inserted. This function will
// take the data provided and create an entry within the database using the key as a lookup value.
func (db *Database) Set(key string, data []byte) error {
	if db == nil || db.conn == nil {
		return hord.ErrNoDial
	}

	if err := hord.ValidKey(key); err != nil {
		return err
	}

	if err := hord.ValidData(data); err != nil {
		return err
	}

	err := db.conn.Query(`UPDATE hord SET data = ? WHERE key = ?`, data, key).Exec()
	return err
}

// Delete is called when data within the database needs to be deleted. This function will delete
// the data stored within the database for the specified key.
func (db *Database) Delete(key string) error {
	if db == nil || db.conn == nil {
		return hord.ErrNoDial
	}

	if err := hord.ValidKey(key); err != nil {
		return err
	}

	err := db.conn.Query(`DELETE FROM hord WHERE key = ?;`, key).Exec()
	if err != nil {
		return err
	}

	return nil
}

// Keys is called to retrieve a list of keys stored within the database. This function will query
// the Cassandra cluster returning all keys used within the hord database.
func (db *Database) Keys() ([]string, error) {
	var keys []string
	var key string

	if db == nil || db.conn == nil {
		return keys, hord.ErrNoDial
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
	if db == nil || db.conn == nil {
		return hord.ErrNoDial
	}
	err := db.conn.Query("SELECT now() FROM system.local;").Exec()
	if err != nil {
		return fmt.Errorf("health check of Cassandra cluster failed")
	}
	return nil
}

// Close will close the connection to Cassandra.
func (db *Database) Close() {
	if db != nil {
		db.conn.Close()
	}
}
