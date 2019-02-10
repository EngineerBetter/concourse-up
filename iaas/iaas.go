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

type Name int

const (
	Unknown = iota
	AWS
	GCP
)

var names = []string{
	"Unknown",
	"AWS",
	"GCP",
}

func (n Name) String() string {
	return names[n]
}

func Assosiate(name string) (Name, error) {
	name = strings.ToUpper(name)
	for n := len(names) - 1; n > 0; n-- {
		if name == names[n] {
			return Name(n), nil
		}
	}
	return Unknown, fmt.Errorf("cannot map iaas [%s] as any of %+v", name, names[1:])
}

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
	IAAS() Name
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
func New(iaasName Name, region string) (Provider, error) {
	switch iaasName {
	case AWS:
		if region == "" {
			region = "eu-west-1"
		}

		return newAWS(region)
	case GCP:
		if region == "" {
			region = "europe-west1"
		}
		return newGCP(region, GCPStorage())
	}

	return nil, fmt.Errorf("IAAS not supported: [%s]", iaasName)
}
