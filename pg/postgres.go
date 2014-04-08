package pg

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/vokalinteractive/vip/fetch"
	"github.com/vokalinteractive/vip/store"
	"log"
	"net/http"
	"strings"
)

type Postgres struct {
	db *sql.DB
}

const (
	tableCreate = `CREATE TABLE vip_serving_keys (
                    id integer primary key, 
                    key varchar(30) NOT NULL, 
                    bucket varchar(30) NOT NULL, 
                    mime varchar(30) NOT NULL)`
	checkCreate = "SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_name='vip_serving_keys')"
	keyQuery    = "SELECT key, bucket, mime FROM vip_serving_keys WHERE key = $1"
	keyInsert   = "INSERT INTO vip_serving_keys (key, bucket, mime) VALUES ($1, $2, $3)"
)

func Dial(conn string) (*Postgres, error) {
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(checkCreate)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		_, err = db.Exec(tableCreate)
	}

	return &Postgres{db}, err
}

func (p *Postgres) FindOriginal(storage store.ImageStore, c *fetch.CacheContext) ([]byte, string, error) {
	var key, bucket, mime string
	err := p.db.QueryRow(keyQuery, c.ImageId).Scan(&key, &bucket, &mime)

	if err == nil {
		data, err := storage.Get(bucket, key)
		if err != nil {
			log.Printf("s3 download: %s", err.Error())
			return nil, "", err
		}

		return data, mime, err
	}

	return nil, "", err
}

func (p *Postgres) FindResized(storage store.ImageStore, c *fetch.CacheContext) ([]byte, string, error) {
	cachekey := fetch.GetCacheKey(c.Bucket, c.ImageId, c.Width)

	var key, bucket, mime string
	err := p.db.QueryRow(keyQuery, cachekey).Scan(&key, &bucket, &mime)

	if err == nil {
		// Strip the bucket out of the cache key
		path := strings.Split(key, c.Bucket+"/")[1]
		data, err := storage.Get(bucket, path)
		if err != nil {
			log.Printf("s3 download: %s", err.Error())
			return nil, "", err
		}

		return data, mime, err
	}

	return nil, "", err
}

func (p *Postgres) WriteResized(buf []byte, storage store.ImageStore, c *fetch.CacheContext) error {
	path := fmt.Sprintf("%s/s/%d", c.ImageId, c.Width)

	go func() {
		err := storage.Put(c.Bucket, path, buf, http.DetectContentType(buf))
		if err != nil {
			log.Printf("s3 upload: %s", err.Error())
		}
	}()

	_, err := p.db.Exec(keyInsert, c.CacheKey, c.Bucket, c.Mime)
	return err
}
