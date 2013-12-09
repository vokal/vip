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
	"launchpad.net/goamz/s3"
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

	var bucket string
	var imageId string
	var width int

	if len(vars) == 0 {
		path := strings.Split(r.URL.Path, "/")
		bucket = path[3]
		imageId = path[4]

		if strings.Index(imageId, "?") > -1 {
			imageId = strings.Split(imageId, "?")[0]
		}

		querystring := strings.Split(r.URL.String(), "=")
		if len(querystring) > 1 {
			width, _ = strconv.Atoi(querystring[1])
		}
	} else {
		width, _ = strconv.Atoi(r.FormValue("s"))
		imageId = vars["image_id"]
		bucket = vars["bucket_id"]
	}

	if width > 720 {
		width = 720
	}

	var cachekey string
	if width == 0 {
		cachekey = fmt.Sprintf("%s/%s", bucket, imageId)
	} else {
		cachekey = fmt.Sprintf("%s/%s/s/%d", bucket, imageId, width)
	}

	ctx := c
	if ctx == nil {
		/*
			ctx = &goat.Context{
				Database: g.CloneDB(),
			}
		*/
		log.Fatalf("No context")
	}

	return &CacheContext{
		CacheKey: cachekey,
		ImageId:  imageId,
		Bucket:   bucket,
		Width:    width,
		Goat:     ctx,
	}
}

func findOriginalImage(result *ServingKey, s3conn *s3.S3, c *CacheContext) ([]byte, string, error) {
	err := c.Goat.Database.C("image_serving_keys").Find(bson.M{
		"key": c.ImageId,
	}).One(result)

	if err == nil {
		bucket := s3conn.Bucket(result.Bucket)
		data, err := bucket.Get(result.Key)
		if err != nil {
			log.Printf("s3 download: %s", err.Error())
			return nil, "", err
		}

		return data, result.Mime, err
	}

	return nil, "", err
}

func findResizedImage(result *ServingKey, s3conn *s3.S3, c *CacheContext) ([]byte, string, error) {
	err := c.Goat.Database.C("image_serving_keys").Find(bson.M{
		"key": fmt.Sprintf("%s/%s/s/%d", c.Bucket, c.ImageId, c.Width),
	}).One(result)

	if err == nil {
		bucket := s3conn.Bucket(result.Bucket)
        // Strip the bucket out of the cache key
		data, err := bucket.Get(strings.Split(result.Key, c.Bucket+"/")[1])
		if err != nil {
			log.Printf("s3 download: %s", err.Error())
			return nil, "", err
		}

		return data, result.Mime, err
	}

	return nil, "", err
}

func writeResizedImage(buf []byte, s3conn *s3.S3, c *CacheContext) error {
	path := fmt.Sprintf("%s/s/%d", c.ImageId, c.Width)

	key := ServingKey{
		Id:     bson.NewObjectId(),
		Key:    c.CacheKey,
		Bucket: c.Bucket,
		Mime:   c.Mime,
		Url: fmt.Sprintf("https://s3.amazonaws.com/%s/%s",
			c.Bucket, path),
	}

	b := s3conn.Bucket(c.Bucket)
	err := b.Put(path, buf, http.DetectContentType(buf), s3.BucketOwnerRead)
	if err != nil {
		return err
	}

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

func ImageData(s3conn *s3.S3, gc groupcache.Context) ([]byte, error) {
	c, ok := gc.(*CacheContext)
	if !ok {
		return nil, errors.New("invalid context")
	}

	// If the image was requested without any size modifier
	if c.Width == 0 {
		var result ServingKey
		data, mime, err := findOriginalImage(&result, s3conn, c)
		if err != nil {
			return nil, err
		}
		c.Mime = mime

		return data, err
	}

	var mime string
	var result ServingKey

	data, mime, err := findResizedImage(&result, s3conn, c)
	if err != nil {
		data, c.Mime, err = findOriginalImage(&result, s3conn, c)
		if err != nil {
			return nil, err
		}
		c.Mime = mime

		// Gifs don't get resized
		if mime == "image/gif" {
			return data, err
		}

		buf, err := Resize(bytes.NewBuffer(data), c)
		if err != nil {
			return nil, err
		}

		err = writeResizedImage(buf, s3conn, c)
		if err != nil {
			return nil, err
		}

		return buf, err
	}

	c.Mime = mime
	return data, err
}
