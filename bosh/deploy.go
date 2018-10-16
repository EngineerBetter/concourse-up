package bosh

import (
	"fmt"
	"strings"

	"github.com/EngineerBetter/concourse-up/bosh/internal/aws"
	"github.com/EngineerBetter/concourse-up/bosh/internal/boshenv"
	"github.com/EngineerBetter/concourse-up/db"
)

// Deploy deploys a new Bosh director or converges an existing deployment
// Returns new contents of bosh state file
func (client *Client) Deploy(state, creds []byte, detach bool) (newState, newCreds []byte, err error) {
	bosh, err := boshenv.New(boshenv.DownloadBOSH())
	if err != nil {
		return state, creds, err
	}

	state, creds, err = client.createEnv(bosh, state, creds)
	if err != nil {
		return state, creds, err
	}

	if err = client.updateCloudConfig(bosh); err != nil {
		return state, creds, err
	}

	if err = client.uploadConcourseStemcell(bosh); err != nil {
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

type temporaryStore map[string][]byte

func (s temporaryStore) Set(key string, value []byte) error {
	s[key] = value
	return nil
}

func (s temporaryStore) Get(key string) ([]byte, error) {
	return s[key], nil
}

func splitTags(ts []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, t := range ts {
		ss := strings.SplitN(t, "=", 2)
		if len(ss) != 2 {
			return nil, fmt.Errorf("could not split tag %q", t)
		}
		m[ss[0]] = ss[1]
	}
	return m, nil
}

func (client *Client) createEnv(bosh *boshenv.BOSHCLI, state, creds []byte) (newState, newCreds []byte, err error) {
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

	boshUserAccessKeyID, err := client.metadata.Get("BoshUserAccessKeyID")
	if err != nil {
		return state, creds, err
	}
	boshSecretAccessKey, err := client.metadata.Get("BoshSecretAccessKey")
	if err != nil {
		return state, creds, err
	}
	publicSubnetID, err := client.metadata.Get("PublicSubnetID")
	if err != nil {
		return state, creds, err
	}
	privateSubnetID, err := client.metadata.Get("PrivateSubnetID")
	if err != nil {
		return state, creds, err
	}
	directorPublicIP, err := client.metadata.Get("DirectorPublicIP")
	if err != nil {
		return state, creds, err
	}
	atcSecurityGroupID, err := client.metadata.Get("ATCSecurityGroupID")
	if err != nil {
		return state, creds, err
	}
	vmSecurityGroupID, err := client.metadata.Get("VMsSecurityGroupID")
	if err != nil {
		return state, creds, err
	}
	blobstoreBucket, err := client.metadata.Get("BlobstoreBucket")
	if err != nil {
		return state, creds, err
	}
	boshDBAddress, err := client.metadata.Get("BoshDBAddress")
	if err != nil {
		return state, creds, err
	}
	boshDbPort, err := client.metadata.Get("BoshDBPort")
	if err != nil {
		return state, creds, err
	}
	blobstoreUserAccessKeyID, err := client.metadata.Get("BlobstoreUserAccessKeyID")
	if err != nil {
		return state, creds, err
	}
	blobstoreSecretAccessKey, err := client.metadata.Get("BlobstoreSecretAccessKey")
	if err != nil {
		return state, creds, err
	}
	directorKeyPair, err := client.metadata.Get("DirectorKeyPair")
	if err != nil {
		return state, creds, err
	}
	directorSecurityGroup, err := client.metadata.Get("DirectorSecurityGroupID")
	if err != nil {
		return state, creds, err
	}
	err = bosh.CreateEnv(store, aws.Environment{
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
	}, client.config.DirectorPassword, client.config.DirectorCert, client.config.DirectorKey, client.config.DirectorCACert, tags)

	return store["state.json"], store["vars.yaml"], err
}

func (client *Client) updateCloudConfig(bosh *boshenv.BOSHCLI) error {
	publicSubnetID, err := client.metadata.Get("PublicSubnetID")
	if err != nil {
		return err
	}
	privateSubnetID, err := client.metadata.Get("PrivateSubnetID")
	if err != nil {
		return err
	}
	aTCSecurityGroupID, err := client.metadata.Get("ATCSecurityGroupID")
	if err != nil {
		return err
	}
	vMsSecurityGroupID, err := client.metadata.Get("VMsSecurityGroupID")
	if err != nil {
		return err
	}
	directorPublicIP, err := client.metadata.Get("DirectorPublicIP")
	if err != nil {
		return err
	}
	return bosh.UpdateCloudConfig(aws.Environment{
		AZ:               client.config.AvailabilityZone,
		PublicSubnetID:   publicSubnetID,
		PrivateSubnetID:  privateSubnetID,
		ATCSecurityGroup: aTCSecurityGroupID,
		VMSecurityGroup:  vMsSecurityGroupID,
		Spot:             client.config.Spot,
		ExternalIP:       directorPublicIP,
	}, directorPublicIP, client.config.DirectorPassword, client.config.DirectorCACert)
}
func (client *Client) uploadConcourseStemcell(bosh *boshenv.BOSHCLI) error {
	directorPublicIP, err := client.metadata.Get("DirectorPublicIP")
	if err != nil {
		return err
	}
	return bosh.UploadConcourseStemcell(aws.Environment{
		ExternalIP: directorPublicIP,
	}, directorPublicIP, client.config.DirectorPassword, client.config.DirectorCACert)
}

func (client *Client) saveStateFile(bytes []byte) (string, error) {
	if bytes == nil {
		return client.director.PathInWorkingDir(StateFilename), nil
	}

	return client.director.SaveFileToWorkingDir(StateFilename, bytes)
}

func (client *Client) saveCredsFile(bytes []byte) (string, error) {
	if bytes == nil {
		return client.director.PathInWorkingDir(CredsFilename), nil
	}

	return client.director.SaveFileToWorkingDir(CredsFilename, bytes)
}
