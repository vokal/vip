package main

import (
	"flag"
	"fmt"
	"github.com/golang/groupcache"
	"github.com/gorilla/mux"
	"github.com/vokalinteractive/go-loggly"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/ec2"
	"launchpad.net/goamz/s3"
	"log"
	"net/http"
	"os"
	"vip/fetch"
	"vip/peer"
	"vip/store"
)

var (
	cache     *groupcache.Group
	peers     peer.CachePool
	storage   store.ImageStore
	authToken string

	httpport *string = flag.String("httpport", "8080", "target port")
)

func listenHttp() {
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
}

func getRegion() aws.Region {
	region := os.Getenv("AWS_REGION")

	switch region {
	case "us-west-1":
		return aws.USWest
	case "us-west-2":
		return aws.USWest2
	case "us-east-1":
		return aws.USEast
	default:
		log.Println("No AWS_REGION parameter provided, defaulting to us-east-1")
		return aws.USEast
	}
}

func init() {
	flag.Parse()

	loggly_key := os.Getenv("LOGGLY_KEY")
	if loggly_key != "" {
		log.SetOutput(loggly.New(loggly_key, "vip"))
	}

	r := mux.NewRouter()

	authToken = os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		log.Println("No AUTH_TOKEN parameter provided, uploads are insecure")
	}

	r.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))
	r.HandleFunc("/{bucket_id}/{image_id}", handleImageRequest)
	r.HandleFunc("/ping", handlePing)
	http.Handle("/", r)
}

func main() {
	awsAuth, err := aws.EnvAuth()
	if err != nil {
		log.Fatalf(err.Error())
	}

	s3conn := s3.New(awsAuth, getRegion())
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
	go listenHttp()

	log.Println("Cache listening on port :" + peers.Port())
	s := &http.Server{
		Addr:    ":" + peers.Port(),
		Handler: peers,
	}
	s.ListenAndServe()
}
