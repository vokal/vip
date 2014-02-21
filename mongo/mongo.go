package mongo

import (
	"fmt"
	"github.com/vokalinteractive/vip/fetch"
	"github.com/vokalinteractive/vip/store"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Mongo struct {
	session *mgo.Session
	Db      *mgo.Database
}

type ServingKey struct {
	Key    string `bson:"key"`
	Bucket string `bson:"bucket"`
	Mime   string `bson:"mime"`
}

func Dial(host string) (*Mongo, error) {
	s, err := mgo.Dial(host)
	if err != nil {
		panic(err.Error())
	}

	parsed, err := url.Parse(host)

	var databaseName string
	if err == nil {
		databaseName = parsed.Path[1:]
	} else {
		databaseName = host
	}

	return &Mongo{
		session: s,
		Db:      s.DB(databaseName),
	}, err
}

func (m *Mongo) FindOriginal(storage store.ImageStore, c *fetch.CacheContext) ([]byte, string, error) {
	var result ServingKey
	err := m.Db.C("image_serving_keys").Find(bson.M{
		"key": c.ImageId,
	}).One(&result)

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

func (m *Mongo) FindResized(storage store.ImageStore, c *fetch.CacheContext) ([]byte, string, error) {
	var result ServingKey
	err := m.Db.C("image_serving_keys").Find(bson.M{
		"key": fetch.GetCacheKey(c.Bucket, c.ImageId, c.Width),
	}).One(&result)

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

func (m *Mongo) WriteResized(buf []byte, storage store.ImageStore, c *fetch.CacheContext) error {
	path := fmt.Sprintf("%s/s/%d", c.ImageId, c.Width)

	key := ServingKey{
		Key:    c.CacheKey,
		Bucket: c.Bucket,
		Mime:   c.Mime,
	}

	go func() {
		err := storage.Put(c.Bucket, path, buf, http.DetectContentType(buf))
		if err != nil {
			log.Printf("s3 upload: %s", err.Error())
		}
	}()

	return m.Db.C("image_serving_keys").Insert(key)
}
