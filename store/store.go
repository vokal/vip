package store

import (
	"io"
	"net/http"

	"github.com/mitchellh/goamz/s3"
)

type ImageStore interface {
	GetReader(string, string) (io.ReadCloser, error)
	PutReader(string, string, io.Reader, int64, string) error
	Put(string, string, []byte, string) error
	Head(string, string) (*http.Response, error)
}

type S3ImageStore struct {
	conn *s3.S3
}

func NewS3Store(conn *s3.S3) *S3ImageStore {
	return &S3ImageStore{conn}
}

func (s *S3ImageStore) GetReader(bucket, path string) (io.ReadCloser, error) {
	return s.conn.Bucket(bucket).GetReader(path)
}

func (s *S3ImageStore) PutReader(bucket, path string, data io.Reader, length int64, content string) error {
	return s.conn.Bucket(bucket).PutReader(path, data, length, content, s3.BucketOwnerRead)
}

func (s *S3ImageStore) Put(bucket, path string, data []byte, content string) error {
	return s.conn.Bucket(bucket).Put(path, data, content, s3.BucketOwnerRead)
}

func (s *S3ImageStore) Head(bucket, path string) (*http.Response, error) {
	return s.conn.Bucket(bucket).Head(path)
}
