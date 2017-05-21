package concourse

import (
	"io"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/db"
	"github.com/EngineerBetter/concourse-up/director"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"
)

// Client is a concrete implementation of IClient interface
type Client struct {
	terraformClientFactory terraform.ClientFactory
	boshClientFactory      bosh.ClientFactory
	certGenerator          func(caName string, ip ...string) (*certs.Certs, error)
	hostedZoneFinder       func(string) (string, string, error)
	configClient           config.IClient
	deployArgs             *config.DeployArgs
	stdout                 io.Writer
	stderr                 io.Writer
}

// IClient represents a concourse-up client
type IClient interface {
	Deploy() error
	Destroy() error
	FetchInfo() (*Info, error)
}

// NewClient returns a new Client
func NewClient(
	terraformClientFactory terraform.ClientFactory,
	boshClientFactory bosh.ClientFactory,
	certGenerator func(caName string, ip ...string) (*certs.Certs, error),
	hostedZoneFinder func(string) (string, string, error),
	configClient config.IClient,
	deployArgs *config.DeployArgs,
	stdout, stderr io.Writer) *Client {
	return &Client{
		terraformClientFactory: terraformClientFactory,
		boshClientFactory:      boshClientFactory,
		configClient:           configClient,
		hostedZoneFinder:       hostedZoneFinder,
		certGenerator:          certGenerator,
		deployArgs:             deployArgs,
		stdout:                 stdout,
		stderr:                 stderr,
	}
}

func (client *Client) buildTerraformClient(config *config.Config) (terraform.IClient, error) {
	terraformFile, err := util.RenderTemplate(terraform.Template, config)
	if err != nil {
		return nil, err
	}

	return client.terraformClientFactory(terraformFile, client.stdout, client.stderr)
}

func (client *Client) buildBoshClient(config *config.Config, metadata *terraform.Metadata) (bosh.IClient, error) {
	director, err := director.NewClient(director.Credentials{
		Username: config.DirectorUsername,
		Password: config.DirectorPassword,
		Host:     metadata.DirectorPublicIP.Value,
		CACert:   config.DirectorCACert,
	})

	if err != nil {
		return nil, err
	}

	dbRunner, err := db.NewRunner(&db.Credentials{
		Username:      config.RDSUsername,
		Password:      config.RDSPassword,
		Address:       metadata.BoshDBAddress.Value,
		Port:          metadata.BoshDBPort.Value,
		DB:            config.RDSDefaultDatabaseName,
		CACert:        db.RDSRootCert,
		SSHPrivateKey: []byte(config.PrivateKey),
		SSHPublicIP:   metadata.DirectorPublicIP.Value,
	})

	if err != nil {
		return nil, err
	}

	return client.boshClientFactory(
		config,
		metadata,
		director,
		dbRunner,
		client.stdout,
		client.stderr,
	), nil
}
