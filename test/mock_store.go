package test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Store struct {
	store map[string][]byte
}

type MockCloser struct {
	io.Reader
}

func (m MockCloser) Close() error {
	return nil
}

func NewStore() *Store {
	return &Store{
		store: make(map[string][]byte),
	}
}

func (s *Store) GetReader(bucket, path string) (io.ReadCloser, error) {
	data := s.store[fmt.Sprintf("%s|%s", bucket, path)]
	if data == nil {
		return nil, errors.New("item doesn't exist")
	}

	return MockCloser{bytes.NewBuffer(data)}, nil
}

func (s *Store) PutReader(bucket, path string, data io.Reader, length int64, content string) error {
	var buf bytes.Buffer
	buf.ReadFrom(data)
	s.store[fmt.Sprintf("%s|%s", bucket, path)] = buf.Bytes()
	return nil
}

func (s *Store) Put(bucket, path string, data []byte, content string) error {
	s.store[fmt.Sprintf("%s|%s", bucket, path)] = data
	return nil
}

func (s *Store) Head(bucket, path string) (*http.Response, error) {
	return nil, errors.New("")
}
