package bosh

import (
	"strconv"

	"github.com/EngineerBetter/concourse-up/config"

	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"
)

// DirectorCPIReleaseSHA1 is a compile-time varaible set with -ldflags
var DirectorCPIReleaseSHA1 = "COMPILE_TIME_VARIABLE_bosh_directorCPIReleaseSHA1"

// DirectorCPIReleaseURL is a compile-time varaible set with -ldflags
var DirectorCPIReleaseURL = "COMPILE_TIME_VARIABLE_bosh_directorCPIReleaseURL"

// DirectorCPIReleaseVersion is a compile-time varaible set with -ldflags
var DirectorCPIReleaseVersion = "COMPILE_TIME_VARIABLE_bosh_directorCPIReleaseVersion"

// DirectorReleaseSHA1 is a compile-time varaible set with -ldflags
var DirectorReleaseSHA1 = "COMPILE_TIME_VARIABLE_bosh_directorReleaseSHA1"

// DirectorReleaseURL is a compile-time varaible set with -ldflags
var DirectorReleaseURL = "COMPILE_TIME_VARIABLE_bosh_directorReleaseURL"

// DirectorReleaseVersion is a compile-time varaible set with -ldflags
var DirectorReleaseVersion = "COMPILE_TIME_VARIABLE_bosh_directorReleaseVersion"

// DirectorStemcellSHA1 is a compile-time varaible set with -ldflags
var DirectorStemcellSHA1 = "COMPILE_TIME_VARIABLE_bosh_directorStemcellSHA1"

// DirectorStemcellURL is a compile-time varaible set with -ldflags
var DirectorStemcellURL = "COMPILE_TIME_VARIABLE_bosh_directorStemcellURL"

// DirectorStemcellVersion is a compile-time varaible set with -ldflags
var DirectorStemcellVersion = "COMPILE_TIME_VARIABLE_bosh_directorStemcellVersion"

// GenerateBoshInitManifest generates a manifest for the bosh director on AWS
func generateBoshInitManifest(conf *config.Config, metadata *terraform.Metadata, privateKeyPath string) ([]byte, error) {
	dbPort, err := strconv.Atoi(metadata.BoshDBPort.Value)
	if err != nil {
		return nil, err
	}

	templateParams := awsDirectorManifestParams{
		AWSRegion:                 conf.Region,
		AdminUserName:             conf.DirectorUsername,
		AdminUserPassword:         conf.DirectorPassword,
		AvailabilityZone:          conf.AvailabilityZone,
		BlobstoreBucket:           metadata.BlobstoreBucket.Value,
		BoshAWSAccessKeyID:        metadata.BoshUserAccessKeyID.Value,
		BoshAWSSecretAccessKey:    metadata.BoshSecretAccessKey.Value,
		BoshSecurityGroupID:       metadata.DirectorSecurityGroupID.Value,
		DBHost:                    metadata.BoshDBAddress.Value,
		DBName:                    conf.RDSDefaultDatabaseName,
		DBPassword:                conf.RDSPassword,
		DBPort:                    dbPort,
		DBUsername:                conf.RDSUsername,
		DirectorCACert:            conf.DirectorCACert,
		DirectorCPIReleaseSHA1:    DirectorCPIReleaseSHA1,
		DirectorCPIReleaseURL:     DirectorCPIReleaseURL,
		DirectorCPIReleaseVersion: DirectorCPIReleaseVersion,
		DirectorCert:              conf.DirectorCert,
		DirectorKey:               conf.DirectorKey,
		DirectorReleaseSHA1:       DirectorReleaseSHA1,
		DirectorReleaseURL:        DirectorReleaseURL,
		DirectorReleaseVersion:    DirectorReleaseVersion,
		DirectorSubnetID:          metadata.PublicSubnetID.Value,
		HMUserPassword:            conf.DirectorHMUserPassword,
		KeyPairName:               metadata.DirectorKeyPair.Value,
		MbusPassword:              conf.DirectorMbusPassword,
		NATSPassword:              conf.DirectorNATSPassword,
		PrivateKeyPath:            privateKeyPath,
		PublicIP:                  metadata.DirectorPublicIP.Value,
		RegistryPassword:          conf.DirectorRegistryPassword,
		S3AWSAccessKeyID:          metadata.BlobstoreUserAccessKeyID.Value,
		S3AWSSecretAccessKey:      metadata.BlobstoreSecretAccessKey.Value,
		StemcellSHA1:              DirectorStemcellSHA1,
		StemcellURL:               DirectorStemcellURL,
		StemcellVersion:           DirectorStemcellVersion,
		VMsSecurityGroupID:        metadata.VMsSecurityGroupID.Value,
	}

	return util.RenderTemplate(awsDirectorManifestTemplate, templateParams)
}

type awsDirectorManifestParams struct {
	AWSRegion                 string
	AdminUserName             string
	AdminUserPassword         string
	AvailabilityZone          string
	BlobstoreBucket           string
	BoshAWSAccessKeyID        string
	BoshAWSSecretAccessKey    string
	BoshSecurityGroupID       string
	DBHost                    string
	DBName                    string
	DBPassword                string
	DBPort                    int
	DBUsername                string
	DirectorCACert            string
	DirectorCPIReleaseSHA1    string
	DirectorCPIReleaseURL     string
	DirectorCPIReleaseVersion string
	DirectorCert              string
	DirectorKey               string
	DirectorReleaseSHA1       string
	DirectorReleaseURL        string
	DirectorReleaseVersion    string
	DirectorSubnetID          string
	HMUserPassword            string
	KeyPairName               string
	MbusPassword              string
	NATSPassword              string
	PrivateKeyPath            string
	PublicIP                  string
	RegistryPassword          string
	S3AWSAccessKeyID          string
	S3AWSSecretAccessKey      string
	StemcellSHA1              string
	StemcellURL               string
	StemcellVersion           string
	VMsSecurityGroupID        string
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
    instance_type: t2.small
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
      address: 10.0.0.6
      user: nats
      password: <% .NATSPassword %>
      tls:
        ca: ((nats_server_tls.ca))
        client_ca:
          certificate: ((nats_ca.certificate))
          private_key: ((nats_ca.private_key))
        server:
          certificate: ((nats_server_tls.certificate))
          private_key: ((nats_server_tls.private_key))
        director:
          certificate: ((nats_clients_director_tls.certificate))
          private_key: ((nats_clients_director_tls.private_key))
        health_monitor:
          certificate: ((nats_clients_health_monitor_tls.certificate))
          private_key: ((nats_clients_health_monitor_tls.private_key))

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

variables:
- name: nats_ca
  type: certificate
  options:
    is_ca: true
    common_name: default.nats-ca.bosh-internal

- name: nats_server_tls
  type: certificate
  options:
    ca: nats_ca
    common_name: default.nats.bosh-internal
    alternative_names: [10.0.0.6]
    extended_key_usage:
    - server_auth

- name: nats_clients_director_tls
  type: certificate
  options:
    ca: nats_ca
    common_name: default.director.bosh-internal
    extended_key_usage:
    - client_auth

- name: nats_clients_health_monitor_tls
  type: certificate
  options:
    ca: nats_ca
    common_name: default.hm.bosh-internal
    extended_key_usage:
    - client_auth
`
