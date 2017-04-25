package bosh

import (
	"io"

	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/director"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
)

const cloudConfigFilename = "cloud-config.yml"

// StateFilename is default name for bosh-init state file
const StateFilename = "director-state.json"

// Client is a concrete implementation of the IClient interface
type Client struct {
	config   *config.Config
	metadata *terraform.Metadata
	stdout   io.Writer
	stderr   io.Writer
	director *director.Client
}

// IClient is a client for performing bosh-init commands
type IClient interface {
	Deploy([]byte) ([]byte, error)
	Delete([]byte) ([]byte, error)
	Cleanup() error
}

// ClientFactory creates a new IClient
type ClientFactory func(config *config.Config, metadata *terraform.Metadata, stdout, stderr io.Writer) (IClient, error)

// NewClient creates a new Client
func NewClient(config *config.Config, metadata *terraform.Metadata, stdout, stderr io.Writer) (IClient, error) {
	director, err := director.NewClient(director.Credentials{
		Username: config.DirectorUsername,
		Password: config.DirectorPassword,
		Host:     metadata.DirectorPublicIP.Value,
		CACert:   config.DirectorCACert,
	}, stdout, stderr)

	if err != nil {
		return nil, err
	}

	return &Client{
		config:   config,
		metadata: metadata,
		stdout:   stdout,
		stderr:   stderr,
		director: director,
	}, nil
}

// Cleanup cleans up temporary files associated with bosh init
func (client *Client) Cleanup() error {
	return client.director.Cleanup()
}

// func writeClientTempFiles(config *config.Config, metadata *terraform.Metadata, stateFileBytes []byte) (*tempFiles, error) {
// 	tempDir, err := util.NewTempDir()
// 	if err != nil {
// 		return nil, err
// 	}

// 	directorManifestBytes, err := generateBoshInitManifest(config, metadata)
// 	if err != nil {
// 		return nil, err
// 	}

// 	directorManifestPath, err := tempDir.Save(directorManifestFilename, directorManifestBytes)
// 	if err != nil {
// 		return nil, err
// 	}

// 	concourseManifestBytes, err := generateConcourseManifest(config, metadata)
// 	if err != nil {
// 		return nil, err
// 	}

// 	concourseManifestPath, err := tempDir.Save(concourseManifestFilename, concourseManifestBytes)
// 	if err != nil {
// 		return nil, err
// 	}

// 	cloudConfigBytes, err := generateCloudConfig(config, metadata)
// 	if err != nil {
// 		return nil, err
// 	}

// 	cloudConfigPath, err := tempDir.Save(cloudConfigFilename, cloudConfigBytes)
// 	if err != nil {
// 		return nil, err
// 	}

// 	keyFilePath, err := tempDir.Save(pemFilename, []byte(config.PrivateKey))
// 	if err != nil {
// 		return nil, err
// 	}

// 	caCertPath, err := tempDir.Save(caCertFilename, []byte(config.DirectorCACert))
// 	if err != nil {
// 		return nil, err
// 	}

// 	stateFilePath, err := tempDir.Save(StateFilename, stateFileBytes)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &tempFiles{
// 		tempDir:               tempDir,
// 		stateFilePath:         stateFilePath,
// 		caCertPath:            caCertPath,
// 		directorManifestPath:  directorManifestPath,
// 		concourseManifestPath: concourseManifestPath,
// 		keyFilePath:           keyFilePath,
// 		cloudConfigPath:       cloudConfigPath,
// 	}, nil
// }
