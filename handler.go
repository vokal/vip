package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/golang/groupcache"
	"github.com/gorilla/mux"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"
	"vip/fetch"
)

type UploadResponse struct {
	Url string `json:"url"`
}

type verifyAuth func(http.ResponseWriter, *http.Request)

func (h verifyAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enable cross-origin requests
	// TODO: Whitelist the domain via an environment variable
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-Vip-Token, Authorization")
	}

	if r.Method == "OPTIONS" {
		return
	}

	auth := r.Header.Get("X-Vip-Token")
	if auth != authToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h(w, r)
}

func fileKey(bucket string) string {
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	key := fmt.Sprintf("%d-%s-%d", seed.Int63(), bucket, time.Now().UnixNano())

	hash := md5.New()
	io.WriteString(hash, key)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func handleImageRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000")

	// Client is checking for a cached URI, assume it is valid
	// and return a 304
	if r.Header.Get("If-Modified-Since") != "" {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	gc := fetch.RequestContext(r)

	var data []byte
	err := cache.Get(gc, gc.CacheKey(), groupcache.AllocatingByteSliceSink(&data))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", http.DetectContentType(data))
	http.ServeContent(w, r, gc.ImageId, time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), bytes.NewReader(data))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket_id"]

	// Set a hard 5mb limit on files
	if r.ContentLength > 5<<20 {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	key := fileKey(bucket)
	err := storage.PutReader(bucket, key, r.Body,
		r.ContentLength, r.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uri := r.URL

	if r.URL.Host == "" {
		uri.Host = os.Getenv("URI_HOSTNAME")
		uri.Scheme = r.URL.Scheme
	}

	uri.Path = fmt.Sprintf("%s/%s", bucket, key)

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(UploadResponse{
		Url: uri.String(),
	})
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "pong")
}
