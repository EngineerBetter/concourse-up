package concourse

import (
	"io"

	"bitbucket.org/engineerbetter/concourse-up/bosh"
	"bitbucket.org/engineerbetter/concourse-up/certs"
	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
	"bitbucket.org/engineerbetter/concourse-up/util"
)

// Client is a concrete implementation of IClient interface
type Client struct {
	terraformClientFactory terraform.ClientFactory
	boshClientFactory      bosh.ClientFactory
	certGenerator          func(caName, ip string) (*certs.Certs, error)
	configClient           config.IClient
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
	certGenerator func(caName, ip string) (*certs.Certs, error),
	configClient config.IClient, stdout, stderr io.Writer) *Client {
	return &Client{
		terraformClientFactory: terraformClientFactory,
		boshClientFactory:      boshClientFactory,
		configClient:           configClient,
		certGenerator:          certGenerator,
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
	return client.boshClientFactory(
		config,
		metadata,
		client.stdout,
		client.stderr,
	)
}

