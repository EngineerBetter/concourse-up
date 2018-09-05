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
	err = bosh.CreateEnv(store, aws.Environment{
		InternalCIDR:    "10.0.0.0/24",
		InternalGateway: "10.0.0.1",
		InternalIP:      "10.0.0.6",
		AccessKeyID:     client.metadata.BoshUserAccessKeyID.Value,
		SecretAccessKey: client.metadata.BoshSecretAccessKey.Value,
		Region:          client.config.Region,
		AZ:              client.config.AvailabilityZone,
		DefaultKeyName:  client.metadata.DirectorKeyPair.Value,
		DefaultSecurityGroups: []string{
			client.metadata.DirectorSecurityGroupID.Value,
			client.metadata.VMsSecurityGroupID.Value,
		},
		PrivateKey:           client.config.PrivateKey,
		PublicSubnetID:       client.metadata.PublicSubnetID.Value,
		PrivateSubnetID:      client.metadata.PrivateSubnetID.Value,
		ExternalIP:           client.metadata.DirectorPublicIP.Value,
		ATCSecurityGroup:     client.metadata.ATCSecurityGroupID.Value,
		VMSecurityGroup:      client.metadata.VMsSecurityGroupID.Value,
		BlobstoreBucket:      client.metadata.BlobstoreBucket.Value,
		DBCACert:             db.RDSRootCert,
		DBHost:               client.metadata.BoshDBAddress.Value,
		DBName:               client.config.RDSDefaultDatabaseName,
		DBPassword:           client.config.RDSPassword,
		DBPort:               client.metadata.BoshDBPort.Value,
		DBUsername:           client.config.RDSUsername,
		S3AWSAccessKeyID:     client.metadata.BlobstoreUserAccessKeyID.Value,
		S3AWSSecretAccessKey: client.metadata.BlobstoreSecretAccessKey.Value,
	}, client.config.DirectorPassword, client.config.DirectorCert, client.config.DirectorKey, client.config.DirectorCACert, nil)
	return store["state.json"], err
}
