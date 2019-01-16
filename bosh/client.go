package bosh

import (
	"fmt"
	"io"

	"github.com/EngineerBetter/concourse-up/iaas"

	"github.com/EngineerBetter/concourse-up/terraform"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/director"
)

const awsConst = "AWS"
const gcpConst = "GCP"

const cloudConfigFilename = "cloud-config.yml"

// StateFilename is default name for bosh-init state file
const StateFilename = "director-state.json"

// CredsFilename is default name for bosh-init creds file
const CredsFilename = "director-creds.yml"

// Client is a concrete implementation of the IClient interface
type Client struct {
	config   config.Config
	metadata terraform.IAASMetadata
	director director.IClient
	db       Opener
	stdout   io.Writer
	stderr   io.Writer
	provider iaas.Provider
}

// IClient is a client for performing bosh-init commands
type IClient interface {
	Deploy([]byte, []byte, bool) ([]byte, []byte, error)
	Delete([]byte) ([]byte, error)
	Cleanup() error
	Instances() ([]Instance, error)
	CreateEnv([]byte, []byte, string) ([]byte, []byte, error)
	Recreate() error
	Locks() ([]byte, error)
}

// Instance represents a vm deployed by BOSH
type Instance struct {
	Name  string
	IP    string
	State string
}

// ClientFactory creates a new IClient
type ClientFactory func(config config.Config, metadata terraform.IAASMetadata, director director.IClient, stdout, stderr io.Writer, provider iaas.Provider) (IClient, error)

//New returns an IAAS specific implementation of BOSH client
func New(config config.Config, metadata terraform.IAASMetadata, director director.IClient, stdout, stderr io.Writer, provider iaas.Provider) (IClient, error) {
	switch provider.IAAS() {
	case awsConst:
		return NewAWSClient(config, metadata, director, stdout, stderr, provider)
	case gcpConst:
		return NewGCPClient(config, metadata, director, stdout, stderr, provider)
	}
	return nil, fmt.Errorf("IAAS not supported: %s", provider.IAAS())
}
