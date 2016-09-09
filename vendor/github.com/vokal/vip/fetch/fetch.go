package fetch

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/vokal/vip/store"

	"github.com/golang/groupcache"
	"github.com/gorilla/mux"
)

var maxWidth = getMaxWidth()

func getMaxWidth() int {
	maxWidth, err := strconv.Atoi(os.Getenv("VIP_MAX_WIDTH"))
	if err != nil {
		return 720
	}

	return maxWidth
}

func RequestContext(r *http.Request) *CacheContext {
	vars := mux.Vars(r)

	width, _ := strconv.Atoi(r.FormValue("s"))

	if width > maxWidth {
		width = maxWidth
	}

	return &CacheContext{
		ImageId: vars["image_id"],
		Bucket:  vars["bucket_id"],
		Width:   width,
		Crop:    strings.ToLower(r.FormValue("c")) == "true",
	}
}

func readImage(r io.Reader) ([]byte, error) {
	var b bytes.Buffer
	_, err := b.ReadFrom(r)

	return b.Bytes(), err
}

func ImageData(storage store.ImageStore, gc groupcache.Context) ([]byte, error) {
	c, ok := gc.(*CacheContext)
	if !ok {
		return nil, errors.New("invalid context")
	}

	var reader io.ReadCloser
	var err error

	defer func() {
		if reader != nil {
			reader.Close()
		}
	}()

	resp, err := storage.Head(c.Bucket, c.ImageId)
	if err != nil {
		// Don't break on an error
		log.Println(err)
	}

	reader, err = c.ReadModified(storage)
	if err != nil {
		reader, err = c.ReadOriginal(storage)
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("Retrieved resized image from S3")
		return readImage(reader)
	}

	var buf io.Reader
	if c.Width != 0 {
		if resp != nil && resp.Header.Get("Content-Type") == "image/gif" {
			buf, err = ResizeGif(reader, c)
		} else {
			buf, err = Resize(reader, c)
		}
		if err != nil {
			return nil, err
		}
	}

	result, err := readImage(buf)
	if err != nil {
		return nil, err
	}

	go func() {
		err = c.WriteModified(result, storage)
		if err != nil {
			log.Printf("s3 upload: %s", err.Error())
		}
	}()

	log.Println("Retrieved original and stored resized image in S3")
	return result, err
}
