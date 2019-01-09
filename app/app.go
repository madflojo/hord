// Package app is the primary application package for Hord.
//
// This package handles all of the primary application responsibilities. These include request handling, logging, and 
// basic runtime functionality.
package app

import (
  "log"
  "errors"
)

var ErrShutdown = errors.New("System was shutdown")

func Run() error {
  log.Printf("I am a little teapot, here is my handle here is my spout.")
  return ErrShutdown
}
