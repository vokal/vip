package main

import (
	"bytes"
	. "gopkg.in/check.v1"
	"image"
	"image/jpeg"
	_ "image/png"
	"io/ioutil"
	"vip/fetch"
	"vip/test"
)

var (
	sizes = map[int]int{
		250:  333,
		500:  666,
		160:  213,
		720:  960,
		1024: 1365,
		683:  910,
		431:  574,
	}
	noExifSizes = map[int]int{
		250:  156,
		500:  312,
		160:  100,
		720:  450,
		1024: 640,
		683:  426,
		431:  269,
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
	file, err := ioutil.ReadFile("test/exif_test_img.jpg")
	c.Assert(err, IsNil)

	for width, height := range sizes {
		ctx := &fetch.CacheContext{
			Width: width,
		}

		data := bytes.NewBuffer(file)

		orig, _, err := fetch.GetRotatedImage(data)
		c.Assert(err, IsNil)

		buf := new(bytes.Buffer)
		jpeg.Encode(buf, orig, nil)

		resized, err := fetch.Resize(buf, ctx)
		c.Assert(err, IsNil)

		image, _, err := image.Decode(resized)
		c.Assert(err, IsNil)
		c.Assert(image.Bounds().Size().X, Equals, width)
		c.Assert(image.Bounds().Size().Y, Equals, height)
	}
}

func (s *ResizeSuite) TestResizeNoExifImage(c *C) {
	file, err := ioutil.ReadFile("test/AWESOME.jpg")
	c.Assert(err, IsNil)

	for width, height := range noExifSizes {
		ctx := &fetch.CacheContext{
			Width: width,
		}

		buf := bytes.NewReader(file)
		resized, err := fetch.Resize(buf, ctx)
		c.Assert(err, IsNil)

		image, _, err := image.Decode(resized)
		c.Assert(err, IsNil)
		c.Assert(image.Bounds().Size().X, Equals, width)
		c.Assert(image.Bounds().Size().Y, Equals, height)
	}
}

func (s *ResizeSuite) insertMockImage() (*fetch.CacheContext, error) {
	file, err := ioutil.ReadFile("test/exif_test_img.jpg")
	if err != nil {
		return nil, err
	}

	data := bytes.NewBuffer(file)

	image, _, err := fetch.GetRotatedImage(data)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	jpeg.Encode(buf, image, nil)

	// Push the file data into the mock datastore
	storage.PutReader("test_bucket", "test_id", buf, int64(len(file)), "image/jpeg")

	return &fetch.CacheContext{
		ImageId: "test_id",
		Bucket:  "test_bucket",
	}, err
}

func (s *ResizeSuite) TestOriginalColdCache(c *C) {
	// Open the file once to get it's size
	file, err := ioutil.ReadFile("test/exif_test_img.jpg")
	c.Assert(err, IsNil)

	img, _, err := image.Decode(bytes.NewReader(file))
	c.Assert(err, IsNil)

	// Since this image should be rotated, height should equal
	// width after it's uploaded.
	originalSize := img.Bounds().Size().Y

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

	for width, height := range sizes {
		ctx := &fetch.CacheContext{
			ImageId: mockCtx.ImageId,
			Bucket:  mockCtx.Bucket,
			Width:   width,
		}

		// Run the image resize request
		data, err := fetch.ImageData(storage, ctx)
		c.Assert(err, IsNil)

		// Verify the size of the resulting byte slice
		img, _, err := image.Decode(bytes.NewReader(data))
		c.Assert(err, IsNil)
		c.Assert(img.Bounds().Size().X, Equals, width)
		c.Assert(img.Bounds().Size().Y, Equals, height)
	}
}

func (s *ResizeSuite) TestResizeCropColdCache(c *C) {
	// A single, unresized image is in the database/store
	mockCtx, err := s.insertMockImage()
	c.Assert(err, IsNil)

	for width, height := range sizes {
		ctx := &fetch.CacheContext{
			ImageId: mockCtx.ImageId,
			Bucket:  mockCtx.Bucket,
			Width:   width,
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
		c.Assert(img.Bounds().Size().X <= height, Equals, true)
	}
}
