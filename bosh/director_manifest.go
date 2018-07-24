package bosh

import (
	"encoding/json"
	"strconv"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/db"

	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"
)

var versionFile = MustAsset("../resources/director-versions.json")

// GenerateBoshInitManifest generates a manifest for the bosh director on AWS
func generateBoshInitManifest(conf *config.Config, metadata *terraform.Metadata, privateKeyPath string) ([]byte, error) {
	dbPort, err := strconv.Atoi(metadata.BoshDBPort.Value)
	if err != nil {
		return nil, err
	}

	var x map[string]map[string]string
	err = json.Unmarshal(versionFile, &x)

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
		DirectorBPMReleaseSHA1:    x["bpm"]["sha1"],
		DirectorBPMReleaseURL:     x["bpm"]["url"],
		DirectorCACert:            conf.DirectorCACert,
		DirectorCPIReleaseSHA1:    x["cpi"]["sha1"],
		DirectorCPIReleaseURL:     x["cpi"]["url"],
		DirectorCPIReleaseVersion: x["cpi"]["version"],
		DirectorCert:              conf.DirectorCert,
		DirectorKey:               conf.DirectorKey,
		DirectorReleaseSHA1:       x["bosh"]["sha1"],
		DirectorReleaseURL:        x["bosh"]["url"],
		DirectorReleaseVersion:    x["bosh"]["version"],
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
		StemcellSHA1:              x["stemcell"]["sha1"],
		StemcellURL:               x["stemcell"]["url"],
		StemcellVersion:           x["stemcell"]["version"],
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
