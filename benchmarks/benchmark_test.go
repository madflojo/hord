package benchmarks

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/madflojo/hord"
	"github.com/madflojo/hord/drivers/bbolt"
	"github.com/madflojo/hord/drivers/cassandra"
	"github.com/madflojo/hord/drivers/hashmap"
	"github.com/madflojo/hord/drivers/nats"
	"github.com/madflojo/hord/drivers/redis"
)

func BenchmarkDrivers(b *testing.B) {
	// Create some test data for Benchmarks
	data := []byte(`
  {
    "userId": 1,
    "id": 1,
    "title": "sunt aut facere repellat provident occaecati excepturi optio reprehenderit",
    "body": "quia et suscipit\nsuscipit recusandae consequuntur expedita et cum\nreprehenderit molestiae ut ut quas totam\nnostrum rerum est autem sunt rem eveniet architecto"
  }
  `)

	// Create a Set of drivers to benchmark
	drivers := []string{"Redis", "Cassandra", "Hashmap", "BoltDB", "NATS", "Dragonfly", "KeyDB"}

	// Loop through the various DBs and TestData
	for _, driver := range drivers {
		b.Run("Bench_"+driver, func(b *testing.B) {
			var db hord.Database
			var err error
			switch driver {
			case "Redis", "Dragonfly", "KeyDB":
				server = "redis:6379"
				// Connect to Dragonfly
				if driver == "Dragonfly" {
					server = "dragonfly:6379"
				}
				// Connect to KeyDB
				if driver == "KeyDB" {
					server = "keydb:6379"
				}

				// Connect to Redis
				db, err = redis.Dial(redis.Config{
					ConnectTimeout: time.Duration(5) * time.Second,
					MaxActive:      500,
					MaxIdle:        100,
					IdleTimeout:    time.Duration(5) * time.Second,
					Server:         "redis:6379",
				})
				if err != nil {
					b.Fatalf("Got unexpected error when connecting to Redis - %s", err)
				}

			case "Cassandra":
				// Connect to Cassandra
				hosts := []string{"cassandra-primary", "cassandra"}
				db, err = cassandra.Dial(cassandra.Config{Hosts: hosts, Keyspace: "hord"})
				if err != nil {
					b.Fatalf("Got unexpected error when connecting to a cassandra cluster - %s", err)
				}

			case "NATS":
				db, err = nats.Dial(nats.Config{URL: "nats", Bucket: "hord"})
				if err != nil {
					b.Fatalf("Got unexpected error when connecting to nats - %s", err)
				}

			case "Hashmap":
				db, err = hashmap.Dial(hashmap.Config{})
				if err != nil {
					b.Fatalf("Got unexpected error when initializing hashmap - %s", err)
				}

			case "BoltDB":
				db, err = bbolt.Dial(bbolt.Config{
					Bucketname: "test",
					Filename:   "/tmp/bbolt-benchmark.db",
				})
				if err != nil {
					b.Fatalf("Got unexpected error when initializing hashmap - %s", err)
				}
				defer os.Remove("/tmp/bbolt-benchmark.db")

			default:
				b.Fatalf("Unknown DB Driver Specified")
			}
			defer db.Close()

			// Setup DB
			err = db.Setup()
			if err != nil {
				b.Fatalf("Unknown error setting up DB - %s", err)
			}

			// Execute HealthCheck
			err = db.HealthCheck()
			if err != nil {
				b.Fatalf("Error while checking health of DB - %s", err)
			}

			b.Run("SET", func(b *testing.B) {
				// Clean up Keys Created for Test
				b.Cleanup(func() {
					keys, _ := db.Keys()
					for _, d := range keys {
						_ = db.Delete(d)
					}
				})

				// Exec Benchmark
				for i := 0; i < b.N; i++ {
					err := db.Set("Test_Keys_"+fmt.Sprintf("%d", i), data)
					if err != nil {
						b.Fatalf("Error when executing Benchmark test - %s", err)
					}
				}
			})

			b.Run("GET", func(b *testing.B) {
				// Clean up Keys Created for Test
				b.Cleanup(func() {
					keys, _ := db.Keys()
					for _, d := range keys {
						_ = db.Delete(d)
					}
				})

				// Setup A Bunch of Keys
				b.StopTimer()
				for i := 0; i < 5000; i++ {
					_ = db.Set("Test_Keys_"+fmt.Sprintf("%d", i), data)
				}

				// Exec Benchmark
				count := 0
				b.StartTimer()
				for i := 0; i < b.N; i++ {
					if count > 4999 {
						count = 0
					}
					_, err := db.Get("Test_Keys_" + fmt.Sprintf("%d", count))
					if err != nil {
						b.Fatalf("Error when executing Benchmark test - %s", err)
					}
				}
			})

		})
	}
}
