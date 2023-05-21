# Hord

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/madflojo/hord)
[![codecov](https://codecov.io/gh/madflojo/hord/branch/main/graph/badge.svg?token=0TTTEWHLVN)](https://codecov.io/gh/madflojo/hord)
[![Go Report Card](https://goreportcard.com/badge/github.com/madflojo/hord)](https://goreportcard.com/report/github.com/madflojo/hord)
[![Documentation](https://godoc.org/github.com/madflojo/hord?status.svg)](http://godoc.org/github.com/madflojo/hord)

Hord is a user-friendly and reliable interface for Go that enables storing and retrieving data from various key-value databases. It offers a straightforward approach to interacting with database backends, prioritizing essential functions like `Get`, `Set`, `Delete`, and `Keys`. Hord also supports multiple storage backends through a suite of drivers, allowing you to choose the one that best suits your needs. 

Additionally, to facilitate testing, Hord includes a mock driver package that enables users to define custom functions and simulate interactions with a Hord driver, making it easier to write unit tests and validate functionality.

## Database Drivers:

| Database | Support | Comments | Protocol Compatible Alternatives |
| -------- | ------- | -------- | -------------------------------- |
| [BoltDB](https://github.com/etcd-io/bbolt) | ✅ | | |
| [Cassandra](https://cassandra.apache.org/) | ✅ | | [ScyllaDB](https://www.scylladb.com/), [YugabyteDB](https://www.yugabyte.com/), [Azure Cosmos DB](https://learn.microsoft.com/en-us/azure/cosmos-db/introduction) |
| [Couchbase](https://www.couchbase.com/) | Pending |||
| Hashmap | ✅ |||
| Mock | ✅ | Mock Database interactions within unit tests ||
| [NATS](https://nats.io/) | ✅ | Experimental ||
| [Redis](https://redis.io/) | ✅ || [Dragonfly](https://www.dragonflydb.io/), [KeyDB](https://docs.keydb.dev/) |

## Usage

The below example shows using Hord to connect and interact with Cassandra.

```go
import "github.com/madflojo/hord"
import "github.com/madflojo/hord/driver/cassandra"

func main() {
  // Define our DB Interface
  var db hord.Database

  // Connect to a Cassandra Cluster
  db, err := cassandra.Dial(&cassandra.Config{})
  if err != nil {
    // do stuff
  }

  // Setup and Initialize the Keyspace if necessary
  err = db.Setup()
  if err != nil {
    // do stuff
  }

  // Write data to the cluster
  err = db.Set("mykey", []byte("My Data"))
  if err != nil {
    // do stuff
  }

  // Fetch the same data
  d, err := db.Get("mykey")
  if err != nil {
    // do stuff
  }
}
```

## Contributing
Thank you for your interest in helping develop Hord. The time, skills, and perspectives you contribute to this project are valued.

Please reference our [Contributing Guide](CONTRIBUTING.md) for details.

## License
[Apache License 2.0](https://choosealicense.com/licenses/apache-2.0/)
