package store

import (
	"launchpad.net/goamz/s3"
)

type ImageStore interface {
	Get(string, string) ([]byte, error)
	Put(string, string, []byte, string) error
}

type S3ImageStore struct {
	conn *s3.S3
}

func NewS3Store(conn *s3.S3) *S3ImageStore {
	return &S3ImageStore{conn}
}

func (s *S3ImageStore) Get(bucket, path string) ([]byte, error) {
	return s.conn.Bucket(bucket).Get(path)
}

func (s *S3ImageStore) Put(bucket, path string, data []byte, content string) error {
	return s.conn.Bucket(bucket).Put(path, data, content, s3.BucketOwnerRead)
}
