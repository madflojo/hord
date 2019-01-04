package cassandra

import (
	"github.com/madflojo/hord/databases"
	"testing"
	"time"
)

func TestDialandSetup(t *testing.T) {
	hosts := []string{"cassandra-primary", "cassandra"}
	db, err := Dial(&Config{Hosts: hosts, Keyspace: "hord"})
	if err != nil {
		t.Errorf("Got unexpected error when connecting to a cassandra cluster - %s", err)
	}
	time.Sleep(30 * time.Second)

	err = db.Initialize()
	if err != nil {
		t.Errorf("Got unexpected error when initializing cassandra cluster - %s", err)
	}
	time.Sleep(10 * time.Second)

	ksMeta, err := db.conn.KeyspaceMetadata(db.config.Keyspace)
	if err != nil {
		t.Errorf("Got unexpected error when connecting to a cassandra cluster - %s", err)
	}

	if ksMeta.Name != db.config.Keyspace {
		t.Errorf("Keyspace name from cluster does not match configured name got %s expected %s", ksMeta.Name, db.config.Keyspace)
	}

	if _, ok := ksMeta.Tables["hord"]; ok {
		return
	}
	t.Errorf("Expected table hord to be created, did not find it within tables list - %v", ksMeta.Tables)
}

func TestDialKeyspaceNotCreated(t *testing.T) {
	hosts := []string{"cassandra", "cassandra-primary"}
	_, err := Dial(&Config{Hosts: hosts, Keyspace: "notcreated"})
	if err == nil {
		t.Errorf("Unexpected nil when connecting to database with not created keyspace")
	}
}

func TestHappyPath(t *testing.T) {
	// Setup Environment
	hosts := []string{"cassandra", "cassandra-primary"}
	db, err := Dial(&Config{Hosts: hosts, Keyspace: "hord"})
	if err != nil {
		t.Errorf("Got unexpected error when connecting to a cassandra cluster - %s", err)
	}
	time.Sleep(30 * time.Second)

	err = db.Initialize()
	if err != nil {
		t.Errorf("Got unexpected error when initializing cassandra cluster - %s", err)
	}

	t.Run("Writing data", func(t *testing.T) {
		data := &databases.Data{}
		data.Data = []byte("Testing")
		now := time.Now()
		data.LastUpdated = now.UnixNano()

		err := db.Set("test_happypath", data)
		if err != nil {
			t.Errorf("Unexpected error when writing data - %s", err)
		}

		err = db.conn.Query(`SELECT data, last_updated FROM hord WHERE key = ?;`, "test_happypath").Scan(&data.Data, &data.LastUpdated)
		if err != nil {
			t.Errorf("Unable to find inserted record after write call, unexpected error - %s", err)
		}
	})

	t.Run("Reading data", func(t *testing.T) {
		data, err := db.Read("test_happypath")
		if err != nil {
			t.Errorf("Unexpected error when reading data - %s", err)
		}

		for i, v := range []byte("Testing") {
			if v != data.Data[i] {
				t.Errorf("Data mismatch from previously set data and data just read, got %+v expected %+v", data.Data[i], v)
			}
		}
	})
}
