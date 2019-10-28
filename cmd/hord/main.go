package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/madflojo/hord/app"
	"github.com/madflojo/hord/config"
	"github.com/madflojo/hord/databases/cassandra"
	log "github.com/sirupsen/logrus"
	"os"
)

type options struct {
	Debug     bool     `long:"debug" description:"Enable debug logging"`
	Trace     bool     `long:"trace" description:"Enable trace logging (this will impact performance)"`
	Listen    string   `long:"listen" description:"Set the listener address" default:"0.0.0.0"`
	GRPCPort  string   `long:"grpcport" description:"Set custom GRPC Port" default:"9000"`
	HttpPort  string   `long:"httpport" description:"Set custom HTTP Port" default:"9090"`
	Peers     []string `short:"p" long:"peer" description:"Peer hord instances used for peer to peer cache notifications"`
	Databases []string `short:"d" long:"database" description:"Database instances this hord instance should frontend"`
	Keyspace  string   `short:"k" long:"keyspace" description:"Keyspace to use when connecting to Cassandra"`
}

func main() {
	// Parse command line arguments
	var opts options
	_, err := flags.ParseArgs(&opts, os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	// Setup config
	cfg := &config.Config{
		Debug:        opts.Debug,
		Trace:        opts.Trace,
		Peers:        opts.Peers,
		Listen:       opts.Listen,
		GRPCPort:     opts.GRPCPort,
		HttpPort:     opts.HttpPort,
		DatabaseType: "Cassandra",
		Databases: config.Databases{
			Cassandra: &cassandra.Config{
				Hosts:    opts.Databases,
				Keyspace: opts.Keyspace,
			},
		},
	}

	// Run primary application
	err = app.Run(cfg)
	if err != nil {
		if err == app.ErrShutdown {
			log.Infof("Hord stopped - %s", err)
			os.Exit(0)
		}
		log.Errorf("Hord stopped - %s", err)
		os.Exit(2)
	}
}
