package storage

import (
	"context"
	"fmt"
	"time"

	minio "github.com/minio/minio-go/v7"
)

const storageOperationTimeout = 3

type ObjectStorage struct {
	bucketName string
	cache      *DiskStorage
	client     *minio.Client
}

func NewObjectStorage(client *minio.Client, bucket string, cache *DiskStorage) *ObjectStorage {
	return &ObjectStorage{
		bucketName: bucket,
		client:     client,
		cache:      cache,
	}
}

func (o *ObjectStorage) GetStorageDirectory() string {
	return o.bucketName
}

func (o *ObjectStorage) GetFile(objectName string) (contents []byte, err error) {
	if contents, err = o.cache.GetFile(objectName); err == nil {
		return contents, nil
	}

	err = o.getAndCacheFile(objectName)
	if err != nil {
		return nil, err
	}

	return o.cache.GetFile(objectName)
}

func (o *ObjectStorage) PutFile(objectName string, contents []byte) (err error) {
	err = o.cache.PutFile(objectName, contents)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.TODO(), storageOperationTimeout*time.Second)
	defer cancel()

	// Create object in a bucket with contents from local file
	_, err = o.client.FPutObject(
		ctx,
		o.bucketName,
		objectName,
		fmt.Sprintf("%s/%s", o.cache.GetStorageDirectory(), objectName), // filePath
		minio.PutObjectOptions{},
	)

	return err
}

func (o *ObjectStorage) getAndCacheFile(fileName string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), storageOperationTimeout*time.Second)
	defer cancel()

	// Download contents of object to local file
	return o.client.FGetObject(
		ctx,
		o.bucketName,
		fileName, // objectName
		fmt.Sprintf("%s/%s", o.cache.GetStorageDirectory(), fileName), // filePath
		minio.GetObjectOptions{},
	)
}
