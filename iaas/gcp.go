package iaas

import (
	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

type GCPProvider struct {
	ctx     context.Context
	storage *storage.Client
	region  string
	project string
}

type GCPOption func(*GCPProvider) error

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
func (g *GCPProvider) DeleteFile(bucket, path string) error {
	return nil
}
func (g *GCPProvider) DeleteVersionedBucket(name string) error {
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

func (g *GCPProvider) BucketExists(name string) (bool, error) {
	return false, nil
}
func (g *GCPProvider) EnsureFileExists(bucket, path string, defaultContents []byte) ([]byte, bool, error) {
	return nil, false, nil
}
func (g *GCPProvider) FindLongestMatchingHostedZone(subdomain string) (string, string, error) {
	return "", "", nil
}
func (g *GCPProvider) HasFile(bucket, path string) (bool, error) {
	return false, nil
}
func (g *GCPProvider) LoadFile(bucket, path string) ([]byte, error) {
	return nil, nil
}
func (g *GCPProvider) WriteFile(bucket, path string, contents []byte) error {
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
