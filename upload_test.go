package main

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	. "gopkg.in/check.v1"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"vip/test"
)

var (
	_ = Suite(&UploadSuite{})
)

type UploadSuite struct{}

func (s *UploadSuite) SetUpSuite(c *C) {
	setUpSuite(c)
}

func (s *UploadSuite) SetUpTest(c *C) {
	setUpTest(c)

	storage = test.NewStore()
}

func (s *UploadSuite) TestUpload(c *C) {
	authToken = "lalalatokenlalala"
	os.Setenv("DOMAIN_DATA", "")

	recorder := httptest.NewRecorder()

	// Mock up a router so that mux.Vars are passed
	// correctly
	m := mux.NewRouter()
	m.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))
	f, err := os.Open("./test/awesome.jpeg")
	c.Assert(err, IsNil)

	req, err := http.NewRequest("POST", "http://localhost:8080/upload/samplebucket", f)
	c.Assert(err, IsNil)
	fstat, err := os.Stat("./test/awesome.jpeg")
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
	c.Assert(uri.Path[len(uri.Path)-9:len(uri.Path)], Equals, "-1024x768")
	c.Assert(recorder.HeaderMap["Content-Type"][0], Equals, "application/json")
}

func (s *UploadSuite) TestEmptyUpload(c *C) {
	authToken = "lalalatokenlalala"
	os.Setenv("ALLOWED_ORIGIN", "")

	recorder := httptest.NewRecorder()

	// Mock up a router so that mux.Vars are passed
	// correctly
	m := mux.NewRouter()
	m.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))
	f := &bytes.Reader{}
	req, err := http.NewRequest("POST", "http://localhost:8080/upload/samplebucket", f)
	c.Assert(err, IsNil)

	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("X-Vip-Token", authToken)

	m.ServeHTTP(recorder, req)
	c.Assert(recorder.Code, Equals, http.StatusBadRequest)

	var u ErrorResponse
	err = json.NewDecoder(recorder.Body).Decode(&u)
	c.Assert(err, IsNil)
	c.Assert(u.Msg, Equals, "File must have size greater than 0")
}

func (s *UploadSuite) TestUnauthorizedUpload(c *C) {
	authToken = "lalalatokenlalala"
	os.Setenv("ALLOWED_ORIGIN", "")

	recorder := httptest.NewRecorder()

	// Mock up a router so that mux.Vars are passed
	// correctly
	m := mux.NewRouter()
	m.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))

	f, err := os.Open("./test/awesome.jpeg")
	c.Assert(err, IsNil)

	req, err := http.NewRequest("POST", "http://localhost:8080/upload/samplebucket", f)

	c.Assert(err, IsNil)

	req.Header.Set("Content-Type", "image/jpeg")

	m.ServeHTTP(recorder, req)

	c.Assert(recorder.Code, Equals, http.StatusUnauthorized)
}

func (s *UploadSuite) TestSetOriginData(c *C) {
	authToken = "heyheyheyimatoken"
	os.Setenv("ALLOWED_ORIGIN", "WHATEVER, MAN")

	recorder := httptest.NewRecorder()

	m := mux.NewRouter()
	m.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))

	f, err := os.Open("./test/awesome.jpeg")
	c.Assert(err, IsNil)

	req, err := http.NewRequest("POST", "http://localhost:8080/upload/samplebucket", f)
	c.Assert(err, IsNil)
	fstat, err := os.Stat("./test/awesome.jpeg")
	c.Assert(err, IsNil)
	req.ContentLength = fstat.Size()
	req.Header.Set("Origin", "WHATEVER, MAN")
	c.Assert(err, IsNil)
	req.Header.Set("Content-Type", "image/jpeg")

	m.ServeHTTP(recorder, req)
	c.Assert(recorder.Code, Equals, http.StatusCreated)
}
