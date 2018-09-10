package iaas

import (
	"fmt"
)

// IClient represents actions taken against AWS
type IClient interface {
	CheckForWhitelistedIP(ip, securityGroup string) (bool, error)
	DeleteFile(bucket, path string) error
	DeleteVersionedBucket(name string) error
	DeleteVMsInVPC(vpcID string) ([]*string, error)
	DeleteVolumes(volumesToDelete []*string, deleteVolume func(ec2Client IEC2, volumeID *string) error) error
	EnsureBucketExists(name string) error
	EnsureFileExists(bucket, path string, defaultContents []byte) ([]byte, bool, error)
	FindLongestMatchingHostedZone(subdomain string) (string, string, error)
	HasFile(bucket, path string) (bool, error)
	LoadFile(bucket, path string) ([]byte, error)
	WriteFile(bucket, path string, contents []byte) error
	Region() string
	IAAS() string
}

// New returns a new IAAS client for a particular IAAS and region
func New(iaas string, region string) (IClient, error) {
	if iaas == "AWS" {
		return newAWS(region)
	}

	return nil, fmt.Errorf("IAAS not supported: %s", iaas)
}
