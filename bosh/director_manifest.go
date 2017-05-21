package bosh

import (
	"strconv"

	"github.com/EngineerBetter/concourse-up/config"

	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"
)

var directorStemcellSHA1 = "COMPILE_TIME_VARIABLE_bosh_directorStemcellSHA1"
var directorStemcellURL = "COMPILE_TIME_VARIABLE_bosh_directorStemcellURL"

var directorCPIReleaseSHA1 = "COMPILE_TIME_VARIABLE_bosh_directorCPIReleaseSHA1"
var directorCPIReleaseURL = "COMPILE_TIME_VARIABLE_bosh_directorCPIReleaseURL"

var directorReleaseSHA1 = "COMPILE_TIME_VARIABLE_bosh_directorReleaseSHA1"
var directorReleaseURL = "COMPILE_TIME_VARIABLE_bosh_directorReleaseURL"

// GenerateBoshInitManifest generates a manifest for the bosh director on AWS
func generateBoshInitManifest(conf *config.Config, metadata *terraform.Metadata, privateKeyPath string) ([]byte, error) {
	dbPort, err := strconv.Atoi(metadata.BoshDBPort.Value)
	if err != nil {
		return nil, err
	}

	templateParams := awsDirectorManifestParams{
		AWSRegion:              conf.Region,
		AdminUserName:          conf.DirectorUsername,
		AdminUserPassword:      conf.DirectorPassword,
		AvailabilityZone:       conf.AvailabilityZone,
		BlobstoreBucket:        metadata.BlobstoreBucket.Value,
		BoshAWSAccessKeyID:     metadata.BoshUserAccessKeyID.Value,
		BoshAWSSecretAccessKey: metadata.BoshSecretAccessKey.Value,
		BoshSecurityGroupID:    metadata.DirectorSecurityGroupID.Value,
		DBHost:                 metadata.BoshDBAddress.Value,
		DBPassword:             conf.RDSPassword,
		DBPort:                 dbPort,
		DBUsername:             conf.RDSUsername,
		DBName:                 conf.RDSDefaultDatabaseName,
		DirectorSubnetID:       metadata.DefaultSubnetID.Value,
		HMUserPassword:         conf.DirectorHMUserPassword,
		KeyPairName:            metadata.DirectorKeyPair.Value,
		MbusPassword:           conf.DirectorMbusPassword,
		NATSPassword:           conf.DirectorNATSPassword,
		PrivateKeyPath:         privateKeyPath,
		PublicIP:               metadata.DirectorPublicIP.Value,
		RegistryPassword:       conf.DirectorRegistryPassword,
		S3AWSAccessKeyID:       metadata.BlobstoreUserAccessKeyID.Value,
		S3AWSSecretAccessKey:   metadata.BlobstoreSecretAccessKey.Value,
		StemcellSHA1:           directorStemcellSHA1,
		StemcellURL:            directorStemcellURL,
		DirectorCPIReleaseSHA1: directorCPIReleaseSHA1,
		DirectorCPIReleaseURL:  directorCPIReleaseURL,
		DirectorReleaseSHA1:    directorReleaseSHA1,
		DirectorReleaseURL:     directorReleaseURL,
		VMsSecurityGroupID:     metadata.VMsSecurityGroupID.Value,
		DirectorCACert:         conf.DirectorCACert,
		DirectorCert:           conf.DirectorCert,
		DirectorKey:            conf.DirectorKey,
	}

	return util.RenderTemplate(awsDirectorManifestTemplate, templateParams)
}

type awsDirectorManifestParams struct {
	AWSRegion              string
	AdminUserName          string
	AdminUserPassword      string
	AvailabilityZone       string
	BlobstoreBucket        string
	BoshAWSAccessKeyID     string
	BoshAWSSecretAccessKey string
	BoshSecurityGroupID    string
	DBHost                 string
	DBName                 string
	DBPassword             string
	DBPort                 int
	DBUsername             string
	DirectorCACert         string
	DirectorCPIReleaseSHA1 string
	DirectorCPIReleaseURL  string
	DirectorCert           string
	DirectorKey            string
	DirectorReleaseSHA1    string
	DirectorReleaseURL     string
	DirectorSubnetID       string
	HMUserPassword         string
	KeyPairName            string
	MbusPassword           string
	NATSPassword           string
	PrivateKeyPath         string
	PublicIP               string
	RegistryPassword       string
	S3AWSAccessKeyID       string
	S3AWSSecretAccessKey   string
	StemcellSHA1           string
	StemcellURL            string
	VMsSecurityGroupID     string
}

// Indent is a helper function to indent the field a given number of spaces
func (params awsDirectorManifestParams) Indent(countStr, field string) string {
	return util.Indent(countStr, field)
}

var awsDirectorManifestTemplate = `---
name: bosh
releases:
- name: bosh
  url: <% .DirectorReleaseURL %>
  sha1: <% .DirectorReleaseSHA1 %>
- name: bosh-aws-cpi
  url: <% .DirectorCPIReleaseURL %>
  sha1: <% .DirectorCPIReleaseSHA1 %>

resource_pools:
- name: vms
  network: private
  stemcell:
    url: <% .StemcellURL %>
    sha1: <% .StemcellSHA1 %>
  cloud_properties:
    instance_type: t2.medium
    ephemeral_disk:
      size: 25_000
      type: gp2
    availability_zone: <% .AvailabilityZone %>
  env:
    bosh:
      # c1oudc0w is a default password for vcap user
      password: "$6$4gDD3aV0rdqlrKC$2axHCxGKIObs6tAmMTqYCspcdvQXh3JJcvWOY2WGb4SrdXtnCyNaWlrf3WEqvYR2MYizEGp3kMmbpwBC6jsHt0"

disk_pools:
- name: disks
  disk_size: 20_000
  cloud_properties:
    type: gp2

networks:
- name: private
  type: manual
  subnets:
  - range: 10.0.0.0/24
    gateway: 10.0.0.1
    dns:
    - 10.0.0.2
    cloud_properties:
      subnet: <% .DirectorSubnetID %>
- name: public
  type: vip

jobs:
- name: bosh
  instances: 1
  templates:
  - name: nats
    release: bosh
  - name: director
    release: bosh
  - name: health_monitor
    release: bosh
  - name: registry
    release: bosh
  - name: aws_cpi
    release: bosh-aws-cpi
  resource_pool: vms
  persistent_disk_pool: disks

  networks:
  - name: private
    static_ips: [10.0.0.6]
    default: [dns, gateway]
  - name: public
    static_ips:
    - <% .PublicIP %>

  properties:
    nats:
      address: 127.0.0.1
      user: nats
      password: <% .NATSPassword %>

    postgres: &db
      host: <% .DBHost %>
      user: <% .DBUsername %>
      password: <% .DBPassword %>
      port: <% .DBPort %>
      database: <% .DBName %>
      adapter: postgres

    registry:
      address: 10.0.0.6
      host: 10.0.0.6
      db: *db
      http:
        user: admin
        password: <% .RegistryPassword %>
        port: 25777
      username: admin
      password: <% .RegistryPassword %>
      port: 25777

    blobstore:
      provider: s3
      s3_region: <% .AWSRegion %>
      access_key_id: "<% .S3AWSAccessKeyID %>"
      secret_access_key: "<% .S3AWSSecretAccessKey %>"
      bucket_name: <% .BlobstoreBucket %>

    director:
      address: 127.0.0.1
      name: bosh
      db: *db
      cpi_job: aws_cpi
      max_threads: 10
      user_management:
        provider: local
        local:
          users:
          - name: <% .AdminUserName %>
            password: <% .AdminUserPassword %>
          - name: hm
            password: <% .HMUserPassword %>
      ssl:
        cert: |-
          <% .Indent "10" .DirectorCert %>
        key: |-
          <% .Indent "10" .DirectorKey %>
    hm:
      resurrector_enabled: true
      director_account:
        user: hm
        password: <% .HMUserPassword %>
        ca_cert: |-
          <% .Indent "10" .DirectorCACert %>

    aws: &aws
      access_key_id: "<% .BoshAWSAccessKeyID %>"
      secret_access_key: "<% .BoshAWSSecretAccessKey %>"
      default_key_name: <% .KeyPairName %>
      default_security_groups:
      - <% .BoshSecurityGroupID %>
      - <% .VMsSecurityGroupID %>
      region: <% .AWSRegion %>
    agent:
      mbus: "nats://nats:<% .NATSPassword %>@10.0.0.6:4222"
    ntp: &ntp
    - 0.pool.ntp.org
    - 1.pool.ntp.org

cloud_provider:
  template:
    name: aws_cpi
    release: bosh-aws-cpi
  ssh_tunnel:
    host: <% .PublicIP %>
    port: 22
    user: vcap
    private_key: <% .PrivateKeyPath %>
  mbus: "https://mbus:<% .MbusPassword %>@<% .PublicIP %>:6868"
  properties:
    aws: *aws
    agent:
      mbus: "https://mbus:<% .MbusPassword %>@0.0.0.0:6868"
    blobstore:
      provider: local
      path: /var/vcap/micro_bosh/data/cache
    ntp: *ntp
`
