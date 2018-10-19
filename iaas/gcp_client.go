package iaas

import "cloud.google.com/go/storage"
import "golang.org/x/net/context"

var ctx = context.Background()

// IAAS represents actions taken against GCP
type IAAS interface {
	CreateBucket(string) error
}

// StorageClient is the implementation of the GCP storage client
type StorageClient struct {
	client *storage.Client
}

// NewGCP returns a new GCP client
func NewGCP() (StorageClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return StorageClient{}, err
	}
	return StorageClient{client}, err
}

// CreateBucket creates a GCP storage bucket
func (client *StorageClient) CreateBucket(name string) error {
	if err := client.client.Bucket(name).Create(ctx, "my-project", nil); err != nil {
		return err
	}
	return nil
}
