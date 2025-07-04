package storage

import (
	"errors"
	"mime/multipart"
)

type StorageDriver string

const (
	DriverLocal StorageDriver = "local"
	DriverMinIO StorageDriver = "minio"
	DriverS3    StorageDriver = "s3"
	DriverAzure StorageDriver = "azure"
	DriverGCP   StorageDriver = "gcp"
	DriverOSS   StorageDriver = "oss"
	DriverCOS   StorageDriver = "cos"
	DriverQiniu StorageDriver = "qiniu"
)

type StorageConfig struct {
	Driver StorageDriver
	// Common
	Bucket    string
	Region    string
	Endpoint  string
	AccessKey string
	SecretKey string
	// Local
	LocalPath string
	// CDN/URL prefix
	CDNPrefix string
}

type StorageService interface {
	Put(key string, file multipart.File) (string, error)
	GetURL(key string) (string, error)
	Delete(key string) error
}

func NewStorageService(cfg StorageConfig) (StorageService, error) {
	switch cfg.Driver {
	case DriverLocal:
		return NewLocalStorage(cfg), nil
	case DriverMinIO:
		return NewMinIOStorage(cfg)
	case DriverS3:
		return NewS3Storage(cfg)
	case DriverAzure:
		return NewAzureStorage(cfg)
	case DriverGCP:
		return NewGCPStorage(cfg)
	case DriverOSS:
		return NewOSSStorage(cfg)
	case DriverCOS:
		return NewCOSStorage(cfg)
	case DriverQiniu:
		return NewQiniuStorage(cfg)
	default:
		return nil, errors.New("unsupported storage driver")
	}
}

// Implementations for each driver would go in their own files (local.go, s3.go, etc.)
func NewMinIOStorage(cfg StorageConfig) (StorageService, error) {
	return nil, errors.New("minio storage driver not implemented")
}

func NewS3Storage(cfg StorageConfig) (StorageService, error) {
	return nil, errors.New("s3 storage driver not implemented")
}

func NewAzureStorage(cfg StorageConfig) (StorageService, error) {
	return nil, errors.New("azure storage driver not implemented")
}

func NewGCPStorage(cfg StorageConfig) (StorageService, error) {
	return nil, errors.New("gcp storage driver not implemented")
}

func NewOSSStorage(cfg StorageConfig) (StorageService, error) {
	return nil, errors.New("oss storage driver not implemented")
}
func NewCOSStorage(cfg StorageConfig) (StorageService, error) {
	return nil, errors.New("cos storage driver not implemented")
}
func NewQiniuStorage(cfg StorageConfig) (StorageService, error) {
	return nil, errors.New("qiniu storage driver not implemented")
}
