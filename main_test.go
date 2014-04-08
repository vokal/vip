package main

import (
	"io/ioutil"
	. "launchpad.net/gocheck"
	"log"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

func setUpSuite(c *C) {
	// Silence the logger
	log.SetOutput(ioutil.Discard)
}

func setUpTest(c *C) {}

func tearDownSuite(c *C) {}
