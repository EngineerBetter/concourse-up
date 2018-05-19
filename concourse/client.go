package concourse

import (
	"io"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/director"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"
)

// Client is a concrete implementation of IClient interface
type Client struct {
	iaasClient             iaas.IClient
	terraformClientFactory terraform.ClientFactory
	boshClientFactory      bosh.ClientFactory
	flyClientFactory       func(fly.Credentials, io.Writer, io.Writer) (fly.IClient, error)
	certGenerator          func(c certs.AcmeClient, caName string, ip ...string) (*certs.Certs, error)
	configClient           config.IClient
	deployArgs             *config.DeployArgs
	stdout                 io.Writer
	stderr                 io.Writer
	ipChecker              func() (string, error)
	acmeClient             certs.AcmeClient
}

// IClient represents a concourse-up client
type IClient interface {
	Deploy() error
	Destroy() error
	FetchInfo() (*Info, error)
}

// NewClient returns a new Client
func NewClient(
	iaasClient iaas.IClient,
	terraformClientFactory terraform.ClientFactory,
	boshClientFactory bosh.ClientFactory,
	flyClientFactory func(fly.Credentials, io.Writer, io.Writer) (fly.IClient, error),
	certGenerator func(c certs.AcmeClient, caName string, ip ...string) (*certs.Certs, error),
	configClient config.IClient,
	deployArgs *config.DeployArgs,
	stdout, stderr io.Writer, ipChecker func() (string, error), acmeClient certs.AcmeClient) *Client {
	return &Client{
		iaasClient:             iaasClient,
		terraformClientFactory: terraformClientFactory,
		boshClientFactory:      boshClientFactory,
		flyClientFactory:       flyClientFactory,
		configClient:           configClient,
		certGenerator:          certGenerator,
		deployArgs:             deployArgs,
		stdout:                 stdout,
		stderr:                 stderr,
		ipChecker:              ipChecker,
		acmeClient:             acmeClient,
	}
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

	return client.boshClientFactory(
		config,
		metadata,
		director,
		client.stdout,
		client.stderr,
	)
}
