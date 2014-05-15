package fetch

import (
	"fmt"
	"github.com/vokalinteractive/vip/store"
	"io"
	"log"
	"net/http"
)

type CacheContext struct {
	ImageId string
	Bucket  string
	Width   int
}

func (c *CacheContext) ReadOriginal(s store.ImageStore) (io.ReadCloser, error) {
	r, err := s.GetReader(c.Bucket, c.ImageId)
	if err != nil {
		log.Printf("s3 download: %s", err.Error())
		return nil, err
	}

	return r, err
}

func (c *CacheContext) ReadResized(s store.ImageStore) (io.ReadCloser, error) {
	data, err := s.GetReader(c.Bucket, c.CacheKey())
	if err != nil {
		log.Printf("s3 download: %s", err.Error())
		return nil, err
	}

	return data, err
}

func (c *CacheContext) WriteResized(buf []byte, s store.ImageStore) error {
	return s.Put(c.Bucket, c.CacheKey(), buf, http.DetectContentType(buf))
}

func (c *CacheContext) CacheKey() string {
	if c.Width == 0 {
		return fmt.Sprintf("%s", c.ImageId)
	}

	return fmt.Sprintf("%s/s/%d", c.ImageId, c.Width)
}
