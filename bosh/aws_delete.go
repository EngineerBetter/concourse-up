package bosh

import (
	"fmt"
	"net"
	"os"

	"github.com/EngineerBetter/concourse-up/bosh/internal/aws"
	"github.com/EngineerBetter/concourse-up/db"
	"github.com/apparentlymart/go-cidr/cidr"
)

// Delete deletes a bosh director
func (client *AWSClient) Delete(stateFileBytes []byte) ([]byte, error) {
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve director IP: [%v]", err)
	}

	if err = client.boshCLI.RunAuthenticatedCommand(
		"delete-deployment",
		directorPublicIP,
		client.config.DirectorPassword,
		client.config.DirectorCACert,
		false,
		os.Stdout,
		"--force",
	); err != nil {
		return nil, err
	}

	//TODO(px): pull up this so that we use aws.Store
	store := temporaryStore{
		"state.json": stateFileBytes,
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

	publicCIDR := client.config.PublicCIDR
	_, pubCIDR, err1 := net.ParseCIDR(publicCIDR)
	if err1 != nil {
		return store["state.json"], err
	}
	internalGateway, err1 := cidr.Host(pubCIDR, 1)
	if err1 != nil {
		return store["state.json"], err
	}
	directorInternalIP, err1 := cidr.Host(pubCIDR, 6)
	if err1 != nil {
		return store["state.json"], err
	}

	err = client.boshCLI.DeleteEnv(store, aws.Environment{
		InternalCIDR:    client.config.PublicCIDR,
		InternalGateway: internalGateway.String(),
		InternalIP:      directorInternalIP.String(),
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
