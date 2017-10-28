package bosh

import (
	"io"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/db"
	"github.com/EngineerBetter/concourse-up/director"
	"github.com/EngineerBetter/concourse-up/terraform"
)

const cloudConfigFilename = "cloud-config.yml"

// StateFilename is default name for bosh-init state file
const StateFilename = "director-state.json"

// Client is a concrete implementation of the IClient interface
type Client struct {
	config   *config.Config
	metadata *terraform.Metadata
	director director.IClient
	dbRunner db.Runner
	stdout   io.Writer
	stderr   io.Writer
}

// IClient is a client for performing bosh-init commands
//go:generate counterfeiter . IClient
type IClient interface {
	Deploy([]byte, bool) ([]byte, error)
	Delete([]byte) ([]byte, error)
	Cleanup() error
	Instances() ([]Instance, error)
}

// ClientFactory creates a new IClient
type ClientFactory func(config *config.Config, metadata *terraform.Metadata, director director.IClient, dbRunner db.Runner, stdout, stderr io.Writer) IClient

// NewClient creates a new Client
func NewClient(config *config.Config, metadata *terraform.Metadata, director director.IClient, dbRunner db.Runner, stdout, stderr io.Writer) IClient {
	return &Client{
		config:   config,
		metadata: metadata,
		director: director,
		dbRunner: dbRunner,
		stdout:   stdout,
		stderr:   stderr,
	}
}

// Cleanup cleans up temporary files associated with bosh init
func (client *Client) Cleanup() error {
	return client.director.Cleanup()
}
