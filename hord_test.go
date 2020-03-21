package hord

import (
	"github.com/madflojo/hord/drivers/cassandra"
	"testing"
)

func TestCassandraDriver(t *testing.T) {
	hosts := []string{"cassandra-primary", "cassandra"}
	var db Database
	db, err := cassandra.Dial(&cassandra.Config{Hosts: hosts, Keyspace: "hord"})
	if err != nil {
		t.Fatalf("Got unexpected error when connecting to a cassandra cluster - %s", err)
	}
	defer db.Close()
}
