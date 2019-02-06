package iaas

import (
	"fmt"
	"strings"
)

// Choice is an interface which can help on the abstraction of provider data
// by defining any kind of data mapped against the available providers
type Choice struct {
	AWS interface{}
	GCP interface{}
}

const (
	awsConst = "AWS"
	gcpConst = "GCP"
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
	DBType(name string) string
	IAAS() string
	LoadFile(bucket, path string) ([]byte, error)
	Region() string
	WorkerType(string)
	WriteFile(bucket, path string, contents []byte) error
	Zone(string) string
	Choose(Choice) interface{}
}

// Factory creates a new IaaS provider, defined for testability
type Factory func(iaasName, region string) (Provider, error)

// New returns a new IAAS client for a particular IAAS and region
func New(iaasName, region string) (Provider, error) {
	switch strings.ToUpper(iaasName) {
	case awsConst:
		return newAWS(region)
	case gcpConst:
		return newGCP(region, GCPStorage())
	}

	return nil, fmt.Errorf("IAAS not supported: [%s]", iaasName)
}
