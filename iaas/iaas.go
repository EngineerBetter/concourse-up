package iaas

import (
	"fmt"
	"strings"
)

// Provider represents actions taken against AWS
type Provider interface {
	CheckForWhitelistedIP(ip, securityGroup string) (bool, error)
	DeleteFile(bucket, path string) error
	DeleteVersionedBucket(name string) error
	DeleteVMsInVPC(vpcID string) ([]string, error)
	DeleteVMsInDeployment(zone, project, deployment string) error
	DeleteVolumes(volumesToDelete []string, deleteVolume func(ec2Client IEC2, volumeID *string) error) error
	CreateBucket(name string) error
	BucketExists(name string) (bool, error)
	EnsureFileExists(bucket, path string, defaultContents []byte) ([]byte, bool, error)
	FindLongestMatchingHostedZone(subdomain string) (string, string, error)
	HasFile(bucket, path string) (bool, error)
	LoadFile(bucket, path string) ([]byte, error)
	WriteFile(bucket, path string, contents []byte) error
	Region() string
	IAAS() string
	Attr(string) (string, error)
	Zone(string) string
	WorkerType(string)
}

// New returns a new IAAS client for a particular IAAS and region
func New(iaasName, region string, gops ...GCPOption) (Provider, error) {
	switch strings.ToUpper(iaasName) {
	case "AWS":
		return newAWS(region)
	case "GCP":
		if len(gops) == 0 {
			gops = append(gops, GCPStorage())
		}
		return newGCP(region, gops...)
	}

	return nil, fmt.Errorf("IAAS not supported: %s", iaasName)
}
