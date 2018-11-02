package gcp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/EngineerBetter/concourse-up/resource"
	"github.com/EngineerBetter/concourse-up/util"
	"github.com/EngineerBetter/concourse-up/util/yaml"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// Environment holds all the parameters GCP IAAS needs
type Environment struct {
	AccessKeyID           string
	ATCSecurityGroup      string
	BlobstoreBucket       string
	DBCACert              string
	DBHost                string
	DBName                string
	DBPassword            string
	DBPort                string
	DBUsername            string
	DefaultKeyName        string
	DefaultSecurityGroups []string
	ExternalIP            string
	GCPCredentialsJSON    string
	InternalCIDR          string
	InternalGateway       string
	InternalIP            string
	Network               string
	Preemptible           bool
	PrivateKey            string
	PrivateSubnetID       string
	ProjectID             string
	PublicSubnetID        string
	Region                string
	S3AWSAccessKeyID      string
	S3AWSSecretAccessKey  string
	SecretAccessKey       string
	Tags                  []string
	VMSecurityGroup       string
	Zone                  string
	Credentials           string
}

var allOperations = resource.AWSCPIOps + resource.ExternalIPOps

// ConfigureDirectorManifestCPI interpolates all the Environment parameters and
// required release versions into ready to use Director manifest
func (e Environment) ConfigureDirectorManifestCPI(manifest string) (string, error) {
	return yaml.Interpolate(manifest, allOperations, map[string]interface{}{
		"external_ip":          e.ExternalIP,
		"gcp_credentials_json": e.DefaultKeyName,
		"network":              e.Network,
		"project_id":           e.ProjectID,
		"subnetwork":           e.PublicSubnetID,
		"tags":                 e.Tags,
		"zone":                 e.Zone,
	})
}

type gcpCloudConfigParams struct {
	ATCSecurityGroupID string
	Zone               string
	PrivateSubnetID    string
	PublicSubnetID     string
	Preemptible        bool
	VMsSecurityGroupID string
}

// ConfigureDirectorCloudConfig inserts values from the environment into the config template passed as argument
func (e Environment) ConfigureDirectorCloudConfig(cloudConfig string) (string, error) {

	templateParams := gcpCloudConfigParams{
		Zone:            e.Zone,
		PublicSubnetID:  e.PublicSubnetID,
		PrivateSubnetID: e.PrivateSubnetID,
		Preemptible:     e.Preemptible,
	}

	cc, err := util.RenderTemplate(cloudConfig, templateParams)
	if cc == nil {
		return "", err
	}
	return string(cc), err
}

// ConfigureConcourseStemcell returns the stemcell location string for an AWS specific stemcell for the required concourse version
func (e Environment) ConfigureConcourseStemcell(versions string) (string, error) {
	var ops []struct {
		Path  string
		Value json.RawMessage
	}
	err := json.Unmarshal([]byte(versions), &ops)
	if err != nil {
		return "", err
	}
	var version string
	for _, op := range ops {
		if op.Path != "/stemcells/alias=xenial/version" {
			continue
		}
		err := json.Unmarshal(op.Value, &version)
		if err != nil {
			return "", err
		}
	}
	if version == "" {
		return "", errors.New("did not find stemcell version in versions.json")
	}
	return fmt.Sprintf("https://s3.amazonaws.com/bosh-aws-light-stemcells/light-bosh-stemcell-%s-aws-xen-hvm-ubuntu-xenial-go_agent.tgz", version), nil
}

// Store holds the abstraction of a aws storage artifact
type Store struct {
	s3     s3iface.S3API
	bucket string
}

// NewStore returns a reference to a new Store
func NewStore(s3 s3iface.S3API, bucket string) *Store {
	return &Store{
		s3:     s3,
		bucket: bucket,
	}
}

// Get returns the contents of a Store element identified with a key
func (s *Store) Get(key string) ([]byte, error) {
	result, err := s.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == s3.ErrCodeNoSuchKey {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	return ioutil.ReadAll(result.Body)
}

// Set stores the contents of a Store element identified with a key
func (s *Store) Set(key string, value []byte) error {
	_, err := s.s3.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader(value),
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}
