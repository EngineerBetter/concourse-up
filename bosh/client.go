package bosh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
	"bitbucket.org/engineerbetter/concourse-up/util"
	boshcmd "github.com/cloudfoundry/bosh-cli/cmd"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

const boshInitLogLevel = boshlog.LevelWarn
const pemFilename = "director.pem"
const directorManifestFilename = "director.yml"
const concourseManifestFilename = "concourse.yml"
const cloudConfigFilename = "cloud-config.yml"
const caCertFilename = "ca-cert.pem"

var defaultBoshArgs = []string{"--non-interactive", "--tty", "--no-color"}

// StateFilename is default name for bosh-init state file
const StateFilename = "director-state.json"

// Client is a concrete implementation of the IClient interface
type Client struct {
	tempDir               string
	directorManifestPath  string
	concourseManifestPath string
	cloudConfigPath       string
	stateFilePath         string
	caCertPath            string
	config                *config.Config
	metadata              *terraform.Metadata
	stdout                io.Writer
	stderr                io.Writer
}

// IClient is a client for performing bosh-init commands
type IClient interface {
	Deploy() ([]byte, error)
	Delete() error
	Cleanup() error
}

// ClientFactory creates a new IClient
type ClientFactory func(config *config.Config, metadata *terraform.Metadata, stateFileBytes []byte, stdout, stderr io.Writer) (IClient, error)

// NewClient creates a new Client
func NewClient(config *config.Config, metadata *terraform.Metadata, stateFileBytes []byte, stdout, stderr io.Writer) (IClient, error) {
	tempDir, err := ioutil.TempDir("", "concourse-up")
	if err != nil {
		return nil, err
	}

	keyfileBytes := []byte(config.PrivateKey)
	keyFilePath := filepath.Join(tempDir, pemFilename)
	if err = ioutil.WriteFile(keyFilePath, keyfileBytes, 0700); err != nil {
		return nil, err
	}

	manifestBytes, err := generateBoshInitManifest(config, metadata)
	if err != nil {
		return nil, err
	}

	directorManifestPath := filepath.Join(tempDir, directorManifestFilename)
	if err = ioutil.WriteFile(directorManifestPath, manifestBytes, 0700); err != nil {
		return nil, err
	}

	concourseManifestBytes, err := generateConcourseManifest(config, metadata)
	if err != nil {
		return nil, err
	}

	concourseManfestPath := filepath.Join(tempDir, concourseManifestFilename)
	if err = ioutil.WriteFile(concourseManfestPath, concourseManifestBytes, 0700); err != nil {
		return nil, err
	}

	cloudConfigBytes, err := generateCloudConfig(config, metadata)
	if err != nil {
		return nil, err
	}

	cloudConfigPath := filepath.Join(tempDir, cloudConfigFilename)
	if err = ioutil.WriteFile(cloudConfigPath, cloudConfigBytes, 0700); err != nil {
		return nil, err
	}

	caCertPath := filepath.Join(tempDir, caCertFilename)
	if err = ioutil.WriteFile(caCertPath, []byte(config.DirectorCACert), 0700); err != nil {
		return nil, err
	}

	stateFilePath := filepath.Join(tempDir, StateFilename)
	if stateFileBytes != nil {
		if err = ioutil.WriteFile(stateFilePath, stateFileBytes, 0700); err != nil {
			return nil, err
		}
	}

	return &Client{
		tempDir:              tempDir,
		directorManifestPath: directorManifestPath,
		cloudConfigPath:      cloudConfigPath,
		stateFilePath:        stateFilePath,
		caCertPath:           caCertPath,
		config:               config,
		metadata:             metadata,
		stdout:               stdout,
		stderr:               stderr,
	}, nil
}

// Cleanup cleans up temporary files associated with bosh init
func (client *Client) Cleanup() error {
	return os.RemoveAll(client.tempDir)
}

// Deploy deploys a new Bosh director or converges an existing deployment
// Returns new contents of bosh state file
func (client *Client) Deploy() ([]byte, error) {
	if err := client.createEnv(); err != nil {
		return nil, err
	}

	if err := client.updateCloudConfig(); err != nil {
		return nil, err
	}

	if err := client.uploadConcourse(); err != nil {
		return nil, err
	}

	if err := client.createDefaultDatabases(); err != nil {
		return nil, err
	}

	return ioutil.ReadFile(client.stateFilePath)
}

func (client *Client) createEnv() error {
	// deploy command needs to be run from directory with bosh state file
	var combinedOutput []byte
	err := util.PushDir(client.tempDir, func() error {
		var e error
		combinedOutput, e = client.runBoshCommand(
			"create-env",
			client.directorManifestPath,
			"--state",
			client.stateFilePath,
		)
		return e
	})
	if err != nil {
		return err
	}
	if !strings.Contains(string(combinedOutput), "Finished deploying") && !strings.Contains(string(combinedOutput), "Skipping deploy") {
		return errors.New("Couldn't find string `Finished deploying` or `Skipping deploy` in bosh stdout/stderr output")
	}

	return nil
}

// Delete deletes a bosh director
func (client *Client) Delete() error {
	_, err := client.runBoshCommand(
		"delete-env",
		client.directorManifestPath,
		"--state",
		client.stateFilePath,
	)

	return err
}

func (client *Client) updateCloudConfig() error {
	_, err := client.runAuthenticatedBoshCommand(
		"update-cloud-config",
		client.cloudConfigPath,
	)

	return err
}

func (client *Client) runAuthenticatedBoshCommand(args ...string) ([]byte, error) {
	args = append([]string{
		"--environment",
		fmt.Sprintf("https://%s", client.metadata.DirectorPublicIP.Value),
		"--ca-cert",
		client.caCertPath,
		"--client",
		client.config.DirectorUsername,
		"--client-secret",
		client.config.DirectorPassword,
	}, args...)

	return client.runBoshCommand(args...)
}

// https://github.com/cloudfoundry/bosh-cli/blob/master/main.go
func (client *Client) runBoshCommand(args ...string) ([]byte, error) {
	combinedOutputBuffer := bytes.NewBuffer(nil)
	stdout := io.MultiWriter(client.stdout, combinedOutputBuffer)
	stderr := io.MultiWriter(client.stderr, combinedOutputBuffer)

	logger := boshlog.NewWriterLogger(boshInitLogLevel,
		stdout,
		stderr,
	)

	ui := boshui.NewConfUI(logger)
	defer ui.Flush()
	writerUI := boshui.NewWriterUI(stdout, stderr, logger)

	// NOTE SetParent is implemented manually on vendored version of bosh-cli
	ui.SetParent(writerUI)

	basicDeps := boshcmd.NewBasicDeps(ui, logger)
	cmdFactory := boshcmd.NewFactory(basicDeps)

	args = append(defaultBoshArgs, args...)

	cmd, err := cmdFactory.New(args)
	if err != nil {
		return nil, err
	}

	if err = cmd.Execute(); err != nil {
		return nil, err
	}

	ui.Flush()

	stdoutBytes, err := ioutil.ReadAll(combinedOutputBuffer)
	if err != nil {
		return nil, err
	}

	return stdoutBytes, nil
}
