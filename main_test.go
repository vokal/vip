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

	// We'll want to use a test database, not the development database
	g.RegisterMiddleware(g.NewDatabaseMiddleware("localhost", "vip-test"))
}

func setUpTest(c *C) {
	g.CloneDB().DropDatabase()
}

func tearDownSuite(c *C) {
	g.Close()
}
