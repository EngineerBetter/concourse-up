package aws

import (
	"bytes"
	"io/ioutil"

	"github.com/EngineerBetter/concourse-up/bosh/internal/resource"
	"github.com/EngineerBetter/concourse-up/bosh/internal/yaml"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// Environment holds all the parameters AWS IAAS needs
type Environment struct {
	InternalCIDR          string
	InternalGateway       string
	InternalIP            string
	AccessKeyID           string
	SecretAccessKey       string
	Region                string
	AZ                    string
	DefaultKeyName        string
	DefaultSecurityGroups []string
	PrivateKey            string
	PublicSubnetID        string
	PrivateSubnetID       string
	ExternalIP            string
	ATCSecurityGroup      string
	VMSecurityGroup       string
	BlobstoreBucket       string
	DBCACert              string
	DBHost                string
	DBName                string
	DBPassword            string
	DBPort                string
	DBUsername            string
	S3AWSAccessKeyID      string
	S3AWSSecretAccessKey  string
}

var allOperations = resource.AWSCPIOps + resource.ExternalIPOps + resource.DirectorCustomOps

// ConfigureDirectorManifestCPI interpolates all the Environment parameters and
// required release versions into ready to use Director manifest
func (e Environment) ConfigureDirectorManifestCPI(manifest string) (string, error) {
	cpiResource := resource.Get(resource.AWSCPI)
	stemcellResource := resource.Get(resource.AWSStemcell)
	return yaml.Interpolate(manifest, allOperations, map[string]interface{}{
		"cpi_url":                  cpiResource.URL,
		"cpi_version":              cpiResource.Version,
		"cpi_sha1":                 cpiResource.SHA1,
		"stemcell_url":             stemcellResource.URL,
		"stemcell_sha1":            stemcellResource.SHA1,
		"internal_cidr":            e.InternalCIDR,
		"internal_gw":              e.InternalGateway,
		"internal_ip":              e.InternalIP,
		"access_key_id":            e.AccessKeyID,
		"secret_access_key":        e.SecretAccessKey,
		"region":                   e.Region,
		"az":                       e.AZ,
		"default_key_name":         e.DefaultKeyName,
		"default_security_groups":  e.DefaultSecurityGroups,
		"private_key":              e.PrivateKey,
		"subnet_id":                e.PublicSubnetID,
		"external_ip":              e.ExternalIP,
		"blobstore_bucket":         e.BlobstoreBucket,
		"db_ca_cert":               e.DBCACert,
		"db_host":                  e.DBHost,
		"db_name":                  e.DBName,
		"db_password":              e.DBPassword,
		"db_port":                  e.DBPort,
		"db_username":              e.DBUsername,
		"s3_aws_access_key_id":     e.S3AWSAccessKeyID,
		"s3_aws_secret_access_key": e.S3AWSSecretAccessKey,
	})
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
