package bosh

import (
	"github.com/EngineerBetter/concourse-up/bosh/internal/aws"
	"github.com/EngineerBetter/concourse-up/bosh/internal/boshenv"
	"github.com/EngineerBetter/concourse-up/db"
)

// Delete deletes a bosh director
func (client *AWSClient) Delete(stateFileBytes []byte) ([]byte, error) {
	if err := client.director.RunAuthenticatedCommand(
		client.stdout,
		client.stderr,
		false,
		"--deployment",
		concourseDeploymentName,
		"delete-deployment",
		"--force",
	); err != nil {
		return nil, err
	}

	//TODO(px): pull up this so that we use aws.Store
	store := temporaryStore{
		"state.json": stateFileBytes,
	}
	bosh, err := boshenv.New()
	if err != nil {
		return store["state.json"], err
	}

	boshUserAccessKeyID, err := client.outputs.Get("BoshUserAccessKeyID")
	if err != nil {
		return store["state.json"], err
	}
	boshSecretAccessKey, err := client.outputs.Get("BoshSecretAccessKey")
	if err != nil {
		return store["state.json"], err
	}
	publicSubnetID, err := client.outputs.Get("PublicSubnetID")
	if err != nil {
		return store["state.json"], err
	}
	privateSubnetID, err := client.outputs.Get("PrivateSubnetID")
	if err != nil {
		return store["state.json"], err
	}
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return store["state.json"], err
	}
	atcSecurityGroupID, err := client.outputs.Get("ATCSecurityGroupID")
	if err != nil {
		return store["state.json"], err
	}
	vmSecurityGroupID, err := client.outputs.Get("VMsSecurityGroupID")
	if err != nil {
		return store["state.json"], err
	}
	blobstoreBucket, err := client.outputs.Get("BlobstoreBucket")
	if err != nil {
		return store["state.json"], err
	}
	boshDBAddress, err := client.outputs.Get("BoshDBAddress")
	if err != nil {
		return store["state.json"], err
	}
	boshDbPort, err := client.outputs.Get("BoshDBPort")
	if err != nil {
		return store["state.json"], err
	}
	blobstoreUserAccessKeyID, err := client.outputs.Get("BlobstoreUserAccessKeyID")
	if err != nil {
		return store["state.json"], err
	}
	blobstoreSecretAccessKey, err := client.outputs.Get("BlobstoreSecretAccessKey")
	if err != nil {
		return store["state.json"], err
	}
	directorKeyPair, err := client.outputs.Get("DirectorKeyPair")
	if err != nil {
		return store["state.json"], err
	}
	directorSecurityGroup, err := client.outputs.Get("DirectorSecurityGroupID")
	if err != nil {
		return store["state.json"], err
	}

	err = bosh.DeleteEnv(store, aws.Environment{
		InternalCIDR:    "10.0.0.0/24",
		InternalGateway: "10.0.0.1",
		InternalIP:      "10.0.0.6",
		AccessKeyID:     boshUserAccessKeyID,
		SecretAccessKey: boshSecretAccessKey,
		Region:          client.config.Region,
		AZ:              client.config.AvailabilityZone,
		DefaultKeyName:  directorKeyPair,
		DefaultSecurityGroups: []string{
			directorSecurityGroup,
			vmSecurityGroupID,
		},
		PrivateKey:           client.config.PrivateKey,
		PublicSubnetID:       publicSubnetID,
		PrivateSubnetID:      privateSubnetID,
		ExternalIP:           directorPublicIP,
		ATCSecurityGroup:     atcSecurityGroupID,
		VMSecurityGroup:      vmSecurityGroupID,
		BlobstoreBucket:      blobstoreBucket,
		DBCACert:             db.RDSRootCert,
		DBHost:               boshDBAddress,
		DBName:               client.config.RDSDefaultDatabaseName,
		DBPassword:           client.config.RDSPassword,
		DBPort:               boshDbPort,
		DBUsername:           client.config.RDSUsername,
		S3AWSAccessKeyID:     blobstoreUserAccessKeyID,
		S3AWSSecretAccessKey: blobstoreSecretAccessKey,
		Spot:                 client.config.Spot,
	}, client.config.DirectorPassword, client.config.DirectorCert, client.config.DirectorKey, client.config.DirectorCACert, nil)
	return store["state.json"], err
}
