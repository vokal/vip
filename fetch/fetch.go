package fetch

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/golang/groupcache"
	"github.com/gorilla/mux"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"strconv"
	"vip/store"
)

func RequestContext(r *http.Request) *CacheContext {
	vars := mux.Vars(r)

	width, _ := strconv.Atoi(r.FormValue("s"))
	imageId := vars["image_id"]
	bucket := vars["bucket_id"]

	if width > 720 {
		width = 720
	}

	return &CacheContext{
		ImageId: imageId,
		Bucket:  bucket,
		Width:   width,
	}
}

func Resize(src io.Reader, c *CacheContext) (io.Reader, error) {
	image, format, err := image.Decode(src)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	buf := new(bytes.Buffer)

	factor := float64(c.Width) / float64(image.Bounds().Size().X)
	height := int(float64(image.Bounds().Size().Y) * factor)

	image = imaging.Resize(image, c.Width, height, imaging.Linear)

	switch format {
	case "jpeg":
		jpeg.Encode(buf, image, nil)
	case "png":
		err = png.Encode(buf, image)
	}

	return buf, err
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

	// If the image was requested without any size modifier
	if c.Width == 0 {
		reader, err = c.ReadOriginal(storage)
		if err != nil {
			return nil, err
		}

		return readImage(reader)
	}

	reader, err = c.ReadResized(storage)
	if err != nil {
		reader, err := c.ReadOriginal(storage)
		if err != nil {
			return nil, err
		}

		// Gifs don't get resized
		/* TODO: Detect mimetype earlier
		if c.Mime == "image/gif" {
			return data, err
		}
		*/

		buf, err := Resize(reader, c)
		if err != nil {
			return nil, err
		}

		result, err := readImage(buf)
		if err != nil {
			return nil, err
		}

		go func() {
			err = c.WriteResized(result, storage)
			if err != nil {
				log.Printf("s3 upload: %s", err.Error())
			}
		}()

		log.Println("Retrieved original and stored resized image in S3")
		return result, err
	}

	log.Println("Retrieved resized image from S3")
	return readImage(reader)
}
