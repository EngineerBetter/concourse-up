package bosh

import (
	"github.com/engineerbetter/concourse-up/config"
	"github.com/engineerbetter/concourse-up/db"
	"github.com/engineerbetter/concourse-up/director"
	"github.com/engineerbetter/concourse-up/terraform"
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
}

// IClient is a client for performing bosh-init commands
type IClient interface {
	Deploy([]byte) ([]byte, error)
	Delete([]byte) ([]byte, error)
	Cleanup() error
	Instances() ([]Instance, error)
}

// ClientFactory creates a new IClient
type ClientFactory func(config *config.Config, metadata *terraform.Metadata, director director.IClient, dbRunner db.Runner) IClient

// NewClient creates a new Client
func NewClient(config *config.Config, metadata *terraform.Metadata, director director.IClient, dbRunner db.Runner) IClient {
	return &Client{
		config:   config,
		metadata: metadata,
		director: director,
		dbRunner: dbRunner,
	}
}

// Cleanup cleans up temporary files associated with bosh init
func (client *Client) Cleanup() error {
	return client.director.Cleanup()
}
