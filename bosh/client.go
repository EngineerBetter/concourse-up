package bosh

import (
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
	director director.IClient
}

// IClient is a client for performing bosh-init commands
type IClient interface {
	Deploy([]byte) ([]byte, error)
	Delete([]byte) ([]byte, error)
	Cleanup() error
}

// ClientFactory creates a new IClient
type ClientFactory func(config *config.Config, metadata *terraform.Metadata, director director.IClient) IClient

// NewClient creates a new Client
func NewClient(config *config.Config, metadata *terraform.Metadata, director director.IClient) IClient {
	return &Client{
		config:   config,
		metadata: metadata,
		director: director,
	}
}

// Cleanup cleans up temporary files associated with bosh init
func (client *Client) Cleanup() error {
	return client.director.Cleanup()
}
