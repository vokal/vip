package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/golang/groupcache"
	"github.com/gorilla/mux"
	"github.com/vokalinteractive/vip/fetch"
	"github.com/vokalinteractive/vip/peer"
	"github.com/vokalinteractive/vip/store"
	"go-loggly"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/ec2"
	"launchpad.net/goamz/s3"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	cache   *groupcache.Group
	peers   peer.CachePool
	storage store.ImageStore

	httpport *string = flag.String("httpport", "8080", "target port")
)

func handleImageRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Client is checking for a cached URI, assume it is valid
	// and return a 304
	if r.Header.Get("If-Modified-Since") != "" {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		w.WriteHeader(http.StatusNotModified)
		log.Printf("Request elapsed time : %s", time.Now().Sub(start))
		return
	}

	gc := fetch.RequestContext(r)

	fmt.Printf("Request for %s from groupcache\n", gc.CacheKey())

	c := time.Now()
	var data []byte
	err := cache.Get(gc, gc.CacheKey(), groupcache.AllocatingByteSliceSink(&data))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Printf("Cache fetch elapsed time (%s): %s", gc.CacheKey(), time.Now().Sub(c))

	send := time.Now()
	w.Header().Set("Content-Type", http.DetectContentType(data))
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	http.ServeContent(w, r, gc.ImageId, time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), bytes.NewReader(data))

	log.Printf("Transmit elapsed time (%s): %s", gc.CacheKey(), time.Now().Sub(send))
	log.Printf("Request elapsed time (%s): %s", gc.CacheKey(), time.Now().Sub(start))

	return
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "pong")

	return
}

func init() {
	flag.Parse()

	loggly_key := os.Getenv("LOGGLY_KEY")
	if loggly_key != "" {
		log.SetOutput(loggly.New(loggly_key, "vip"))
	}

	r := mux.NewRouter()
	r.HandleFunc("/{bucket_id}/{image_id}", handleImageRequest)
	r.HandleFunc("/ping", handlePing)
	http.Handle("/", r)
}

func main() {
	awsAuth, err := aws.EnvAuth()
	if err != nil {
		log.Fatalf(err.Error())
	}

	s3conn := s3.New(awsAuth, aws.USEast)
	storage = store.NewS3Store(s3conn)

	if os.Getenv("DEBUG") == "True" {
		peers = peer.DebugPool()
	} else {
		peers = peer.Pool(ec2.New(awsAuth, aws.USEast))
	}

	peers.SetContext(func(r *http.Request) groupcache.Context {
		return fetch.RequestContext(r)
	})

	cache = groupcache.NewGroup("ImageProxyCache", 64<<20, groupcache.GetterFunc(
		func(c groupcache.Context, key string, dest groupcache.Sink) error {
			log.Printf("Cache MISS for key -> %s", key)
			// Get image data from S3
			b, err := fetch.ImageData(storage, c)
			if err != nil {
				return err
			}

			return dest.SetBytes(b)
		}))

	go peers.Listen()

	go func() {
		log.Printf("Listening on port :%s\n", *httpport)
		cert := os.Getenv("SSL_CERT")
		key := os.Getenv("SSL_KEY")

		port := fmt.Sprintf(":%s", *httpport)

		if cert != "" && key != "" {
			log.Println("Serving via SSL")
			if err := http.ListenAndServeTLS(port, cert, key, nil); err != nil {
				log.Fatalf("Error starting server: %s\n", err.Error())
			}
		} else {
			if err := http.ListenAndServe(port, nil); err != nil {
				log.Fatalf("Error starting server: %s\n", err.Error())
			}
		}
	}()

	log.Println("Cache listening on port :" + peers.Port())
	s := &http.Server{
		Addr:    ":" + peers.Port(),
		Handler: peers,
	}
	s.ListenAndServe()
}
