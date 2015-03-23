package main

import (
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"runtime"
	"vip/fetch"
	"vip/peer"
	"vip/q"
	"vip/store"

	"github.com/bradfitz/http2"
	"github.com/golang/groupcache"
	"github.com/gorilla/mux"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

const (
	KeyFilePath  = "/etc/vip/application.key"
	CertFilePath = "/etc/vip/application.pem"
)

var (
	cache     *groupcache.Group
	peers     peer.CachePool
	storage   store.ImageStore
	authToken string
	origins   []string
	verbose   *bool   = flag.Bool("verbose", false, "verbose logging")
	httpport  *string = flag.String("httpport", "8080", "target port")
	secure    bool    = false
	Queue     q.Queue
)

func listenHttp() {
	log.Printf("Listening on port :%s\n", *httpport)

	port := fmt.Sprintf(":%s", *httpport)

	if secure {
		log.Println("Serving via TLS")
		server := &http.Server{Addr: port, Handler: nil}

		if os.Getenv("DISABLE_HTTP2") == "" {
			http2.ConfigureServer(server, nil)
		}

		if err := server.ListenAndServeTLS(CertFilePath, KeyFilePath); err != nil {
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
	aws_region, ok := aws.Regions[region]
	if ok {
		return aws_region
	} else {
		log.Printf(
			"\"%s\" is not a valid AWS_REGION parameter provided, defaulting to us-east-1",
			region)
		return aws.USEast
	}
}

func init() {
	flag.Parse()
	var err error
	hasKey := true
	hasCert := true
	_, err = os.Stat(KeyFilePath)
	if err != nil {
		log.Printf("No key found at %s\n", KeyFilePath)
		hasKey = false
	}

	_, err = os.Stat(CertFilePath)
	if err != nil {
		log.Printf("No certificate found at %s\n", CertFilePath)
		hasCert = false
	}

	secure = hasCert && hasKey
	Queue = q.New(100)

	r := mux.NewRouter()
	authToken = os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		log.Println("No AUTH_TOKEN parameter provided, uploads are insecure")
	}

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		log.Println("No ALLOWED_ORIGIN set, CORS support is disabled.")
	} else {
		origins = strings.Split(allowedOrigin, ",")
	}

	r.Handle("/upload/{bucket_id}", verifyAuth(handleUpload))
	r.HandleFunc("/{bucket_id}/{image_id}/warmup", handleWarmup)
	r.HandleFunc("/{bucket_id}/{image_id}", handleImageRequest)
	r.HandleFunc("/ping", handlePing)
	http.Handle("/", r)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	awsAuth, err := aws.EnvAuth()
	if err != nil {
		log.Fatalf(err.Error())
	}

	s3conn := s3.New(awsAuth, getRegion())
	storage = store.NewS3Store(s3conn)

	peers = peer.DebugPool()

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

	if !*verbose {
		logwriter, err := syslog.Dial("udp", "app_syslog:514", syslog.LOG_NOTICE, "vip")
		if err != nil {
			log.Println(err.Error())
			log.Println("using default logger")
		} else {
			log.SetOutput(logwriter)
		}
	}

	go peers.Listen()
	go listenHttp()
	go Queue.Start(4)
	log.Println("Cache listening on port :" + peers.Port())
	s := &http.Server{
		Addr:    ":" + peers.Port(),
		Handler: peers,
	}
	s.ListenAndServe()
}
