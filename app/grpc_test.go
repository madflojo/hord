package app

import (
	"context"
	"fmt"
	"github.com/madflojo/hord/config"
	"github.com/madflojo/hord/databases"
	pb "github.com/madflojo/hord/proto/client"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"testing"
	"time"
)

// mockDB is a mock Database used to satisfy the db global
type mockDB struct{}

func (db *mockDB) Initialize() error {
	return nil
}

func (db *mockDB) Get(key string) (*databases.Data, error) {
	d := &databases.Data{
		Data:        []byte("Hello"),
		LastUpdated: 1970,
	}
	if key == "fail" {
		return d, fmt.Errorf("Unknown key %s", key)
	}
	return d, nil
}

func (db *mockDB) Set(key string, data *databases.Data) error {
	if key == "fail" {
		return fmt.Errorf("Bad key yo")
	}
	return nil
}

func (db *mockDB) Delete(key string) error {
	if key == "fail" {
		return fmt.Errorf("Bad key yo")
	}
	return nil
}

func (db *mockDB) Keys() ([]string, error) {
	return []string{""}, nil
}

func (db *mockDB) HealthCheck() error {
	return nil
}

func TestGRPC(t *testing.T) {
	// Setup base config
	Config = &config.Config{}
	Config.Listen = "0.0.0.0"
	Config.GRPCPort = "9000"

	// Create a DB mock
	db = &mockDB{}
	err := db.HealthCheck()
	if err != nil {
		t.Errorf("Could not setup Database Mock properly for test execution - %s", err)
	}

	// Create a logger
	log = logrus.New()

	// Start listener in background
	go func() {
		err := Listen()
		if err != nil {
			t.Logf("Failed to start GRPC Listener - %s", err)
			t.FailNow()
		}
	}()

	time.Sleep(20 * time.Millisecond)

	// Connect to newly started listener
	c, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
	if err != nil {
		t.Logf("Failed to connect to GRPC Listener - %s", err)
		t.FailNow()
	}
	defer c.Close()

	// Create grpc client
	client := pb.NewHordClient(c)

	t.Run("Set data to DB", func(t *testing.T) {
		msg := &pb.SetRequest{
			Key:  "testing",
			Data: []byte("hello"),
		}

		m, err := client.Set(context.Background(), msg)
		if err != nil {
			t.Errorf("Unexpected failure when calling Set - %s", err)
		}

		if m.Status.Code > 0 {
			t.Errorf("Unexpected error returned from Set call - %s", m.Status.Description)
		}

		if msg.Key != m.Key {
			t.Errorf("Set function should echo key, it returned different results expected %s got %s", msg.Key, m.Key)
		}
	})

	t.Run("Set data without a key", func(t *testing.T) {
		msg := &pb.SetRequest{
			Data: []byte("hello"),
		}

		m, err := client.Set(context.Background(), msg)
		if err != nil {
			t.Errorf("Unexpected failure when calling Set - %s", err)
		}

		if m.Status.Code != 4 {
			t.Errorf("Unexpected error returned from Set call - %s", m.Status.Description)
		}
	})

	t.Run("Set data with failed DB call", func(t *testing.T) {
		msg := &pb.SetRequest{
			Key:  "fail",
			Data: []byte("hello"),
		}

		m, err := client.Set(context.Background(), msg)
		if err != nil {
			t.Errorf("Unexpected failure when calling Set - %s", err)
		}

		if m.Status.Code != 5 {
			t.Errorf("Unexpected error returned from Set call - %s", m.Status.Description)
		}
	})

	t.Run("Get data from DB", func(t *testing.T) {
		msg := &pb.GetRequest{
			Key: "testing",
		}

		m, err := client.Get(context.Background(), msg)
		if err != nil {
			t.Errorf("Unexepcted failure when calling Get - %s", err)
		}

		if m.Status.Code > 0 {
			t.Errorf("Unexpected error returned from Get call - %s", m.Status.Description)
		}

		if msg.Key != m.Key {
			t.Errorf("Get function should echo key, it returned different results. Expected %s got %s", msg.Key, m.Key)
		}

		if m.LastUpdated != 1970 {
			t.Errorf("Get function returned an unexpected last updated time got %d", m.LastUpdated)
		}
	})

	t.Run("Get data with no key", func(t *testing.T) {
		msg := &pb.GetRequest{}

		m, err := client.Get(context.Background(), msg)
		if err != nil {
			t.Errorf("Unexepcted failure when calling Get - %s", err)
		}

		if m.Status.Code != 4 {
			t.Errorf("Expected error return from Get call got success - %s", m.Status.Description)
		}
	})

	t.Run("Get data with failed DB call", func(t *testing.T) {
		msg := &pb.GetRequest{
			Key: "fail",
		}

		m, err := client.Get(context.Background(), msg)
		if err != nil {
			t.Errorf("Unexepcted failure when calling Get - %s", err)
		}

		if m.Status.Code != 5 {
			t.Errorf("Expected error code 5 from Get call got %d - %s", m.Status.Code, m.Status.Description)
		}
	})

	t.Run("Delete data from DB", func(t *testing.T) {
		msg := &pb.DeleteRequest{
			Key: "testing",
		}

		m, err := client.Delete(context.Background(), msg)
		if err != nil {
			t.Errorf("Unexpected failure when calling Delete - %s", err)
		}

		if m.Status.Code > 0 {
			t.Errorf("Unexpected error returned from Delete call - %s", m.Status.Description)
		}

		if msg.Key != m.Key {
			t.Errorf("Delete function should echo key, it returned different results. Expected %s got %s", msg.Key, m.Key)
		}
	})

	t.Run("Delete data from DB without a key", func(t *testing.T) {
		msg := &pb.DeleteRequest{}

		m, err := client.Delete(context.Background(), msg)
		if err != nil {
			t.Errorf("Unexpected failure when calling Delete - %s", err)
		}

		if m.Status.Code > 4 {
			t.Errorf("Unexpected error returned from Delete call - %s", m.Status.Description)
		}
	})

	t.Run("Delete data with a bad DB call", func(t *testing.T) {
		msg := &pb.DeleteRequest{
			Key: "fail",
		}

		m, err := client.Delete(context.Background(), msg)
		if err != nil {
			t.Errorf("Unexpected failure when calling Delete - %s", err)
		}

		if m.Status.Code > 5 {
			t.Errorf("Unexpected error returned from Delete call - %s", m.Status.Description)
		}
	})

}
