package cassandra

import (
	"fmt"
	"testing"
	"time"
)

func TestErrNoDial(t *testing.T) {
	var db Database
	err := db.Setup()
	if err != ErrNoDial {
		t.Errorf("Expected no dialing error but got - %s", err)
	}

	err = db.HealthCheck()
	if err != ErrNoDial {
		t.Errorf("Expected no dialing error but got - %s", err)
	}

	err = db.Set("key", []byte("test"))
	if err != ErrNoDial {
		t.Errorf("Expected no dialing error but got - %s", err)
	}

	_, err = db.Get("key")
	if err != ErrNoDial {
		t.Errorf("Expected no dialing error but got - %s", err)
	}

	err = db.Delete("key")
	if err != ErrNoDial {
		t.Errorf("Expected no dialing error but got - %s", err)
	}

	_, err = db.Keys()
	if err != ErrNoDial {
		t.Errorf("Expected no dialing error but got - %s", err)
	}
}

func TestDialErrors(t *testing.T) {
	t.Run("No Hosts", func(t *testing.T) {
		_, err := Dial(&Config{})
		if err == nil {
			t.Errorf("Expected error when hosts are not specified, got nil")
		}
	})
}

func TestDialandSetup(t *testing.T) {
	hosts := []string{"cassandra-primary", "cassandra"}
	db, err := Dial(&Config{Hosts: hosts, Keyspace: "hord"})
	if err != nil {
		t.Fatalf("Got unexpected error when connecting to a cassandra cluster - %s", err)
	}
	time.Sleep(8 * time.Second)

	err = db.Setup()
	if err != nil {
		t.Fatalf("Got unexpected error when initializing cassandra cluster - %s", err)
	}

	// Let setup replicate across nodes
	time.Sleep(1 * time.Second)

	ksMeta, err := db.conn.KeyspaceMetadata(db.config.Keyspace)
	if err != nil {
		t.Fatalf("Got unexpected error when connecting to a cassandra cluster - %s", err)
	}

	if ksMeta.Name != db.config.Keyspace {
		t.Fatalf("Keyspace name from cluster does not match configured name got %s expected %s", ksMeta.Name, db.config.Keyspace)
	}

	if _, ok := ksMeta.Tables["hord"]; ok {
		return
	}
	t.Fatalf("Expected table hord to be created, did not find it within tables list - %v", ksMeta.Tables)
}

func TestDialKeyspaceNotCreated(t *testing.T) {
	hosts := []string{"cassandra", "cassandra-primary"}
	_, err := Dial(&Config{Hosts: hosts, Keyspace: "notcreated"})
	if err == nil {
		t.Fatalf("Unexpected nil when connecting to database with unsetup keyspace")
	}
}

func TestUsage(t *testing.T) {
	// Setup Environment
	hosts := []string{"cassandra", "cassandra-primary"}
	db, err := Dial(&Config{
		Hosts:               hosts,
		Keyspace:            "hord",
		Port:                7000,
		Consistency:         "Quorum",
		ReplicationStrategy: "SimpleStrategy",
	})
	if err != nil {
		t.Fatalf("Got unexpected error when connecting to a cassandra cluster - %s", err)
	}
	time.Sleep(8 * time.Second)

	err = db.Setup()
	if err != nil {
		t.Fatalf("Got unexpected error when initializing cassandra cluster - %s", err)
	}

	if db == nil {
		t.Fatalf("Database has not been configured , db = t %v", db)
	}

	t.Run("Writing data", func(t *testing.T) {
		data := []byte("Testing")
		err := db.Set("test_happypath", data)
		if err != nil {
			t.Fatalf("Unexpected error when writing data - %s", err)
		}

		var data2 []byte
		err = db.conn.Query(`SELECT data FROM hord WHERE key = ?;`, "test_happypath").Scan(&data2)
		if err != nil {
			t.Fatalf("Unable to find inserted record after write call, unexpected error - %s", err)
		}

		for k, v := range data2 {
			if v != data[k] {
				t.Errorf("Data mismatch expected %s got %s", data, data2)
			}
		}
	})

	t.Run("Writing Empty data", func(t *testing.T) {
		err := db.Set("test_emptydata", []byte(""))
		if err == nil {
			t.Errorf("Expected Error when writing with an empty byte slice, got nil")
		}
	})

	t.Run("Reading data", func(t *testing.T) {

		data, err := db.Get("test_happypath")
		if err != nil {
			t.Fatalf("Unexpected error when reading data - %s", err)
		}

		for i, v := range []byte("Testing") {
			if v != data[i] {
				t.Fatalf("Data mismatch from previously set data and data just read, got %+v expected %+v", data[i], v)
			}
		}
	})

	t.Run("Deleting data", func(t *testing.T) {
		err := db.Delete("test_happypath")
		if err != nil {
			t.Fatalf("Unexpected error when deleting data - %s", err)
		}

		var data []byte
		err = db.conn.Query(`SELECT data FROM hord WHERE key = ?;`, "test_happypath").Scan(&data)
		if err == nil {
			t.Fatalf("It does not appear data was completely deleted from table found - %+v", data)
		}
	})
}

func TestHealthCheck(t *testing.T) {
	hosts := []string{"cassandra-primary", "cassandra"}
	db, err := Dial(&Config{Hosts: hosts, Keyspace: "hord"})
	if err != nil {
		t.Fatalf("Got unexpected error when connecting to a cassandra cluster - %s", err)
	}
	time.Sleep(8 * time.Second)

	err = db.Setup()
	if err != nil {
		t.Fatalf("Got unexpected error when initializing cassandra cluster - %s", err)
	}
	time.Sleep(1 * time.Second)

	err = db.HealthCheck()
	if err != nil {
		t.Fatalf("Unexpected error when performing health check against cassandra cluster - %s", err)
	}
}

func TestKeys(t *testing.T) {
	hosts := []string{"cassandra-primary", "cassandra"}
	db, err := Dial(&Config{Hosts: hosts, Keyspace: "hord"})
	if err != nil {
		t.Fatalf("Got unexpected error when connecting to a cassandra cluster - %s", err)
	}
	time.Sleep(8 * time.Second)

	err = db.Setup()
	if err != nil {
		t.Fatalf("Got unexpected error when initializing cassandra cluster - %s", err)
	}
	time.Sleep(1 * time.Second)

	t.Run("No Keys", func(t *testing.T) {
		keys, err := db.Keys()
		if err != nil {
			t.Fatalf("Unexecpted error when obtaining a list of keys from the cassandra cluster - %s", err)
		}

		if len(keys) > 0 {
			t.Fatalf("Unexpected keys found in key list got - %+v", keys)
		}
	})

	t.Run("5 keys", func(t *testing.T) {
		// Setup
		data := []byte("Testing")
		for i := 0; i < 5; i++ {
			err := db.Set(fmt.Sprintf("Testing Keys %d", i), data)
			if err != nil {
				t.Fatalf("Error setting up test keys for testcase - %s", err)
			}
		}
		time.Sleep(5 * time.Second)

		keys, err := db.Keys()
		if err != nil {
			t.Fatalf("Unexecpted error when obtaining a list of keys from the cassandra cluster - %s", err)
		}

		if len(keys) != 5 {
			t.Fatalf("Unexpected number of keys found in key list got - %+v", keys)
		}

		// Tear down
		for i := 0; i < 5; i++ {
			_ = db.Delete(fmt.Sprintf("Testing Keys %d", i))
		}
	})
}
