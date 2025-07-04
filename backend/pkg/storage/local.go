package storage

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	basePath  string
	cdnPrefix string
}

func NewLocalStorage(cfg StorageConfig) *LocalStorage {
	return &LocalStorage{
		basePath:  cfg.LocalPath,
		cdnPrefix: cfg.CDNPrefix,
	}
}

func (s *LocalStorage) Put(key string, file multipart.File) (string, error) {
	path := filepath.Join(s.basePath, key)
	os.MkdirAll(filepath.Dir(path), 0755)
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		return "", err
	}
	return s.GetURL(key)
}

func (s *LocalStorage) GetURL(key string) (string, error) {
	if s.cdnPrefix != "" {
		return s.cdnPrefix + "/" + key, nil
	}
	return "/uploads/" + key, nil
}

func (s *LocalStorage) Delete(key string) error {
	path := filepath.Join(s.basePath, key)
	return os.Remove(path)
}
