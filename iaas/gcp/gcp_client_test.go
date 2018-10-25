package gcpclient

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

var ctx = context.Background()

// StorageClient is the implementation of the GCP storage client
type StorageClient struct {
	client *storage.Client
}

// IAAS represents actions taken against GCP
type IAAS interface {
	BucketExists(string) (bool, error)
	CreateBucket(string) error
	DeleteBucket(string) error
	DeleteFile(string, string) error
	EnsureFileExists(string, string, string) ([]byte, bool, error)
	HasFile(string string) (bool, error)
	LoadFile(string, string) ([]byte, error)
	WriteFile(string, string) error
}

// NewGCP returns a new GCP client
func NewGCP() (StorageClient, error) {

	client, err := storage.NewClient(ctx)

	if err != nil {
		return StorageClient{}, err
	}

	return StorageClient{client}, err
}

// CreateBucket creates a GCP storage bucket with defaults of the us multi-regional location and a storage class of Standard Storage
func (client *StorageClient) CreateBucket(bucketName string) error {

	if err := client.client.Bucket(bucketName).Create(ctx, "my-project-name", nil); err != nil {
		return err
	}

	return nil
}

// BucketExists checks if the named bucket exists and creates it if it doesn't
func (client *StorageClient) BucketExists(bucketName string) (bool, error) {

	b := client.client.Bucket(bucketName)
	_, err := b.Attrs(ctx)

	if err != nil {
		return false, err
	}

	return true, nil
}

// DeleteBucket deletes a bucket and its content from GCP
func (client *StorageClient) DeleteBucket(bucketName string) error {

	if err := client.client.Bucket(bucketName).Delete(ctx); err != nil {
		return err
	}

	return nil
}

// WriteFile writes the specified file to GCP storage
func (client *StorageClient) WriteFile(bucketName, objectName, filePath string) error {

	f, err := os.Open(filePath)

	if err != nil {
		return err
	}
	defer f.Close()

	wc := client.client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}

// HasFile returns true if the specified GCP file exists
func (client *StorageClient) HasFile(bucketName, objectName string) (bool, error) {

	o := client.client.Bucket(bucketName).Object(objectName)
	attrs, err := o.Attrs(ctx)

	if err != nil {
		return false, err
	}

	if attrs.Name != fmt.Sprintf("Name: %v\n", objectName) {
		return false, errors.New("Specified file does not exist")
	}

	return true, nil
}

// LoadFile loads a file from GCP bucket
func (client *StorageClient) LoadFile(bucketName, objectName string) ([]byte, error) {

	rc, err := client.client.Bucket(bucketName).Object(objectName).NewReader(ctx)

	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// DeleteFile deletes a file from GCP bucket
func (client *StorageClient) DeleteFile(bucketName, objectName string) error {

	o := client.client.Bucket(bucketName).Object(objectName)

	if err := o.Delete(ctx); err != nil {
		return err
	}

	return nil
}

// EnsureFileExists checks for the named file in GCP and creates it if it doesn't exist
func (client *StorageClient) EnsureFileExists(bucketName, objectName, file string) ([]byte, bool, error) {

	o := client.client.Bucket(bucketName).Object(objectName)
	attrs, err := o.Attrs(ctx)

	if err != nil {
		return nil, false, err
	}

	if attrs.Name != fmt.Sprintf("Name: %v\n", objectName) {

		f, err := os.Open(file)

		if err != nil {
			return nil, false, err
		}
		defer f.Close()

		wc := client.client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
		if _, err = io.Copy(wc, f); err != nil {
			return nil, false, err
		}

		if err := wc.Close(); err != nil {
			return nil, false, err
		}
	}

	return nil, true, nil

}
