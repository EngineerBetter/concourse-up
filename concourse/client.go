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
	flyClientFactory       func(fly.Credentials, io.Writer, io.Writer, []byte) (fly.IClient, error)
	certGenerator          func(constructor func(u *certs.User) (certs.AcmeClient, error), caName string, ip ...string) (*certs.Certs, error)
	configClient           config.IClient
	deployArgs             *config.DeployArgs
	stdout                 io.Writer
	stderr                 io.Writer
	ipChecker              func() (string, error)
	acmeClientConstructor  func(u *certs.User) (certs.AcmeClient, error)
	versionFile            []byte
}

// IClient represents a concourse-up client
type IClient interface {
	Deploy() error
	Destroy() error
	FetchInfo() (*Info, error)
}

//go:generate go-bindata -pkg $GOPACKAGE ../resources/director-versions.json
var versionFile = MustAsset("../resources/director-versions.json")

// NewClient returns a new Client
func NewClient(
	iaasClient iaas.IClient,
	terraformClientFactory terraform.ClientFactory,
	boshClientFactory bosh.ClientFactory,
	flyClientFactory func(fly.Credentials, io.Writer, io.Writer, []byte) (fly.IClient, error),
	certGenerator func(constructor func(u *certs.User) (certs.AcmeClient, error), caName string, ip ...string) (*certs.Certs, error),
	configClient config.IClient,
	deployArgs *config.DeployArgs,
	stdout, stderr io.Writer, ipChecker func() (string, error), acmeClientConstructor func(u *certs.User) (certs.AcmeClient, error)) *Client {
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
		acmeClientConstructor:  acmeClientConstructor,
		versionFile:            versionFile,
	}
}

func (client *Client) buildBoshClient(config *config.Config, metadata *terraform.Metadata) (bosh.IClient, error) {
	director, err := director.NewClient(director.Credentials{
		Username: config.DirectorUsername,
		Password: config.DirectorPassword,
		Host:     metadata.DirectorPublicIP.Value,
		CACert:   config.DirectorCACert,
	}, versionFile)
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
