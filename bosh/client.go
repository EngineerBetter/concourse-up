package bosh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
	config   *config.Config
	metadata *terraform.Metadata
	stdout   io.Writer
	stderr   io.Writer
	*tempFiles
}

type tempFiles struct {
	tempDir               *util.TempDir
	directorManifestPath  string
	concourseManifestPath string
	cloudConfigPath       string
	stateFilePath         string
	caCertPath            string
	keyFilePath           string
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
	tempFiles, err := writeClientTempFiles(config, metadata, stateFileBytes)
	if err != nil {
		return nil, err
	}

	return &Client{
		config:    config,
		metadata:  metadata,
		stdout:    stdout,
		stderr:    stderr,
		tempFiles: tempFiles,
	}, nil
}

// Cleanup cleans up temporary files associated with bosh init
func (client *Client) Cleanup() error {
	return client.tempFiles.tempDir.Cleanup()
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
	err := client.tempDir.PushDir(func() error {
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

func writeClientTempFiles(config *config.Config, metadata *terraform.Metadata, stateFileBytes []byte) (*tempFiles, error) {
	tempDir, err := util.NewTempDir()
	if err != nil {
		return nil, err
	}

	directorManifestBytes, err := generateBoshInitManifest(config, metadata)
	if err != nil {
		return nil, err
	}

	directorManifestPath, err := tempDir.Save(directorManifestFilename, directorManifestBytes)
	if err != nil {
		return nil, err
	}

	concourseManifestBytes, err := generateConcourseManifest(config, metadata)
	if err != nil {
		return nil, err
	}

	concourseManifestPath, err := tempDir.Save(concourseManifestFilename, concourseManifestBytes)
	if err != nil {
		return nil, err
	}

	cloudConfigBytes, err := generateCloudConfig(config, metadata)
	if err != nil {
		return nil, err
	}

	cloudConfigPath, err := tempDir.Save(cloudConfigFilename, cloudConfigBytes)
	if err != nil {
		return nil, err
	}

	keyFilePath, err := tempDir.Save(pemFilename, []byte(config.PrivateKey))
	if err != nil {
		return nil, err
	}

	caCertPath, err := tempDir.Save(caCertFilename, []byte(config.DirectorCACert))
	if err != nil {
		return nil, err
	}

	stateFilePath, err := tempDir.Save(StateFilename, stateFileBytes)
	if err != nil {
		return nil, err
	}

	return &tempFiles{
		tempDir:               tempDir,
		stateFilePath:         stateFilePath,
		caCertPath:            caCertPath,
		directorManifestPath:  directorManifestPath,
		concourseManifestPath: concourseManifestPath,
		keyFilePath:           keyFilePath,
		cloudConfigPath:       cloudConfigPath,
	}, nil
}
