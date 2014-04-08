package fetch

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/golang/groupcache"
	"github.com/gorilla/mux"
	"github.com/vokalinteractive/vip/store"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"strconv"
)

type FetchWriter interface {
	FindOriginal(storage store.ImageStore, c *CacheContext) ([]byte, string, error)
	FindResized(storage store.ImageStore, c *CacheContext) ([]byte, string, error)
	WriteResized(buf []byte, storage store.ImageStore, c *CacheContext) error
}

type CacheContext struct {
	CacheKey string
	ImageId  string
	Bucket   string
	Mime     string
	Width    int
	Fw       FetchWriter
}

func GetCacheKey(bucket, id string, width int) string {
	if width == 0 {
		return fmt.Sprintf("%s/%s", bucket, id)
	}

	return fmt.Sprintf("%s/%s/s/%d", bucket, id, width)
}

func RequestContext(r *http.Request, fw FetchWriter) *CacheContext {
	vars := mux.Vars(r)

	width, _ := strconv.Atoi(r.FormValue("s"))
	imageId := vars["image_id"]
	bucket := vars["bucket_id"]

	if width > 720 {
		width = 720
	}

	return &CacheContext{
		CacheKey: GetCacheKey(bucket, imageId, width),
		ImageId:  imageId,
		Bucket:   bucket,
		Width:    width,
		Fw:       fw,
	}
}

func Resize(src io.Reader, c *CacheContext) ([]byte, error) {
	image, format, err := image.Decode(src)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	buf := new(bytes.Buffer)

	dst := imaging.Clone(image)

	factor := float64(c.Width) / float64(image.Bounds().Size().X)
	height := int(float64(image.Bounds().Size().Y) * factor)

	dst = imaging.Resize(dst, c.Width, height, imaging.Linear)

	switch format {
	case "jpeg":
		jpeg.Encode(buf, dst, nil)
	case "png":
		err = png.Encode(buf, dst)
	}

	return buf.Bytes(), err
}

func ImageData(storage store.ImageStore, gc groupcache.Context) ([]byte, error) {
	c, ok := gc.(*CacheContext)
	if !ok {
		return nil, errors.New("invalid context")
	}

	var data []byte
	var err error

	// If the image was requested without any size modifier
	if c.Width == 0 {
		data, c.Mime, err = c.Fw.FindOriginal(storage, c)
		if err != nil {
			return nil, err
		}

		return data, err
	}

	data, c.Mime, err = c.Fw.FindResized(storage, c)
	if err != nil {
		data, c.Mime, err = c.Fw.FindOriginal(storage, c)
		if err != nil {
			return nil, err
		}

		// Gifs don't get resized
		if c.Mime == "image/gif" {
			return data, err
		}

		buf, err := Resize(bytes.NewBuffer(data), c)
		if err != nil {
			return nil, err
		}

		err = c.Fw.WriteResized(buf, storage, c)
		if err != nil {
			return nil, err
		}

		log.Println("Retrieved original and stored resized image in S3")
		return buf, err
	}

	log.Println("Retrieved resized image from S3")
	return data, err
}
