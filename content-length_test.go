package main

import (
	"bytes"
	"github.com/gorilla/mux"
	. "gopkg.in/check.v1"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"vip/fetch"
	"vip/test"
)

var (
	_ = Suite(&ContentLengthSuite{})
)

type ContentLengthSuite struct{}

func (s *ContentLengthSuite) SetUpSuite(c *C) {
	setUpSuite(c)
}

func (s *ContentLengthSuite) SetUpTest(c *C) {
	setUpTest(c)

	storage = test.NewStore()
}

//Check Content-Length of JPG File
func (s *ContentLengthSuite) TestContentLengthJpg(c *C) {
	authToken = "ihopeyoureallyliketokenscausethisisone"

	m := mux.NewRouter()
	m.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))
	f, err := os.Open("./test/exif_test_img.jpg")
	c.Assert(err, IsNil)

	req, err := http.NewRequest("POST", "http://localhost:8080/upload/samplebucket", f)
	c.Assert(err, IsNil)
	fstat, err := os.Stat("./test/exif_test_img.jpg")
	c.Assert(err, IsNil)
	req.ContentLength = fstat.Size()
	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("X-Vip-Token", authToken)

	image, format, err := fetch.GetRotatedImage(req.Body)
	c.Assert(err, IsNil)
	c.Assert(format, Equals, "jpeg")

	data := new(bytes.Buffer)
	err = jpeg.Encode(data, image, nil)
	c.Assert(err, IsNil)
	length := int64(data.Len())
	c.Assert(strconv.FormatInt(length, 10), Equals, "655872")
}

//Check Content-Length of PNG File
func (s *ContentLengthSuite) TestContentLengthPng(c *C) {
	authToken = "ihopeyoureallyliketokenscausethisisone"

	m := mux.NewRouter()
	m.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))

	f, err := os.Open("./test/test_inspiration.png")
	c.Assert(err, IsNil)

	req, err := http.NewRequest("POST", "http://localhost:8080/upload/samplebucket", f)
	c.Assert(err, IsNil)
	fstat, err := os.Stat("./test/test_inspiration.png")
	c.Assert(err, IsNil)
	req.ContentLength = fstat.Size()
	req.Header.Set("Content-Type", "image/png")
	req.Header.Set("X-Vip-Token", authToken)

	raw, err := ioutil.ReadAll(req.Body)
	c.Assert(err, IsNil)

	data := bytes.NewReader(raw)
	length := int64(data.Len())
	c.Assert(err, IsNil)
	c.Assert(strconv.FormatInt(length, 10), Equals, "305197")
}
