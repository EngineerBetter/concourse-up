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
	state, creds, err = client.createEnv(state, creds)
	if err != nil {
		return state, creds, err
	}

	if err = client.updateCloudConfig(); err != nil {
		return state, creds, err
	}

	if err = client.uploadConcourseStemcell(); err != nil {
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

func (client *Client) createEnv(state, creds []byte) (newState, newCreds []byte, err error) {
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
	bosh, err := boshenv.New(boshenv.DownloadBOSH())
	if err != nil {
		return state, creds, err
	}
	if err != nil {
		return store["state.json"], store["vars.yaml"], err
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
	}, client.config.DirectorPassword, client.config.DirectorCert, client.config.DirectorKey, client.config.DirectorCACert, tags)
	return store["state.json"], store["vars.yaml"], err
}

func (client *Client) updateCloudConfig() error {
	cloudConfigBytes, err := generateCloudConfig(client.config, client.metadata)
	if err != nil {
		return err
	}

	cloudConfigPath, err := client.director.SaveFileToWorkingDir(cloudConfigFilename, cloudConfigBytes)
	if err != nil {
		return err
	}

	return client.director.RunAuthenticatedCommand(
		client.stdout,
		client.stderr,
		false,
		"update-cloud-config",
		cloudConfigPath,
	)
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
