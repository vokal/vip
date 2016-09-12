package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vokal/vip/test"
	. "gopkg.in/check.v1"
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
	recorder := httptest.NewRecorder()
	router := mux.NewRouter()
	router.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))
	f, err := os.Open("./test/exif_test_img.jpg")
	c.Assert(err, IsNil)

	req, err := http.NewRequest("POST", "http://localhost:8080/upload/samplebucket", f)
	c.Assert(err, IsNil)
	fstat, err := os.Stat("./test/exif_test_img.jpg")
	c.Assert(err, IsNil)
	req.ContentLength = fstat.Size()
	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("X-Vip-Token", authToken)

	router.ServeHTTP(recorder, req)

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
}

func (s *UploadSuite) TestBadUploadRequestMethod(c *C) {
	authToken = "lalalatokenlalala"
	recorder := httptest.NewRecorder()
	router := mux.NewRouter()
	// issue non POST requests to route
	router.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))
	f, err := os.Open("./test/exif_test_img.jpg")
	c.Assert(err, IsNil)

	req, err := http.NewRequest("DELETE", "http://localhost:8080/upload/samplebucket", f)
	req.Header.Set("X-Vip-Token", authToken)
	router.ServeHTTP(recorder, req)
	c.Assert(recorder.Code, Equals, http.StatusMethodNotAllowed)

	req, err = http.NewRequest("GET", "http://localhost:8080/upload/samplebucket", f)
	req.Header.Set("X-Vip-Token", authToken)
	router.ServeHTTP(recorder, req)
	c.Assert(recorder.Code, Equals, http.StatusMethodNotAllowed)

	req, err = http.NewRequest("PUT", "http://localhost:8080/upload/samplebucket", f)
	req.Header.Set("X-Vip-Token", authToken)
	router.ServeHTTP(recorder, req)
	c.Assert(recorder.Code, Equals, http.StatusMethodNotAllowed)

	req, err = http.NewRequest("PATCH", "http://localhost:8080/upload/samplebucket", f)
	req.Header.Set("X-Vip-Token", authToken)
	router.ServeHTTP(recorder, req)
	c.Assert(recorder.Code, Equals, http.StatusMethodNotAllowed)

}

func (s *UploadSuite) TestImageTooBig(c *C) {
	authToken = "lalalatokenlalala"
	recorder := httptest.NewRecorder()
	router := mux.NewRouter()
	router.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))
	f, err := os.Open("./test/exif_test_img.jpg")
	c.Assert(err, IsNil)

	req, err := http.NewRequest("POST", "http://localhost:8080/upload/samplebucket", f)
	c.Assert(err, IsNil)
	fstat, err := os.Stat("./test/exif_test_img.jpg")
	c.Assert(err, IsNil)
	req.ContentLength = fstat.Size()
	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("X-Vip-Token", authToken)

	limit = 1
	router.ServeHTTP(recorder, req)
	limit = 10
	c.Assert(recorder.Code, Equals, http.StatusRequestEntityTooLarge)

}

func (s *UploadSuite) TestPing(c *C) {
	authToken = "lalalatokenlalala"
	recorder := httptest.NewRecorder()
	router := mux.NewRouter()
	router.Handle("/ping", verifyAuth(handlePing))
	req, err := http.NewRequest("POST", "http://localhost:8080/ping", nil)
	req.Header.Set("X-Vip-Token", authToken)
	c.Assert(err, IsNil)
	router.ServeHTTP(recorder, req)
	c.Assert(recorder.Code, Equals, http.StatusOK)
}

func (s *UploadSuite) TestUploadWarmup(c *C) {
	authToken = "lalalatokenlalala"
	recorder := httptest.NewRecorder()
	router := mux.NewRouter()
	router.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))
	f, err := os.Open("./test/exif_test_img.jpg")
	c.Assert(err, IsNil)

	req, err := http.NewRequest("POST", "http://localhost:8080/upload/samplebucket", f)
	c.Assert(err, IsNil)
	fstat, err := os.Stat("./test/exif_test_img.jpg")
	c.Assert(err, IsNil)
	req.ContentLength = fstat.Size()
	req.Header.Set("X-Vip-Warmup", "s=3,s=100&c=true")
	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("X-Vip-Token", authToken)

	router.ServeHTTP(recorder, req)

	var u UploadResponse
	err = json.NewDecoder(recorder.Body).Decode(&u)
	c.Assert(err, IsNil)
}

func (s *UploadSuite) TestSecureWarmup(c *C) {
	authToken = "lalalatokenlalala"
	recorder := httptest.NewRecorder()
	router := mux.NewRouter()
	router.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))
	f, err := os.Open("./test/exif_test_img.jpg")
	c.Assert(err, IsNil)

	req, err := http.NewRequest("POST", "https://localhost:8080/upload/samplebucket", f)
	req.Header.Set("X-Vip-Token", authToken)
	c.Assert(err, IsNil)
	router.ServeHTTP(recorder, req)

	var u UploadResponse
	err = json.NewDecoder(recorder.Body).Decode(&u)
	c.Assert(err, IsNil)
}

func (s *UploadSuite) TestEmptyUpload(c *C) {
	authToken = "lalalatokenlalala"
	recorder := httptest.NewRecorder()
	router := mux.NewRouter()
	os.Setenv("ALLOWED_ORIGIN", "")

	router.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))
	f := &bytes.Reader{}
	req, err := http.NewRequest("POST", "http://localhost:8080/upload/samplebucket", f)
	c.Assert(err, IsNil)

	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("X-Vip-Token", authToken)

	router.ServeHTTP(recorder, req)
	c.Assert(recorder.Code, Equals, http.StatusBadRequest)

	var u ErrorResponse
	err = json.NewDecoder(recorder.Body).Decode(&u)
	c.Assert(err, IsNil)
	c.Assert(u.Msg, Equals, "File must have size greater than 0")
}

func (s *UploadSuite) TestUnauthorizedUpload(c *C) {
	recorder := httptest.NewRecorder()
	router := mux.NewRouter()
	router.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))

	f, err := os.Open("./test/awesome.jpeg")
	c.Assert(err, IsNil)

	req, err := http.NewRequest("POST", "http://localhost:8080/upload/samplebucket", f)
	c.Assert(err, IsNil)
	fstat, err := os.Stat("./test/awesome.jpeg")
	c.Assert(err, IsNil)
	req.ContentLength = fstat.Size()
	req.Header.Set("Content-Type", "image/jpeg")

	router.ServeHTTP(recorder, req)

	c.Assert(recorder.Code, Equals, http.StatusUnauthorized)
}

func (s *UploadSuite) TestSetOriginData(c *C) {
	authToken = "lalalatokenlalala"
	recorder := httptest.NewRecorder()
	router := mux.NewRouter()
	origins = []string{"localhost", "*.vokal.io"}

	router.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))

	f, err := os.Open("./test/awesome.jpeg")
	c.Assert(err, IsNil)

	req, err := http.NewRequest("POST", "http://localhost:8080/upload/samplebucket", f)
	c.Assert(err, IsNil)
	fstat, err := os.Stat("./test/awesome.jpeg")
	c.Assert(err, IsNil)
	req.ContentLength = fstat.Size()
	req.Header.Set("Origin", "http://images.vokal.io")
	req.Header.Set("Content-Type", "image/jpeg")

	router.ServeHTTP(recorder, req)
	c.Assert(recorder.Code, Equals, http.StatusCreated)
}

// Test localhost with a port number
func (s *UploadSuite) TestSetOriginDataLocalhost(c *C) {
	authToken = "lalalatokenlalala"
	recorder := httptest.NewRecorder()
	router := mux.NewRouter()
	origins = []string{"localhost", "*.vokal.io"}
	router.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))

	f, err := os.Open("./test/awesome.jpeg")
	c.Assert(err, IsNil)

	req, err := http.NewRequest("POST", "http://localhost:8080/upload/samplebucket", f)
	c.Assert(err, IsNil)
	fstat, err := os.Stat("./test/awesome.jpeg")
	c.Assert(err, IsNil)
	req.ContentLength = fstat.Size()
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Content-Type", "image/jpeg")

	router.ServeHTTP(recorder, req)
	c.Assert(recorder.Code, Equals, http.StatusCreated)
}

func (s *UploadSuite) TestRespondCorsHeaders(c *C) {
	authToken = "lalalatokenlalala"
	recorder := httptest.NewRecorder()
	router := mux.NewRouter()
	origins = []string{"localhost", "*.vokal.io"}

	router.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))

	f, err := os.Open("./test/awesome.jpeg")
	c.Assert(err, IsNil)

	req, err := http.NewRequest("OPTIONS", "http://localhost:8080/upload/samplebucket", f)
	c.Assert(err, IsNil)

	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Content-Type", "image/jpeg")

	router.ServeHTTP(recorder, req)
	c.Assert(recorder.Code, Equals, http.StatusOK)
	c.Assert(recorder.HeaderMap.Get("Access-Control-Allow-Origin"), Equals, "*")
}

//Check Content-Length of JPG File
func (s *UploadSuite) TestContentLengthJpg(c *C) {
	f, err := os.Open("./test/exif_test_img.jpg")
	c.Assert(err, IsNil)

	fstat, err := os.Stat("./test/exif_test_img.jpg")
	c.Assert(err, IsNil)

	data, err := processFile(f, "image/jpeg", "")
	c.Assert(err, IsNil)
	c.Assert(data.Length, Not(Equals), fstat.Size())
}

//Check Content-Length of PNG File
func (s *UploadSuite) TestContentLengthPng(c *C) {
	f, err := os.Open("./test/test_inspiration.png")
	c.Assert(err, IsNil)

	fstat, err := os.Stat("./test/test_inspiration.png")
	c.Assert(err, IsNil)

	data, err := processFile(f, "image/png", "")
	c.Assert(err, IsNil)
	c.Assert(data.Length, Equals, fstat.Size())
}
