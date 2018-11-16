package bosh

import (
	"fmt"
	"io"
	"net"

	"github.com/EngineerBetter/concourse-up/iaas"

	"github.com/EngineerBetter/concourse-up/terraform"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/director"
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
}

// ClientFactory creates a new IClient
type ClientFactory func(config config.Config, metadata terraform.IAASMetadata, director director.IClient, stdout, stderr io.Writer, provider iaas.Provider) (IClient, error)

// NewClient creates a new Client
func NewClient(config config.Config, metadata terraform.IAASMetadata, director director.IClient, stdout, stderr io.Writer, provider iaas.Provider) (IClient, error) {
	directorPublicIP, err := metadata.Get("DirectorPublicIP")
	if err != nil {
		return nil, err
	}
	addr := net.JoinHostPort(directorPublicIP, "22")
	key, err := ssh.ParsePrivateKey([]byte(config.PrivateKey))
	if err != nil {
		return nil, err
	}
	conf := &ssh.ClientConfig{
		User:            "vcap",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(key)},
	}
	var boshDBAddress, boshDBPort string

	switch provider.IAAS() {
	case "AWS": // nolint
		var err1 error
		boshDBAddress, err1 = metadata.Get("BoshDBAddress")
		if err1 != nil {
			return nil, err1
		}
		boshDBPort, err1 = metadata.Get("BoshDBPort")
		if err1 != nil {
			return nil, err1
		}
	case "GCP": // nolint
		var err1 error
		boshDBAddress, err1 = metadata.Get("DBAddress")
		if err1 != nil {
			return nil, err1
		}
		boshDBPort = "5432"
	}

	db, err := newProxyOpener(addr, conf, &pq.Driver{},
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
			config.RDSUsername,
			config.RDSPassword,
			boshDBAddress,
			boshDBPort,
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
		provider: provider,
	}, nil
}

// Cleanup cleans up temporary files associated with bosh init
func (client *Client) Cleanup() error {
	return client.director.Cleanup()
}
