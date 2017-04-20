package bosh

import (
	"strconv"

	"bitbucket.org/engineerbetter/concourse-up/config"

	"bitbucket.org/engineerbetter/concourse-up/terraform"
	"bitbucket.org/engineerbetter/concourse-up/util"
)

const stemcellSHA1 = "f8f1b0f31135d9bdd527a67de1ebbdfed332f3ed"
const stemcellURL = "https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent?v=3312.23"

const boshCPIReleaseSHA1 = "239fc7797d280f140fc03009fb39060107ff0ee1"
const boshCPIReleaseURL = "https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?v=63"

const boshReleaseSHA1 = "4da9cedbcc8fbf11378ef439fb89de08300ad091"
const boshReleaseURL = "https://bosh.io/d/github.com/cloudfoundry/bosh?v=261.4"

// GenerateAWSDirectorManifest generates a manifest for the bosh director on AWS
func GenerateAWSDirectorManifest(conf *config.Config, privateKeyPath string, metadata *terraform.Metadata) ([]byte, error) {
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
		DBPassword:             metadata.BoshDBPassword.Value,
		DBPort:                 dbPort,
		DBUsername:             metadata.BoshDBUsername.Value,
		DBName:                 conf.RDSDefaultDatabaseName,
		DirectorSubnetID:       metadata.DirectorSubnetID.Value,
		HMUserPassword:         conf.DirectorHMUserPassword,
		KeyPairName:            metadata.DirectorKeyPair.Value,
		MbusPassword:           conf.DirectorMbusPassword,
		NATSPassword:           conf.DirectorNATSPassword,
		PrivateKeyPath:         privateKeyPath,
		PublicIP:               metadata.DirectorPublicIP.Value,
		RegistryPassword:       conf.DirectorRegistryPassword,
		S3AWSAccessKeyID:       metadata.BlobstoreUserAccessKeyID.Value,
		S3AWSSecretAccessKey:   metadata.BlobstoreSecretAccessKey.Value,
		StemcellSHA1:           stemcellSHA1,
		StemcellURL:            stemcellURL,
		BoshAWSCPIReleaseSHA1:  boshCPIReleaseSHA1,
		BoshAWSCPIReleaseURL:   boshCPIReleaseURL,
		BoshReleaseSHA1:        boshReleaseSHA1,
		BoshReleaseURL:         boshReleaseURL,
		VMsSecurityGroupID:     metadata.VMsSecurityGroupID.Value,
	}

	return util.RenderTemplate(awsDirectorManifestTemplate, templateParams)
}

type awsDirectorManifestParams struct {
	AdminUserName          string
	AdminUserPassword      string
	AvailabilityZone       string
	AWSRegion              string
	BlobstoreBucket        string
	BoshAWSAccessKeyID     string
	BoshAWSCPIReleaseSHA1  string
	BoshAWSCPIReleaseURL   string
	BoshAWSSecretAccessKey string
	BoshReleaseSHA1        string
	BoshReleaseURL         string
	BoshSecurityGroupID    string
	DBHost                 string
	DBPassword             string
	DBPort                 int
	DBUsername             string
	DBName                 string
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

var awsDirectorManifestTemplate = `---
name: bosh
releases:
- name: bosh
  url: <% .BoshReleaseURL %>
  sha1: <% .BoshReleaseSHA1 %>
- name: bosh-aws-cpi
  url: <% .BoshAWSCPIReleaseURL %>
  sha1: <% .BoshAWSCPIReleaseSHA1 %>

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
    hm:
      director_account:
        user: hm
        password: <% .HMUserPassword %>
      resurrector_enabled: true
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
