package director

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"bitbucket.org/engineerbetter/concourse-up/util"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

const boshInitLogLevel = boshlog.LevelWarn
const pemFilename = "director.pem"
const manifestFilename = "director.yml"

// StateFilename is default name for bosh-init state file
const StateFilename = "director-state.json"

// BoshInitClient is a concrete implementation of the IBoshInitClient interface
type BoshInitClient struct {
	tempDir       string
	manifestPath  string
	stateFilePath string
	stdout        io.Writer
	stderr        io.Writer
}

// IBoshInitClient is a client for performing bosh-init commands
type IBoshInitClient interface {
	Deploy() ([]byte, error)
	Delete() error
	Cleanup() error
}

// BoshInitClientFactory creates a new IBoshInitClient
type BoshInitClientFactory func(manifestBytes, stateFileBytes, keyfileBytes []byte, stdout, stderr io.Writer) (IBoshInitClient, error)

// NewBoshInitClient creates a new BoshInitClient
func NewBoshInitClient(manifestBytes, stateFileBytes, keyfileBytes []byte, stdout, stderr io.Writer) (IBoshInitClient, error) {
	tempDir, err := ioutil.TempDir("", "concourse-up")
	if err != nil {
		return nil, err
	}

	keyFilePath := filepath.Join(tempDir, pemFilename)
	if err = ioutil.WriteFile(keyFilePath, keyfileBytes, 0700); err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(tempDir, manifestFilename)
	if err = ioutil.WriteFile(manifestPath, manifestBytes, 0700); err != nil {
		return nil, err
	}

	stateFilePath := filepath.Join(tempDir, StateFilename)
	if stateFileBytes != nil {
		if err = ioutil.WriteFile(stateFilePath, stateFileBytes, 0700); err != nil {
			return nil, err
		}
	}

	return &BoshInitClient{
		tempDir:       tempDir,
		manifestPath:  manifestPath,
		stateFilePath: stateFilePath,
		stdout:        stdout,
		stderr:        stderr,
	}, nil
}

// Cleanup cleans up temporary files associated with bosh init
func (client *BoshInitClient) Cleanup() error {
	return os.RemoveAll(client.tempDir)
}

// Deploy deploys a new Bosh director or converges an existing deployment
func (client *BoshInitClient) Deploy() ([]byte, error) {
	// deploy command needs to be run from directory with bosh state file
	err := util.PushDir(client.tempDir, func() error {
		return client.runBoshCommand(
			"--non-interactive",
			"create-env",
			client.manifestPath,
			"--state",
			client.stateFilePath,
		)
	})
	if err != nil {
		return nil, err
	}

	return ioutil.ReadFile(client.stateFilePath)
}

// Delete deletes a bosh director
func (client *BoshInitClient) Delete() error {
	return client.runBoshCommand(
		"--non-interactive",
		"delete-env",
		client.manifestPath,
		"--state",
		client.stateFilePath,
	)
}
