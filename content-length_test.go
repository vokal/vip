package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	. "gopkg.in/check.v1"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
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
	os.Setenv("DOMAIN_DATA", "")

	recorder := httptest.NewRecorder()
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

	m.ServeHTTP(recorder, req)

	var u UploadResponse
	err = json.NewDecoder(recorder.Body).Decode(&u)
	c.Assert(err, IsNil)
	c.Assert(len(u.Url), Not(Equals), 0)

	uri, err := url.Parse(u.Url)
	c.Assert(err, IsNil)

	c.Assert(uri.Scheme, Equals, "http")
	c.Assert(uri.Host, Equals, "localhost:8080")
	c.Assert(uri.Path[1:13], Equals, "samplebucket")
	c.Assert(strings.HasSuffix(uri.Path, "-2448x3264"), Equals, true)
	c.Assert(recorder.HeaderMap["Content-Type"][0], Equals, "application/json")
	c.Assert(recorder.HeaderMap["Content-Length"][0], Equals, "655872")
}

//Check Content-Length of PNG File
func (s *ContentLengthSuite) TestContentLengthPng(c *C) {
	authToken = "ihopeyoureallyliketokenscausethisisone"
	os.Setenv("DOMAIN_DATA", "")

	recorder := httptest.NewRecorder()
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
	m.ServeHTTP(recorder, req)

	var u UploadResponse
	err = json.NewDecoder(recorder.Body).Decode(&u)
	c.Assert(err, IsNil)
	c.Assert(len(u.Url), Not(Equals), 0)

	uri, err := url.Parse(u.Url)
	c.Assert(err, IsNil)

	c.Assert(uri.Scheme, Equals, "http")
	c.Assert(uri.Host, Equals, "localhost:8080")
	c.Assert(uri.Path[1:13], Equals, "samplebucket")
	c.Assert(strings.HasSuffix(uri.Path, "-1142x781"), Equals, true)
	c.Assert(recorder.HeaderMap["Content-Type"][0], Equals, "application/json")
	c.Assert(recorder.HeaderMap["Content-Length"][0], Equals, "305197")
}
