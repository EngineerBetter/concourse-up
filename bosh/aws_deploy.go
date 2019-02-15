package bosh

import (
	"net"

	"github.com/EngineerBetter/concourse-up/bosh/internal/aws"
	"github.com/EngineerBetter/concourse-up/bosh/internal/boshenv"
	"github.com/EngineerBetter/concourse-up/db"
	"github.com/apparentlymart/go-cidr/cidr"
)

// Deploy implements deploy for AWS client
func (client *AWSClient) Deploy(state, creds []byte, detach bool) (newState, newCreds []byte, err error) {
	state, creds, err = client.createEnv(client.boshCLI, state, creds, "")
	if err != nil {
		return state, creds, err
	}

	if err = client.updateCloudConfig(client.boshCLI); err != nil {
		return state, creds, err
	}
	if err = client.uploadConcourseStemcell(client.boshCLI); err != nil {
		return state, creds, err
	}
	if err = client.createDefaultDatabases(); err != nil {
		return state, creds, err
	}

	creds, err = client.deployConcourse(creds, detach)
	if err != nil {
		return state, creds, err
	}

	return state, creds, err
}

// Locks implements locks for AWS client
func (client *AWSClient) Locks() ([]byte, error) {
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return nil, err
	}
	return client.boshCLI.Locks(aws.Environment{
		ExternalIP: directorPublicIP,
	}, directorPublicIP, client.config.DirectorPassword, client.config.DirectorCACert)

}

// CreateEnv exposes bosh create-env functionality
func (client *AWSClient) CreateEnv(state, creds []byte, customOps string) (newState, newCreds []byte, err error) {
	return client.createEnv(client.boshCLI, state, creds, customOps)
}

// Recreate exposes BOSH recreate
func (client *AWSClient) Recreate() error {
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return err
	}
	return client.boshCLI.Recreate(aws.Environment{
		ExternalIP: directorPublicIP,
	}, directorPublicIP, client.config.DirectorPassword, client.config.DirectorCACert)
}

func (client *AWSClient) createEnv(bosh boshenv.IBOSHCLI, state, creds []byte, customOps string) (newState, newCreds []byte, err error) {
	tags, err := splitTags(client.config.Tags)
	if err != nil {
		return state, creds, err
	}
	tags["concourse-up-project"] = client.config.Project
	tags["concourse-up-component"] = "concourse"
	//TODO(px): pull up this so that we use aws.Store
	store := temporaryStore{
		"vars.yaml":  creds,
		"state.json": state,
	}

	boshUserAccessKeyID, err1 := client.outputs.Get("BoshUserAccessKeyID")
	if err1 != nil {
		return state, creds, err1
	}
	boshSecretAccessKey, err1 := client.outputs.Get("BoshSecretAccessKey")
	if err1 != nil {
		return state, creds, err1
	}
	publicSubnetID, err1 := client.outputs.Get("PublicSubnetID")
	if err1 != nil {
		return state, creds, err1
	}
	privateSubnetID, err1 := client.outputs.Get("PrivateSubnetID")
	if err1 != nil {
		return state, creds, err1
	}
	directorPublicIP, err1 := client.outputs.Get("DirectorPublicIP")
	if err1 != nil {
		return state, creds, err1
	}
	atcSecurityGroupID, err1 := client.outputs.Get("ATCSecurityGroupID")
	if err1 != nil {
		return state, creds, err1
	}
	vmSecurityGroupID, err1 := client.outputs.Get("VMsSecurityGroupID")
	if err1 != nil {
		return state, creds, err1
	}
	blobstoreBucket, err1 := client.outputs.Get("BlobstoreBucket")
	if err1 != nil {
		return state, creds, err1
	}
	boshDBAddress, err1 := client.outputs.Get("BoshDBAddress")
	if err1 != nil {
		return state, creds, err1
	}
	boshDbPort, err1 := client.outputs.Get("BoshDBPort")
	if err1 != nil {
		return state, creds, err1
	}
	blobstoreUserAccessKeyID, err1 := client.outputs.Get("BlobstoreUserAccessKeyID")
	if err1 != nil {
		return state, creds, err1
	}
	blobstoreSecretAccessKey, err1 := client.outputs.Get("BlobstoreSecretAccessKey")
	if err1 != nil {
		return state, creds, err1
	}
	directorKeyPair, err1 := client.outputs.Get("DirectorKeyPair")
	if err1 != nil {
		return state, creds, err1
	}
	directorSecurityGroup, err1 := client.outputs.Get("DirectorSecurityGroupID")
	if err1 != nil {
		return state, creds, err1
	}

	publicCIDR := client.config.PublicCIDR
	_, pubCIDR, err1 := net.ParseCIDR(publicCIDR)
	if err1 != nil {
		return state, creds, err1
	}
	internalGateway, err1 := cidr.Host(pubCIDR, 1)
	if err1 != nil {
		return state, creds, err1
	}
	directorInternalIP, err1 := cidr.Host(pubCIDR, 6)
	if err1 != nil {
		return state, creds, err1
	}

	err1 = bosh.CreateEnv(store, aws.Environment{
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
		WorkerType:           client.config.WorkerType,
		CustomOperations:     customOps,
	}, client.config.DirectorPassword, client.config.DirectorCert, client.config.DirectorKey, client.config.DirectorCACert, tags)
	if err1 != nil {
		return store["state.json"], store["vars.yaml"], err1
	}
	return store["state.json"], store["vars.yaml"], err
}

func (client *AWSClient) updateCloudConfig(bosh boshenv.IBOSHCLI) error {
	publicSubnetID, err := client.outputs.Get("PublicSubnetID")
	if err != nil {
		return err
	}
	privateSubnetID, err := client.outputs.Get("PrivateSubnetID")
	if err != nil {
		return err
	}
	aTCSecurityGroupID, err := client.outputs.Get("ATCSecurityGroupID")
	if err != nil {
		return err
	}
	vMsSecurityGroupID, err := client.outputs.Get("VMsSecurityGroupID")
	if err != nil {
		return err
	}
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return err
	}

	publicCIDR := client.config.PublicCIDR
	_, pubCIDR, err := net.ParseCIDR(publicCIDR)
	if err != nil {
		return err
	}
	pubGateway, err := cidr.Host(pubCIDR, 1)
	if err != nil {
		return err
	}
	publicCIDRGateway := pubGateway.String()

	publicCIDRDNS, err := formatIPRange(publicCIDR, "", []int{2})
	if err != nil {
		return err
	}
	publicCIDRStatic, err := formatIPRange(publicCIDR, ", ", []int{6, 7})
	if err != nil {
		return err
	}
	publicCIDRReserved, err := formatIPRange(publicCIDR, "-", []int{1, 5})
	if err != nil {
		return err
	}

	privateCIDR := client.config.PrivateCIDR
	_, privCIDR, err := net.ParseCIDR(privateCIDR)
	if err != nil {
		return err
	}
	privGateway, err := cidr.Host(privCIDR, 1)
	if err != nil {
		return err
	}
	privateCIDRGateway := privGateway.String()

	privateCIDRDNS, err := formatIPRange(privateCIDR, "", []int{2})
	if err != nil {
		return err
	}
	privateCIDRReserved, err := formatIPRange(privateCIDR, "-", []int{1, 5})
	if err != nil {
		return err
	}

	return bosh.UpdateCloudConfig(aws.Environment{
		AZ:                  client.config.AvailabilityZone,
		PublicSubnetID:      publicSubnetID,
		PrivateSubnetID:     privateSubnetID,
		ATCSecurityGroup:    aTCSecurityGroupID,
		VMSecurityGroup:     vMsSecurityGroupID,
		Spot:                client.config.Spot,
		ExternalIP:          directorPublicIP,
		WorkerType:          client.config.WorkerType,
		PublicCIDR:          publicCIDR,
		PublicCIDRGateway:   publicCIDRGateway,
		PublicCIDRDNS:       publicCIDRDNS,
		PublicCIDRStatic:    publicCIDRStatic,
		PublicCIDRReserved:  publicCIDRReserved,
		PrivateCIDR:         privateCIDR,
		PrivateCIDRGateway:  privateCIDRGateway,
		PrivateCIDRDNS:      privateCIDRDNS,
		PrivateCIDRReserved: privateCIDRReserved,
	}, directorPublicIP, client.config.DirectorPassword, client.config.DirectorCACert)
}
func (client *AWSClient) uploadConcourseStemcell(bosh boshenv.IBOSHCLI) error {
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return err
	}
	return bosh.UploadConcourseStemcell(aws.Environment{
		ExternalIP: directorPublicIP,
	}, directorPublicIP, client.config.DirectorPassword, client.config.DirectorCACert)
}

func (client *AWSClient) saveStateFile(bytes []byte) (string, error) {
	if bytes == nil {
		return client.director.PathInWorkingDir(StateFilename), nil
	}

	return client.director.SaveFileToWorkingDir(StateFilename, bytes)
}

func (client *AWSClient) saveCredsFile(bytes []byte) (string, error) {
	if bytes == nil {
		return client.director.PathInWorkingDir(CredsFilename), nil
	}

	return client.director.SaveFileToWorkingDir(CredsFilename, bytes)
}
