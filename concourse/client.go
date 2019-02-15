package concourse

import (
	"io"

	"github.com/EngineerBetter/concourse-up/commands/maintain"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/commands/deploy"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"

	"github.com/xenolf/lego/lego"
)

// client is a concrete implementation of IClient interface
type Client struct {
	acmeClientConstructor func(u *certs.User) (*lego.Client, error)
	boshClientFactory     bosh.ClientFactory
	certGenerator         func(constructor func(u *certs.User) (*lego.Client, error), caName string, provider iaas.Provider, ip ...string) (*certs.Certs, error)
	configClient          config.IClient
	deployArgs            *deploy.Args
	eightRandomLetters    func() string
	flyClientFactory      func(iaas.Provider, fly.Credentials, io.Writer, io.Writer, []byte) (fly.IClient, error)
	ipChecker             func() (string, error)
	passwordGenerator     func(int) string
	provider              iaas.Provider
	sshGenerator          func() ([]byte, []byte, string, error)
	stderr                io.Writer
	stdout                io.Writer
	tfCLI                 terraform.CLIInterface
	tfInputVarsFactory    TFInputVarsFactory
	version               string
	versionFile           []byte
}

// IClient represents a concourse-up client
type IClient interface {
	Deploy() error
	Destroy() error
	FetchInfo() (*Info, error)
	Maintain(maintain.Args) error
}

//go:generate go-bindata -pkg $GOPACKAGE ../../concourse-up-ops/director-versions-aws.json ../../concourse-up-ops/director-versions-gcp.json
var awsVersionFile = MustAsset("../../concourse-up-ops/director-versions-aws.json")
var gcpVersionFile = MustAsset("../../concourse-up-ops/director-versions-gcp.json")

// New returns a new client
func NewClient(
	provider iaas.Provider,
	tfCLI terraform.CLIInterface,
	tfInputVarsFactory TFInputVarsFactory,
	boshClientFactory bosh.ClientFactory,
	flyClientFactory func(iaas.Provider, fly.Credentials, io.Writer, io.Writer, []byte) (fly.IClient, error),
	certGenerator func(constructor func(u *certs.User) (*lego.Client, error), caName string, provider iaas.Provider, ip ...string) (*certs.Certs, error),
	configClient config.IClient,
	deployArgs *deploy.Args,
	stdout, stderr io.Writer,
	ipChecker func() (string, error),
	acmeClientConstructor func(u *certs.User) (*lego.Client, error),
	passwordGenerator func(int) string,
	eightRandomLetters func() string,
	sshGenerator func() ([]byte, []byte, string, error),
	version string) *Client {
	v, _ := provider.Choose(iaas.Choice{
		AWS: awsVersionFile,
		GCP: gcpVersionFile,
	}).([]byte)
	return &Client{
		acmeClientConstructor: acmeClientConstructor,
		boshClientFactory:     boshClientFactory,
		certGenerator:         certGenerator,
		configClient:          configClient,
		deployArgs:            deployArgs,
		eightRandomLetters:    eightRandomLetters,
		flyClientFactory:      flyClientFactory,
		ipChecker:             ipChecker,
		passwordGenerator:     passwordGenerator,
		provider:              provider,
		sshGenerator:          sshGenerator,
		stderr:                stderr,
		stdout:                stdout,
		tfCLI:                 tfCLI,
		tfInputVarsFactory:    tfInputVarsFactory,
		version:               version,
		versionFile:           v,
	}
}

func (client *Client) buildBoshClient(config config.Config, tfOutputs terraform.Outputs) (bosh.IClient, error) {

	return client.boshClientFactory(
		config,
		tfOutputs,
		client.stdout,
		client.stderr,
		client.provider,
		client.versionFile,
	)
}
