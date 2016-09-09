package fetch

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/vokal/vip/store"
)

type CacheContext struct {
	ImageId string
	Bucket  string
	Width   int
	Crop    bool
}

func (c *CacheContext) ReadOriginal(s store.ImageStore) (io.ReadCloser, error) {
	r, err := s.GetReader(c.Bucket, c.ImageId)
	if err != nil {
		log.Printf("s3 download: %s", err.Error())
		return nil, err
	}

	return r, err
}

func (c *CacheContext) ReadModified(s store.ImageStore) (io.ReadCloser, error) {
	data, err := s.GetReader(c.Bucket, c.CacheKey())
	if err != nil {
		log.Printf("s3 download: %s", err.Error())
		return nil, err
	}

	return data, err
}

func (c *CacheContext) WriteModified(buf []byte, s store.ImageStore) error {
	return s.Put(c.Bucket, c.CacheKey(), buf, http.DetectContentType(buf))
}

func (c *CacheContext) CacheKey() string {
	switch {
	case c.Crop && c.Width != 0:
		return fmt.Sprintf("%s/c/s/%d", c.ImageId, c.Width)
	case c.Width != 0:
		return fmt.Sprintf("%s/s/%d", c.ImageId, c.Width)
	}

	return fmt.Sprintf("%s", c.ImageId)
}
