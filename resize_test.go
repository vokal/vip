package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/scottferg/goat"
	"github.com/vokalinteractive/vip/fetch"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	. "launchpad.net/gocheck"
	"sync"
)

var sizes = []int{
	250,
	500,
	160,
	720,
	1024,
	683,
	431,
}

type DebugStore struct {
	store map[string][]byte
}

func NewDebugStore() *DebugStore {
	return &DebugStore{
		store: make(map[string][]byte),
	}
}

func (s *DebugStore) Get(bucket, path string) ([]byte, error) {
	data := s.store[fmt.Sprintf("%s|%s", bucket, path)]
	if data == nil {
		return nil, errors.New("item doesn't exist")
	}

	return data, nil
}

func (s *DebugStore) Put(bucket, path string, data []byte, content string) error {
	s.store[fmt.Sprintf("%s|%s", bucket, path)] = data
	return nil
}

var (
	_ = Suite(&ResizeSuite{})
)

type ResizeSuite struct{}

func (s *ResizeSuite) SetUpSuite(c *C) {
	setUpSuite(c)
}

func (s *ResizeSuite) SetUpTest(c *C) {
	setUpTest(c)

	storage = NewDebugStore()
}

func (s *ResizeSuite) TearDownSuite(c *C) {
	tearDownSuite(c)
}

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

func (s *ResizeSuite) TestResizeImage(c *C) {
	file, err := ioutil.ReadFile("test/AWESOME.jpg")
	c.Assert(err, IsNil)

	for _, size := range sizes {
		ctx := &fetch.CacheContext{
			Width: size,
		}

		buf := bytes.NewBuffer(file)
		resized, err := fetch.Resize(buf, ctx)
		c.Assert(err, IsNil)

		image, _, err := image.Decode(bytes.NewBuffer(resized))
		c.Assert(err, IsNil)
		c.Assert(image.Bounds().Size().X, Equals, size)
	}
}

func (s *ResizeSuite) insertMockImage() (*fetch.CacheContext, error) {
	file, err := ioutil.ReadFile("test/AWESOME.jpg")
	if err != nil {
		return nil, err
	}

	// Push the file data into the mock datastore
	storage.Put("test_bucket", "test_id", file, "image/jpeg")

	// Create a mock serving key in the database
	key := fetch.ServingKey{
		Id:     bson.NewObjectId(),
		Key:    "test_id",
		Bucket: "test_bucket",
		Mime:   "image/jpeg",
	}
	err = g.CloneDB().C("image_serving_keys").Insert(key)

	return &fetch.CacheContext{
		CacheKey: key.Key,
		ImageId:  "test_id",
		Bucket:   key.Bucket,
	}, err
}

func (s *ResizeSuite) TestOriginalColdCache(c *C) {
	// Open the file once to get it's size
	file, err := ioutil.ReadFile("test/AWESOME.jpg")
	c.Assert(err, IsNil)

	img, _, err := image.Decode(bytes.NewBuffer(file))
	c.Assert(err, IsNil)

	originalSize := img.Bounds().Size().X

	// A single, unresized image is in the database/store
	ctx, err := s.insertMockImage()
	c.Assert(err, IsNil)

	// Bootstrap the db connection
	ctx.Goat = &goat.Context{
		Database: g.CloneDB(),
	}

	// Run the image resize request
	data, err := fetch.ImageData(storage, ctx)
	c.Assert(err, IsNil)

	// Verify the size of the resulting byte slice
	img, _, err = image.Decode(bytes.NewBuffer(data))
	c.Assert(err, IsNil)
	c.Assert(img.Bounds().Size().X, Equals, originalSize)

	// Remember to cleanup
	ctx.Goat.Close()
}

func (s *ResizeSuite) TestResizeColdCache(c *C) {
	// A single, unresized image is in the database/store
	mockCtx, err := s.insertMockImage()
	c.Assert(err, IsNil)

	for _, size := range sizes {
		ctx := &fetch.CacheContext{
			CacheKey: fetch.GetCacheKey(mockCtx.Bucket, mockCtx.ImageId, size),
			ImageId:  mockCtx.ImageId,
			Bucket:   mockCtx.Bucket,
			Width:    size,
			Goat: &goat.Context{
				Database: g.CloneDB(),
			},
		}

		// Run the image resize request
		data, err := fetch.ImageData(storage, ctx)
		c.Assert(err, IsNil)

		// Verify the size of the resulting byte slice
		img, _, err := image.Decode(bytes.NewBuffer(data))
		c.Assert(err, IsNil)
		c.Assert(img.Bounds().Size().X, Equals, size)

		// Verify that the resized image was stored in the data store
		data, _, err = fetch.FindResizedImage(storage, ctx)
		c.Assert(err, IsNil)

		img, _, err = image.Decode(bytes.NewBuffer(data))
		c.Assert(err, IsNil)
		c.Assert(img.Bounds().Size().X, Equals, size)

		// Remember to cleanup
		ctx.Goat.Close()
	}
}
