package cassandra

import (
	"testing"
)

func TestDialandSetup(t *testing.T) {
	hosts := []string{"cassandra", "cassandra-primary"}
	db, err := Dial(&Config{Hosts: hosts, Keyspace: "hord"})
	if err != nil {
		t.Errorf("Got unexpected error when connecting to a cassandra cluster - %s", err)
	}

	err = db.Initialize()
	if err != nil {
		t.Errorf("Got unexpected error when initializing cassandra cluster - %s", err)
	}

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
