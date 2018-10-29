package gcpclient

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

// GCP authorisation requires that the environment variable $GOOGLE_APPLICATION_CREDENTIALS be set

var ctx = context.Background()

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

// CreateBucket creates a GCP storage bucket with defaults of the US multi-regional location, and a storage class of Standard Storage
func CreateBucket(bucketName, projectName string) error {

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Bucket(bucketName).Create(ctx, projectName, nil); err != nil {
		return err
	}

	return nil
}

// BucketExists checks if the named bucket exists and creates it if it doesn't
func BucketExists(bucketName, projectName string) (bool, error) {

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	var buckets []string
	it := client.Buckets(ctx, projectName)

	for {
		battrs, err1 := it.Next()
		if err1 == iterator.Done {
			break
		}
		if err != nil {
			return false, err
		}
		buckets = append(buckets, battrs.Name)
		for _, vv := range buckets {
			if vv == bucketName {
				return true, nil
			}
		}

		return false, err
	}

	return false, err
}

// DeleteBucket deletes a bucket and its content from GCP
func DeleteBucket(bucketName string) error {

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Bucket(bucketName).Delete(ctx); err != nil {
		return err
	}

	return nil
}

// WriteFile writes the specified file to GCP storage
func WriteFile(bucketName, objectName, filePath string) error {

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	wc := client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}

// HasFile returns true if the specified GCP file exists
func HasFile(bucketName, objectName string) (bool, error) {

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	o := client.Bucket(bucketName).Object(objectName)
	attrs, err := o.Attrs(ctx)

	if err != nil {
		return false, err
	}

	if attrs.Name != objectName {
		return false, errors.New("Specified file does not exist")
	}

	return true, nil
}

// LoadFile loads a file from GCP bucket
func LoadFile(bucketName, objectName string) ([]byte, error) {

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	rc, err := client.Bucket(bucketName).Object(objectName).NewReader(ctx)

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
func DeleteFile(bucketName, objectName string) error {

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	o := client.Bucket(bucketName).Object(objectName)

	if err := o.Delete(ctx); err != nil {
		return err
	}

	return nil
}
