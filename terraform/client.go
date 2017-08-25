package terraform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/util"
)

const appName string = "concourse-up"

// IClient is an interface for the terraform Client
type IClient interface {
	Output() (*Metadata, error)
	Apply(dryrun bool) error
	Destroy() error
	Cleanup() error
}

// Client wraps common terraform commands
type Client struct {
	iaas      string
	configDir string
	stdout    io.Writer
	stderr    io.Writer
}

// ClientFactory is a function that builds a client interface
type ClientFactory func(iaas string, config *config.Config, stdout, stderr io.Writer) (IClient, error)

// NewClient is a concrete implementation of ClientFactory
func NewClient(iaas string, config *config.Config, stdout, stderr io.Writer) (IClient, error) {
	if iaas != "AWS" {
		return nil, fmt.Errorf("IAAS not supported: %s", iaas)
	}

	terraformFile, err := util.RenderTemplate(AWSTemplate, config)
	if err != nil {
		return nil, err
	}

	if err := checkTerraformOnPath(stderr, stderr); err != nil {
		return nil, err
	}

	configDir, err := initConfig(terraformFile, stderr, stderr)
	if err != nil {
		return nil, err
	}

	return &Client{
		configDir: configDir,
		stdout:    stdout,
		stderr:    stderr,
	}, nil
}

// Cleanup cleans up the temporary directory used by terraform
func (client *Client) Cleanup() error {
	return os.RemoveAll(client.configDir)
}

// Output fetches the terraform output/metadata
func (client *Client) Output() (*Metadata, error) {
	stdoutBuffer := bytes.NewBuffer(nil)
	if err := terraform([]string{
		"output",
		"-json",
	}, client.configDir, stdoutBuffer, client.stderr); err != nil {
		return nil, err
	}

	metadata := Metadata{
		AWS: &AWSMetadata{},
	}
	if err := json.NewDecoder(stdoutBuffer).Decode(metadata.AWS); err != nil {
		return nil, err
	}

	return &metadata, nil
}
