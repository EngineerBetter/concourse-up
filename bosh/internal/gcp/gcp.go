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
	InternalCIDR       string
	InternalGW         string
	InternalIP         string
	DirectorName       string
	Zone               string
	Network            string
	PublicSubnetwork   string
	PrivateSubnetwork  string
	Tags               string
	ProjectID          string
	GcpCredentialsJSON string
	ExternalIP         string
	Preemptible        bool
}

var allOperations = resource.GCPCPIOps + resource.GCPExternalIPOps + resource.GCPDirectorCustomOps

// ConfigureDirectorManifestCPI interpolates all the Environment parameters and
// required release versions into ready to use Director manifest
func (e Environment) ConfigureDirectorManifestCPI(manifest string) (string, error) {
	gcpCreds, err := ioutil.ReadFile(e.GcpCredentialsJSON)
	if err != nil {
		return "", err
	}

	return yaml.Interpolate(manifest, allOperations, map[string]interface{}{
		"internal_cidr":        e.InternalCIDR,
		"internal_gw":          e.InternalGW,
		"internal_ip":          e.InternalIP,
		"director_name":        e.DirectorName,
		"zone":                 e.Zone,
		"network":              e.Network,
		"subnetwork":           e.PublicSubnetwork,
		"private_subnetwork":   e.PrivateSubnetwork,
		"project_id":           e.ProjectID,
		"gcp_credentials_json": string(gcpCreds),
		"external_ip":          e.ExternalIP,
		"preemptible":          false,
	})
}

type gcpCloudConfigParams struct {
	Zone              string
	Preemptible       bool
	PublicSubnetID    string
	PublicSubnetwork  string
	PrivateSubnetwork string
}

// IAASCheck returns the IAAS provider
func (e Environment) IAASCheck() string {
	return "GCP"
}

// ConfigureDirectorCloudConfig inserts values from the environment into the config template passed as argument
func (e Environment) ConfigureDirectorCloudConfig(cloudConfig string) (string, error) {

	templateParams := gcpCloudConfigParams{
		Zone:              e.Zone,
		PublicSubnetwork:  e.PublicSubnetwork,
		PrivateSubnetwork: e.PrivateSubnetwork,
		Preemptible:       e.Preemptible,
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
	return fmt.Sprintf("https://s3.amazonaws.com/bosh-gce-light-stemcells/light-bosh-stemcell-%s-google-kvm-ubuntu-xenial-go_agent.tgz", version), nil
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
