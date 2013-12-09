package main

import (
	"bytes"
	"github.com/vokalinteractive/vip/fetch"
	"io/ioutil"
	. "launchpad.net/gocheck"
)

var (
	_ = Suite(&ResizeSuite{})
)

type ResizeSuite struct{}

func (s *ResizeSuite) BenchmarkThumbnailResize(c *C) {
	file, err := ioutil.ReadFile("test/AWESOME.jpg")
	c.Assert(err, IsNil)

	ctx := &fetch.CacheContext{
		Width: 160,
	}

	for i := 0; i < c.N; i++ {
		// Need a new io.Reader on every iteration
		buf := bytes.NewBuffer(file)
		_, err := fetch.Resize(buf, ctx)
		c.Assert(err, IsNil)
	}
}

func (s *ResizeSuite) BenchmarkLargeResize(c *C) {
	file, err := ioutil.ReadFile("test/AWESOME.jpg")
	c.Assert(err, IsNil)

	ctx := &fetch.CacheContext{
		Width: 720,
	}

	for i := 0; i < c.N; i++ {
		// Need a new io.Reader on every iteration
		buf := bytes.NewBuffer(file)
		_, err := fetch.Resize(buf, ctx)
		c.Assert(err, IsNil)
	}
}
