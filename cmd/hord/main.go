package main

import (
  "github.com/jessevdk/go-flags"
  "github.com/madflojo/hord/app"
  "fmt"
  "os"
)

type options struct {
  Debug bool   `long:"debug" description:"Enable debug logging"`
  Peers []string  `short:"p" long:"peer" description:"Peer hord instances used for peer to peer cache notifications"`
  Databases []string `short:"d" long:"database" description:"Database instances this hord instance should frontend"`
}

func main() {
  // Parse command line arguments
  args, err := flags.ParseArgs(&opts, os.Args[1:])
  if err != nil {
    os.Exit(1)
  }

  // Run primary application
  err := app.Run()
  if err != nil {
    fmt.Printf("hord stopped: %s\r\n", err)
    if err == app.ErrShutdown {
      os.Exit(0)
    }
    os.Exit(1)
  }
}
