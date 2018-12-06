package iaas

import (
	"fmt"
	"strings"
)

// Provider represents actions taken against AWS
type Provider interface {
	Attr(string) (string, error)
	BucketExists(name string) (bool, error)
	CheckForWhitelistedIP(ip, securityGroup string) (bool, error)
	CreateBucket(name string) error
	CreateDatabases(name, username, password string) error
	DeleteFile(bucket, path string) error
	DeleteVersionedBucket(name string) error
	DeleteVMsInDeployment(zone, project, deployment string) error
	DeleteVMsInVPC(vpcID string) ([]string, error)
	DeleteVolumes(volumesToDelete []string, deleteVolume func(ec2Client IEC2, volumeID *string) error) error
	EnsureFileExists(bucket, path string, defaultContents []byte) ([]byte, bool, error)
	FindLongestMatchingHostedZone(subdomain string) (string, string, error)
	HasFile(bucket, path string) (bool, error)
	IAAS() string
	LoadFile(bucket, path string) ([]byte, error)
	Region() string
	WorkerType(string)
	WriteFile(bucket, path string, contents []byte) error
	Zone(string) string
}

// Factory creates a new IaaS provider, defined for testability
type Factory func(iaasName, region string, gops ...GCPOption) (Provider, error)

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
