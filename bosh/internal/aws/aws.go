package aws

import (
	"bytes"
	"io/ioutil"

	"github.com/EngineerBetter/concourse-up/bosh/internal/cli"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type Config struct {
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
	DBPort                int
	DBUsername            string
	S3AWSAccessKeyID      string
	S3AWSSecretAccessKey  string
}

func (c *Config) Operations() []string {
	return []string{cpiOperation, exernalIPOperation, concourseSpecifics}
}

func (c *Config) Vars() map[string]interface{} {
	return map[string]interface{}{
		"director_name":            "bosh",
		"internal_cidr":            c.InternalCIDR,
		"internal_gw":              c.InternalGateway,
		"internal_ip":              c.InternalIP,
		"access_key_id":            c.AccessKeyID,
		"secret_access_key":        c.SecretAccessKey,
		"region":                   c.Region,
		"az":                       c.AZ,
		"default_key_name":         c.DefaultKeyName,
		"default_security_groups":  c.DefaultSecurityGroups,
		"private_key":              c.PrivateKey,
		"subnet_id":                c.PublicSubnetID,
		"external_ip":              c.ExternalIP,
		"blobstore_bucket":         c.BlobstoreBucket,
		"db_ca_cert":               c.DBCACert,
		"db_host":                  c.DBHost,
		"db_name":                  c.DBName,
		"db_password":              c.DBPassword,
		"db_port":                  c.DBPort,
		"db_username":              c.DBUsername,
		"s3_aws_access_key_id":     c.S3AWSAccessKeyID,
		"s3_aws_secret_access_key": c.S3AWSSecretAccessKey,
	}
}

func (c *Config) Address() string {
	return c.ExternalIP
}

func (c *Config) CloudConfig() (string, error) {
	return cli.Interpolate(cloudConfig, nil, map[string]interface{}{
		"az":                 c.AZ,
		"vm_security_group":  c.VMSecurityGroup,
		"public_subnet_id":   c.PublicSubnetID,
		"private_subnet_id":  c.PrivateSubnetID,
		"atc_security_group": c.ATCSecurityGroup,
	})
}

type Store struct {
	s3     s3iface.S3API
	bucket string
}

func NewStore(s3 s3iface.S3API, bucket string) *Store {
	return &Store{
		s3:     s3,
		bucket: bucket,
	}
}

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

func (s *Store) Set(key string, value []byte) error {
	_, err := s.s3.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader(value),
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

const cpiOperation = `---
- type: replace
  path: /releases/-
  value:
    name: bosh-aws-cpi
    version: "72"
    url: https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?v=72
    sha1: b7999e95115bb691a630c07f0439ef5b577884c2

- type: replace
  path: /resource_pools/name=vms/stemcell?
  value:
    url: https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-xenial-go_agent?v=97.12
    sha1: 41014fc439770a6f403dfaf6963a03400d47149d

# Configure AWS sizes
- type: replace
  path: /resource_pools/name=vms/cloud_properties?
  value:
    instance_type: m4.xlarge
    ephemeral_disk:
      type: gp2
      size: 25_000
    availability_zone: ((az))

- type: replace
  path: /disk_pools/name=disks/cloud_properties?
  value:
    type: gp2

- type: replace
  path: /networks/name=default/subnets/0/cloud_properties?
  value:
    subnet: ((subnet_id))

# Enable registry job
- type: replace
  path: /instance_groups/name=bosh/jobs/-
  value:
    name: registry
    release: bosh

- type: replace
  path: /instance_groups/name=bosh/properties/registry?
  value:
    address: ((internal_ip))
    host: ((internal_ip))
    db: # todo remove
      host: 127.0.0.1
      user: postgres
      password: ((postgres_password))
      database: bosh
      adapter: postgres
    http:
      user: registry
      password: ((registry_password))
      port: 25777
    username: registry
    password: ((registry_password))
    port: 25777

# Add CPI job
- type: replace
  path: /instance_groups/name=bosh/jobs/-
  value: &cpi_job
    name: aws_cpi
    release: bosh-aws-cpi

- type: replace
  path: /instance_groups/name=bosh/properties/director/cpi_job?
  value: aws_cpi

- type: replace
  path: /cloud_provider/template?
  value: *cpi_job

- type: replace
  path: /instance_groups/name=bosh/properties/aws?
  value: &aws
    access_key_id: ((access_key_id))
    secret_access_key: ((secret_access_key))
    default_key_name: ((default_key_name))
    default_security_groups: ((default_security_groups))
    region: ((region))

- type: replace
  path: /cloud_provider/ssh_tunnel?
  value:
    host: ((internal_ip))
    port: 22
    user: vcap
    private_key: ((private_key))

- type: replace
  path: /cloud_provider/properties/aws?
  value: *aws

- type: replace
  path: /variables/-
  value:
    name: registry_password
    type: password
`

const exernalIPOperation = `---
- type: replace
  path: /networks/-
  value:
    name: public
    type: vip

- type: replace
  path: /instance_groups/name=bosh/networks/0/default?
  value: [dns, gateway]

- type: replace
  path: /instance_groups/name=bosh/networks/-
  value:
    name: public
    static_ips: [((external_ip))]

- type: replace
  path: /instance_groups/name=bosh/properties/director/default_ssh_options?/gateway_host
  value: ((external_ip))

# todo should not access non-defined vars
- type: replace
  path: /cloud_provider/mbus
  value: https://mbus:((mbus_bootstrap_password))@((external_ip)):6868

- type: replace
  path: /cloud_provider/ssh_tunnel/host
  value: ((external_ip))

- type: replace
  path: /variables/name=mbus_bootstrap_ssl/options/alternative_names/-
  value: ((external_ip))

- type: replace
  path: /variables/name=director_ssl/options/alternative_names/-
  value: ((external_ip))`

const concourseSpecifics = `---
- type: replace
  path: /disk_pools/name=disks/disk_size
  value: 20000

- type: replace
  path: /instance_groups/name=bosh/properties/blobstore
  value:
    access_key_id: ((s3_aws_access_key_id))
    bucket_name: ((blobstore_bucket))
    provider: s3
    s3_region: ((region))
    secret_access_key: ((s3_aws_secret_access_key))

- type: replace
  path: /instance_groups/name=bosh/properties/director/db
  value:
    adapter: postgres
    database: ((db_name))
    host: ((db_host))
    password: ((db_password))
    port: ((db_port))
    user: ((db_username))

- type: replace
  path: /instance_groups/name=bosh/properties/director/max_threads?
  value: 10

- type: replace
  path: /instance_groups/name=bosh/properties/director/trusted_certs?
  value: ((db_ca_cert))

- type: replace
  path: /instance_groups/name=bosh/properties/postgres
  value:
    adapter: postgres
    database: ((db_name))
    host: ((db_host))
    password: ((db_password))
    port: ((db_port))
    user: ((db_username))

- type: replace
  path: /instance_groups/name=bosh/properties/registry/db
  value:
    adapter: postgres
    database: ((db_name))
    host: ((db_host))
    password: ((db_password))
    port: ((db_port))
    user: ((db_username))

- type: replace
  path: /instance_groups/name=bosh/properties/registry/http/user
  value: admin

- type: replace
  path: /instance_groups/name=bosh/properties/registry/username
  value: admin

- type: replace
  path: /resource_pools/name=vms/cloud_properties/instance_type
  value: t2.small

- type: replace
  path: /resource_pools/name=vms/env/bosh/password
  value: "$6$4gDD3aV0rdqlrKC$2axHCxGKIObs6tAmMTqYCspcdvQXh3JJcvWOY2WGb4SrdXtnCyNaWlrf3WEqvYR2MYizEGp3kMmbpwBC6jsHt0"

- type: remove
  path: /instance_groups/name=bosh/properties/agent/env

- type: remove
  path: /variables/name=blobstore_ca

- type: remove
  path: /variables/name=blobstore_server_tls

- type: remove
  path: /instance_groups/name=bosh/jobs/name=postgres-9.4

- type: remove
  path: /instance_groups/name=bosh/jobs/name=blobstore

- type: remove
  path: /cloud_provider/cert

- type: remove
  path: /resource_pools/name=vms/env/bosh/mbus

- type: remove
  path: /variables/name=mbus_bootstrap_ssl

- type: remove
  path: /instance_groups/name=bosh/properties/director/workers

`

const cloudConfig = `---
azs:
- name: z1
  cloud_properties:
    availability_zone: ((az))

vm_types:
- name: concourse-web-small
  cloud_properties:
    instance_type: t2.small
    ephemeral_disk:
      size: 20_000
      type: gp2
      encrypted: true
    security_groups: [((vm_security_group))]

- name: concourse-web-medium
  cloud_properties:
    instance_type: t2.medium
    ephemeral_disk:
      size: 20_000
      type: gp2
      encrypted: true
    security_groups: [((vm_security_group))]

- name: concourse-web-large
  cloud_properties:
    instance_type: t2.large
    ephemeral_disk:
      size: 20_000
      type: gp2
      encrypted: true
    security_groups: [((vm_security_group))]

- name: concourse-web-xlarge
  cloud_properties:
    instance_type: t2.xlarge
    ephemeral_disk:
      size: 20_000
      type: gp2
      encrypted: true
    security_groups: [((vm_security_group))]

- name: concourse-web-2xlarge
  cloud_properties:
    instance_type: t2.2xlarge
    ephemeral_disk:
      size: 20_000
      type: gp2
      encrypted: true
    security_groups: [((vm_security_group))]

- name: concourse-medium
  cloud_properties:
    instance_type: t2.medium
    ephemeral_disk:
      size: 200_000
      type: gp2
      encrypted: true
    security_groups: [((vm_security_group))]

- name: concourse-large
  cloud_properties:
    instance_type: m4.large
    spot_bid_price: 0.13 # on-demand price: 0.111
    spot_ondemand_fallback: true
    ephemeral_disk:
      size: 200_000
      type: gp2
      encrypted: true
    security_groups: [((vm_security_group))]

- name: concourse-xlarge
  cloud_properties:
    instance_type: m4.xlarge
    spot_bid_price: 0.27 # on-demand price: 0.222
    spot_ondemand_fallback: true
    ephemeral_disk:
      size: 200_000
      type: gp2
      encrypted: true
    security_groups: [((vm_security_group))]

- name: concourse-2xlarge
  cloud_properties:
    instance_type: m4.2xlarge
    spot_bid_price: 0.53 # on-demand price: 0.444
    spot_ondemand_fallback: true
    ephemeral_disk:
      size: 200_000
      type: gp2
      encrypted: true
    security_groups: [((vm_security_group))]

- name: concourse-4xlarge
  cloud_properties:
    instance_type: m4.4xlarge
    spot_bid_price: 1.07 # on-demand price: 0.888
    spot_ondemand_fallback: true
    ephemeral_disk:
      size: 200_000
      type: gp2
      encrypted: true
    security_groups: [((vm_security_group))]

- name: concourse-10xlarge
  cloud_properties:
    instance_type: m4.10xlarge
    spot_bid_price: 2.67 # on-demand price: 2.22
    spot_ondemand_fallback: true
    ephemeral_disk:
      size: 200_000
      type: gp2
      encrypted: true
    security_groups: [((vm_security_group))]

- name: concourse-16xlarge
  cloud_properties:
    instance_type: m4.16xlarge
    spot_bid_price: 4.26 # on-demand price: 3.55
    spot_ondemand_fallback: true
    ephemeral_disk:
      size: 200_000
      type: gp2
      encrypted: true
    security_groups: [((vm_security_group))]

- name: compilation
  cloud_properties:
    instance_type: m4.large
    spot_bid_price: 0.13 # on-demand price: 0.111
    spot_ondemand_fallback: true
    ephemeral_disk:
      size: 5_000
      type: gp2
      encrypted: true
    security_groups: [((vm_security_group))]

disk_types:
- name: default
  disk_size: 50_000
  cloud_properties:
    type: gp2
    encrypted: true
- name: large
  disk_size: 200_000
  cloud_properties:
    type: gp2
    encrypted: true

networks:
- name: public
  type: manual
  subnets:
  - range: 10.0.0.0/24
    gateway: 10.0.0.1
    dns:
    - 10.0.0.2
    az: z1
    static:
    - 10.0.0.6
    - 10.0.0.7
    reserved:
    - 10.0.0.1-10.0.0.5
    cloud_properties:
      subnet: ((public_subnet_id))
- name: private
  type: manual
  subnets:
  - range: 10.0.1.0/24
    gateway: 10.0.1.1
    dns:
    - 10.0.0.2
    az: z1
    reserved:
    - 10.0.1.1-10.0.1.5
    cloud_properties:
      subnet: ((private_subnet_id))

- name: vip
  type: vip

vm_extensions:
- name: atc
  cloud_properties:
    security_groups:
    - ((vm_security_group))
    - ((atc_security_group))

compilation:
  workers: 5
  reuse_compilation_vms: true
  az: z1
  vm_type: compilation
  network: private`
