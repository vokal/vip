package fetch

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/golang/groupcache"
	"github.com/gorilla/mux"
	"github.com/scottferg/goat"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type CacheContext struct {
	CacheKey string
	ImageId  string
	Bucket   string
	Mime     string
	Width    int
	Goat     *goat.Context
}

type ServingKey struct {
	Id     bson.ObjectId `bson:"_id"`
	Key    string        `bson:"key"`
	Bucket string        `bson:"bucket"`
	Mime   string        `bson:"mime"`
	Url    string        `bson:"url"`
}

func RequestContext(r *http.Request, c *goat.Context) *CacheContext {
	vars := mux.Vars(r)

	width, _ := strconv.Atoi(r.FormValue("s"))
	imageId := vars["image_id"]
	bucket := vars["bucket_id"]

	if width > 720 {
		width = 720
	}

	var cachekey string
	if width == 0 {
		cachekey = fmt.Sprintf("%s/%s", bucket, imageId)
	} else {
		cachekey = fmt.Sprintf("%s/%s/s/%d", bucket, imageId, width)
	}

	return &CacheContext{
		CacheKey: cachekey,
		ImageId:  imageId,
		Bucket:   bucket,
		Width:    width,
		Goat:     c,
	}
}

func findOriginalImage(result *ServingKey, storage ImageStore, c *CacheContext) ([]byte, string, error) {
	err := c.Goat.Database.C("image_serving_keys").Find(bson.M{
		"key": c.ImageId,
	}).One(result)

	if err == nil {
		data, err := storage.Get(result.Bucket, result.Key)
		if err != nil {
			log.Printf("s3 download: %s", err.Error())
			return nil, "", err
		}

		return data, result.Mime, err
	}

	return nil, "", err
}

func findResizedImage(result *ServingKey, storage ImageStore, c *CacheContext) ([]byte, string, error) {
	err := c.Goat.Database.C("image_serving_keys").Find(bson.M{
		"key": fmt.Sprintf("%s/%s/s/%d", c.Bucket, c.ImageId, c.Width),
	}).One(result)

	if err == nil {
		// Strip the bucket out of the cache key
		path := strings.Split(result.Key, c.Bucket+"/")[1]
		data, err := storage.Get(result.Bucket, path)
		if err != nil {
			log.Printf("s3 download: %s", err.Error())
			return nil, "", err
		}

		return data, result.Mime, err
	}

	return nil, "", err
}

func writeResizedImage(buf []byte, storage ImageStore, c *CacheContext) error {
	path := fmt.Sprintf("%s/s/%d", c.ImageId, c.Width)

	key := ServingKey{
		Id:     bson.NewObjectId(),
		Key:    c.CacheKey,
		Bucket: c.Bucket,
		Mime:   c.Mime,
		Url: fmt.Sprintf("https://s3.amazonaws.com/%s/%s",
			c.Bucket, path),
	}

	go func() {
		err := storage.Put(c.Bucket, path, buf, http.DetectContentType(buf))
		if err != nil {
			log.Printf("s3 upload: %s", err.Error())
		}
	}()

	return c.Goat.Database.C("image_serving_keys").Insert(key)
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

func ImageData(storage ImageStore, gc groupcache.Context) ([]byte, error) {
	c, ok := gc.(*CacheContext)
	if !ok {
		return nil, errors.New("invalid context")
	}

	var data []byte
	var result ServingKey
	var err error

	// If the image was requested without any size modifier
	if c.Width == 0 {
		var result ServingKey
		data, c.Mime, err = findOriginalImage(&result, storage, c)
		if err != nil {
			return nil, err
		}

		return data, err
	}

	data, c.Mime, err = findResizedImage(&result, storage, c)
	if err != nil {
		data, c.Mime, err = findOriginalImage(&result, storage, c)
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

		err = writeResizedImage(buf, storage, c)
		if err != nil {
			return nil, err
		}

		log.Println("Retrieved original and stored resized image in S3")
		return buf, err
	}

	log.Println("Retrieved resized image from S3")
	return data, err
}
