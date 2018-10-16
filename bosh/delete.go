package bosh

import (
	"github.com/EngineerBetter/concourse-up/bosh/internal/aws"
	"github.com/EngineerBetter/concourse-up/bosh/internal/boshenv"
	"github.com/EngineerBetter/concourse-up/db"
)

// Delete deletes a bosh director
func (client *Client) Delete(stateFileBytes []byte) ([]byte, error) {
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

	boshUserAccessKeyID, err := client.metadata.Get("BoshUserAccessKeyID")
	if err != nil {
		return store["state.json"], err
	}
	boshSecretAccessKey, err := client.metadata.Get("BoshSecretAccessKey")
	if err != nil {
		return store["state.json"], err
	}
	publicSubnetID, err := client.metadata.Get("PublicSubnetID")
	if err != nil {
		return store["state.json"], err
	}
	privateSubnetID, err := client.metadata.Get("PrivateSubnetID")
	if err != nil {
		return store["state.json"], err
	}
	directorPublicIP, err := client.metadata.Get("DirectorPublicIP")
	if err != nil {
		return store["state.json"], err
	}
	atcSecurityGroupID, err := client.metadata.Get("ATCSecurityGroupID")
	if err != nil {
		return store["state.json"], err
	}
	vmSecurityGroupID, err := client.metadata.Get("VMsSecurityGroupID")
	if err != nil {
		return store["state.json"], err
	}
	blobstoreBucket, err := client.metadata.Get("BlobstoreBucket")
	if err != nil {
		return store["state.json"], err
	}
	boshDBAddress, err := client.metadata.Get("BoshDBAddress")
	if err != nil {
		return store["state.json"], err
	}
	boshDbPort, err := client.metadata.Get("BoshDBPort")
	if err != nil {
		return store["state.json"], err
	}
	blobstoreUserAccessKeyID, err := client.metadata.Get("BlobstoreUserAccessKeyID")
	if err != nil {
		return store["state.json"], err
	}
	blobstoreSecretAccessKey, err := client.metadata.Get("BlobstoreSecretAccessKey")
	if err != nil {
		return store["state.json"], err
	}
	directorKeyPair, err := client.metadata.Get("DirectorKeyPair")
	if err != nil {
		return store["state.json"], err
	}
	directorSecurityGroup, err := client.metadata.Get("DirectorSecurityGroupID")
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
