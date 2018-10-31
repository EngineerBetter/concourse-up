package iaas

import (
	"io/ioutil"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

// GCPProvider is the concrete implementation of GCP Provider
type GCPProvider struct {
	ctx     context.Context
	storage *storage.Client
	region  string
	project string
}

// GCPOption is the signature of the option function
type GCPOption func(*GCPProvider) error

// GCPStorage returns an option function with storage initialised
func GCPStorage() GCPOption {
	return func(c *GCPProvider) error {
		s, err := storage.NewClient(c.ctx)
		if err != nil {
			return err
		}
		c.storage = s
		return nil
	}
}

func newGCP(region, project string, ops ...GCPOption) (Provider, error) {
	ctx := context.Background()

	g := &GCPProvider{ctx, &storage.Client{}, region, project}
	for _, op := range ops {
		if err := op(g); err != nil {
			return nil, err
		}
	}
	return g, nil
}

func (g *GCPProvider) CheckForWhitelistedIP(ip, securityGroup string) (bool, error) {
	return false, nil
}

// DeleteFile deletes a file from GCP bucket
func (g *GCPProvider) DeleteFile(bucket, path string) error {
	o := g.storage.Bucket(bucket).Object(path)

	if err := o.Delete(g.ctx); err != nil {
		return err
	}

	return nil
}

// DeleteVersionedBucket deletes a bucket and its content from GCP
func (g *GCPProvider) DeleteVersionedBucket(name string) error {
	if err := g.storage.Bucket(name).Delete(g.ctx); err != nil {
		return err
	}

	return nil
}

func (g *GCPProvider) DeleteVMsInVPC(vpcID string) ([]*string, error) {
	return nil, nil
}
func (g *GCPProvider) DeleteVolumes(volumesToDelete []*string, deleteVolume func(ec2Client IEC2, volumeID *string) error) error {
	return nil
}

// CreateBucket creates a GCP storage bucket with defaults of the US multi-regional location, and a storage class of Standard Storage
func (g *GCPProvider) CreateBucket(name string) error {
	if err := g.storage.Bucket(name).Create(g.ctx, g.project, nil); err != nil {
		return err
	}
	return nil
}

// BucketExists checks if the named bucket exists
func (g *GCPProvider) BucketExists(name string) (bool, error) {
	it := g.storage.Buckets(g.ctx, g.project)

	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return false, err
		}
		if battrs.Name == name {
			return true, nil
		}
		return false, nil
	}

	return false, nil
}

func (g *GCPProvider) EnsureFileExists(bucket, path string, defaultContents []byte) ([]byte, bool, error) {

	return nil, false, nil
}
func (g *GCPProvider) FindLongestMatchingHostedZone(subdomain string) (string, string, error) {
	return "", "", nil
}

// HasFile returns true if the specified GCP file exists
func (g *GCPProvider) HasFile(bucket, path string) (bool, error) {
	o := g.storage.Bucket(bucket).Object(path)
	_, err := o.Attrs(g.ctx)

	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// LoadFile loads a file from GCP bucket
func (g *GCPProvider) LoadFile(bucket, path string) ([]byte, error) {
	rc, err := g.storage.Bucket(bucket).Object(path).NewReader(g.ctx)

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

// WriteFile writes the specified file to GCP storage
func (g *GCPProvider) WriteFile(bucket, path string, contents []byte) error {
	wc := g.storage.Bucket(bucket).Object(path).NewWriter(g.ctx)
	defer wc.Close()

	if _, err := wc.Write(contents); err != nil {
		return err
	}

	return nil
}

// Region returns the region used by the Provider
func (g *GCPProvider) Region() string {
	return g.region
}

// IAAS returns the name of the Provider
func (g *GCPProvider) IAAS() string {
	return "GCP"
}
