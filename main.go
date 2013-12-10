package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/golang/groupcache"
	"github.com/scottferg/goat"
	"github.com/vokalinteractive/vip/fetch"
	"github.com/vokalinteractive/vip/peer"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/ec2"
	"launchpad.net/goamz/s3"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	g      *goat.Goat
	cache  *groupcache.Group
	peers  peer.CachePool
	s3conn *s3.S3

	httpport *string = flag.String("httpport", "8080", "target port")
)

func handleImageRequest(w http.ResponseWriter, r *http.Request, c *goat.Context) error {
	start := time.Now()

	gc := fetch.RequestContext(r, c)

	var data []byte
	fmt.Printf("Request for %s from groupcache\n", gc.CacheKey)
	err := cache.Get(gc, gc.CacheKey,
		groupcache.AllocatingByteSliceSink(&data))
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", gc.Mime)
	http.ServeContent(w, r, gc.ImageId, time.Now(), bytes.NewReader(data))

	log.Printf("Request elapsed time (%s): %s", gc.CacheKey, time.Now().Sub(start))

	return err
}

func handlePing(w http.ResponseWriter, r *http.Request, c *goat.Context) error {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "pong")

	return nil
}

func init() {
	flag.Parse()
	g = goat.NewGoat()

	database := os.Getenv("DATABASE_URL")
	if database == "" {
		database = "localhost"
	}

	g.RegisterMiddleware(g.NewDatabaseMiddleware(database, ""))
	g.RegisterRoute("/{bucket_id}/{image_id}", "image_request",
		goat.GET, handleImageRequest)
	g.RegisterRoute("/ping", "ping", goat.GET, handlePing)
}

func main() {
	awsAuth, err := aws.EnvAuth()
	if err != nil {
		log.Fatalf(err.Error())
	}

	s3conn = s3.New(awsAuth, aws.USEast)

	if os.Getenv("DEBUG") == "True" {
		peers = peer.DebugPool()
	} else {
		peers = peer.Pool(ec2.New(awsAuth, aws.USEast))
	}

	peers.SetContext(func(r *http.Request) groupcache.Context {
		return fetch.RequestContext(r, &goat.Context{
			Database: g.CloneDB(),
		})
	})

	cache = groupcache.NewGroup("ImageProxyCache", 64<<20, groupcache.GetterFunc(
		func(c groupcache.Context, key string, dest groupcache.Sink) error {
			log.Printf("Cache MISS for key -> %s", key)
			// Get image data from S3
			data, err := fetch.ImageData(s3conn, c)
			if err != nil {
				return err
			}

			dest.SetBytes(data)
			return nil
		}))

	go peers.Listen()

	go func() {
		log.Printf("Listening on port :%s\n", *httpport)
		if err := g.ListenAndServe(*httpport); err != nil {
			log.Fatalf("Error starting server: %s\n", err.Error())
		}
	}()

	log.Println("Cache listening on port :" + peers.Port())
	server := &http.Server{
		Addr:    ":" + peers.Port(),
		Handler: peers,
	}
	server.ListenAndServe()
}
