package main

import (
	"bytes"
	. "gopkg.in/check.v1"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"vip/fetch"
	"vip/test"
)

var (
	sizes = []int{
		250,
		500,
		160,
		720,
		1024,
		683,
		431,
	}
)

var (
	_ = Suite(&ResizeSuite{})
)

type ResizeSuite struct{}

func (s *ResizeSuite) SetUpSuite(c *C) {
	setUpSuite(c)
}

func (s *ResizeSuite) SetUpTest(c *C) {
	setUpTest(c)

	storage = test.NewStore()
}

func (s *ResizeSuite) BenchmarkThumbnailResize(c *C) {
	file, err := ioutil.ReadFile("test/AWESOME.jpg")
	c.Assert(err, IsNil)

	ctx := &fetch.CacheContext{
		Width: 160,
	}

	for i := 0; i < c.N; i++ {
		// Need a new io.Reader on every iteration
		buf := bytes.NewReader(file)
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
		buf := bytes.NewReader(file)
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

		buf := bytes.NewReader(file)
		resized, err := fetch.Resize(buf, ctx)
		c.Assert(err, IsNil)

		image, _, err := image.Decode(resized)
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

	return &fetch.CacheContext{
		ImageId: "test_id",
		Bucket:  "test_bucket",
	}, err
}

func (s *ResizeSuite) TestOriginalColdCache(c *C) {
	// Open the file once to get it's size
	file, err := ioutil.ReadFile("test/AWESOME.jpg")
	c.Assert(err, IsNil)

	img, _, err := image.Decode(bytes.NewReader(file))
	c.Assert(err, IsNil)

	originalSize := img.Bounds().Size().X

	// A single, unresized image is in the database/store
	ctx, err := s.insertMockImage()
	c.Assert(err, IsNil)

	// Run the image resize request
	data, err := fetch.ImageData(storage, ctx)
	c.Assert(err, IsNil)

	// Verify the size of the resulting byte slice
	img, _, err = image.Decode(bytes.NewReader(data))
	c.Assert(err, IsNil)
	c.Assert(img.Bounds().Size().X, Equals, originalSize)
}

func (s *ResizeSuite) TestResizeColdCache(c *C) {
	// A single, unresized image is in the database/store
	mockCtx, err := s.insertMockImage()
	c.Assert(err, IsNil)

	for _, size := range sizes {
		ctx := &fetch.CacheContext{
			ImageId: mockCtx.ImageId,
			Bucket:  mockCtx.Bucket,
			Width:   size,
		}

		// Run the image resize request
		data, err := fetch.ImageData(storage, ctx)
		c.Assert(err, IsNil)

		// Verify the size of the resulting byte slice
		img, _, err := image.Decode(bytes.NewReader(data))
		c.Assert(err, IsNil)
		c.Assert(img.Bounds().Size().X, Equals, size)
	}
}

func (s *ResizeSuite) TestResizeCropColdCache(c *C) {
	// A single, unresized image is in the database/store
	mockCtx, err := s.insertMockImage()
	c.Assert(err, IsNil)

	for _, size := range sizes {
		ctx := &fetch.CacheContext{
			ImageId: mockCtx.ImageId,
			Bucket:  mockCtx.Bucket,
			Width:   size,
			Crop:    true,
		}

		// Run the image resize request
		data, err := fetch.ImageData(storage, ctx)
		c.Assert(err, IsNil)

		// Verify the size of the resulting byte slice
		img, _, err := image.Decode(bytes.NewReader(data))
		c.Assert(err, IsNil)
		c.Assert(img.Bounds().Size().X, Equals, img.Bounds().Size().Y)
		c.Assert(img.Bounds().Size().X > 0, Equals, true)
		c.Assert(img.Bounds().Size().X <= size, Equals, true)
	}
}
