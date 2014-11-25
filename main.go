package main

import (
	"flag"
	"fmt"
	"github.com/golang/groupcache"
	"github.com/gorilla/mux"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"vip/fetch"
	"vip/peer"
	"vip/store"
)

const (
	KeyFilePath  = "/etc/vip/application.key"
	CertFilePath = "/etc/vip/application.cert"
)

var (
	cache     *groupcache.Group
	peers     peer.CachePool
	storage   store.ImageStore
	authToken string

	verbose  *bool   = flag.Bool("verbose", false, "verbose logging")
	httpport *string = flag.String("httpport", "8080", "target port")
	cert     string  = ""
	key      string  = ""
	secure   bool    = false
)

func listenHttp() {
	log.Printf("Listening on port :%s\n", *httpport)

	port := fmt.Sprintf(":%s", *httpport)

	if cert != "" && key != "" {
		log.Println("Serving via TSL")
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

func getFile(path string) (string, error) {
	val := ""
	if _, err := os.Stat(path); err == nil {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return val, err
		}
		val = string(b[:])
	}
	return val, nil
}

func init() {
	flag.Parse()
	var err error

	key, err = getFile(KeyFilePath)
	if err != nil {
		log.Fatal(err.Error())
	}

	cert, err = getFile(CertFilePath)
	if err != nil {
		log.Fatal(err.Error())
	}

	secure = (key != "" && cert != "")

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
			log.Fatal(err.Error())
		}
		log.SetOutput(logwriter)
	}

	go peers.Listen()
	go listenHttp()

	log.Println("Cache listening on port :" + peers.Port())
	s := &http.Server{
		Addr:    ":" + peers.Port(),
		Handler: peers,
	}
	s.ListenAndServe()
}
