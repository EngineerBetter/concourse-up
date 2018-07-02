package bosh

import (
	"strconv"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/db"

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

// DirectorBPMReleaseSHA1 is a compile-time varaible set with -ldflags
var DirectorBPMReleaseSHA1 = "COMPILE_TIME_VARIABLE_bosh_directorBPMReleaseSHA1"

// DirectorBPMReleaseURL is a compile-time varaible set with -ldflags
var DirectorBPMReleaseURL = "COMPILE_TIME_VARIABLE_bosh_directorBPMReleaseURL"

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
		DBCACert:                  db.RDSRootCert,
		DBHost:                    metadata.BoshDBAddress.Value,
		DBName:                    conf.RDSDefaultDatabaseName,
		DBPassword:                conf.RDSPassword,
		DBPort:                    dbPort,
		DBUsername:                conf.RDSUsername,
		DirectorBPMReleaseSHA1:    DirectorBPMReleaseSHA1,
		DirectorBPMReleaseURL:     DirectorBPMReleaseURL,
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
	DBCACert                  string
	DBHost                    string
	DBName                    string
	DBPassword                string
	DBPort                    int
	DBUsername                string
	DirectorBPMReleaseSHA1    string
	DirectorBPMReleaseURL     string
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

var awsDirectorManifestTemplate = string(MustAsset("assets/director.yml"))
