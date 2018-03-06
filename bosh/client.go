package bosh

import (
	"fmt"
	"io"
	"net"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/director"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/lib/pq"
	"golang.org/x/crypto/ssh"
)

const cloudConfigFilename = "cloud-config.yml"

// StateFilename is default name for bosh-init state file
const StateFilename = "director-state.json"

// CredsFilename is default name for bosh-init creds file
const CredsFilename = "director-creds.yml"

// Client is a concrete implementation of the IClient interface
type Client struct {
	config   *config.Config
	metadata *terraform.Metadata
	director director.IClient
	db       Opener
	stdout   io.Writer
	stderr   io.Writer
}

// IClient is a client for performing bosh-init commands
type IClient interface {
	Deploy([]byte, []byte, bool) ([]byte, []byte, error)
	Delete([]byte) ([]byte, error)
	Cleanup() error
	Instances() ([]Instance, error)
}

// ClientFactory creates a new IClient
type ClientFactory func(config *config.Config, metadata *terraform.Metadata, director director.IClient, stdout, stderr io.Writer) (IClient, error)

// NewClient creates a new Client
func NewClient(config *config.Config, metadata *terraform.Metadata, director director.IClient, stdout, stderr io.Writer) (IClient, error) {
	key, err := ssh.ParsePrivateKey([]byte(config.PrivateKey))
	if err != nil {
		return nil, err
	}
	dialer, err := ssh.Dial("tcp",
		net.JoinHostPort(metadata.DirectorPublicIP.Value, "22"),
		&ssh.ClientConfig{
			User:            "vcap",
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(key)},
		},
	)
	if err != nil {
		return nil, err
	}
	db, err := newProxyOpener(dialer, &pq.Driver{},
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
			config.RDSUsername,
			config.RDSPassword,
			metadata.BoshDBAddress.Value,
			metadata.BoshDBPort.Value,
			config.RDSDefaultDatabaseName,
		),
	)
	if err != nil {
		return nil, err
	}
	return &Client{
		config:   config,
		metadata: metadata,
		director: director,
		db:       db,
		stdout:   stdout,
		stderr:   stderr,
	}, nil
}

// Cleanup cleans up temporary files associated with bosh init
func (client *Client) Cleanup() error {
	return client.director.Cleanup()
}
