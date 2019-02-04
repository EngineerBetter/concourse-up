package concourse

import (
	"io"

	"github.com/EngineerBetter/concourse-up/commands/maintain"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/commands/deploy"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/director"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"

	"github.com/xenolf/lego/lego"
)

// Client is a concrete implementation of IClient interface
type Client struct {
	provider              iaas.Provider
	tfCLI                 terraform.CLIInterface
	boshClientFactory     bosh.ClientFactory
	flyClientFactory      func(iaas.Provider, fly.Credentials, io.Writer, io.Writer, []byte) (fly.IClient, error)
	certGenerator         func(constructor func(u *certs.User) (*lego.Client, error), caName string, provider iaas.Provider, ip ...string) (*certs.Certs, error)
	configClient          config.IClient
	deployArgs            *deploy.Args
	stdout                io.Writer
	stderr                io.Writer
	ipChecker             func() (string, error)
	acmeClientConstructor func(u *certs.User) (*lego.Client, error)
	versionFile           []byte
	version               string
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

// NewClient returns a new Client
func NewClient(
	provider iaas.Provider,
	tfCLI terraform.CLIInterface,
	boshClientFactory bosh.ClientFactory,
	flyClientFactory func(iaas.Provider, fly.Credentials, io.Writer, io.Writer, []byte) (fly.IClient, error),
	certGenerator func(constructor func(u *certs.User) (*lego.Client, error), caName string, provider iaas.Provider, ip ...string) (*certs.Certs, error),
	configClient config.IClient,
	deployArgs *deploy.Args,
	stdout, stderr io.Writer,
	ipChecker func() (string, error),
	acmeClientConstructor func(u *certs.User) (*lego.Client, error),
	version string) *Client {
	v, _ := provider.Choose(iaas.Choice{
		AWS: awsVersionFile,
		GCP: gcpVersionFile,
	}).([]byte)
	return &Client{
		provider:              provider,
		tfCLI:                 tfCLI,
		boshClientFactory:     boshClientFactory,
		flyClientFactory:      flyClientFactory,
		configClient:          configClient,
		certGenerator:         certGenerator,
		deployArgs:            deployArgs,
		stdout:                stdout,
		stderr:                stderr,
		ipChecker:             ipChecker,
		acmeClientConstructor: acmeClientConstructor,
		versionFile:           v,
		version:               version,
	}
}

func (client *Client) buildBoshClient(config config.Config, metadata terraform.IAASMetadata) (bosh.IClient, error) {
	directorPublicIP, err := metadata.Get("DirectorPublicIP")
	if err != nil {
		return nil, err
	}
	v, _ := client.provider.Choose(iaas.Choice{
		AWS: awsVersionFile,
		GCP: gcpVersionFile,
	}).([]byte)

	director, err := director.NewClient(director.Credentials{
		Username: config.DirectorUsername,
		Password: config.DirectorPassword,
		Host:     directorPublicIP,
		CACert:   config.DirectorCACert,
	}, v)
	if err != nil {
		return nil, err
	}

	return client.boshClientFactory(
		config,
		metadata,
		director,
		client.stdout,
		client.stderr,
		client.provider,
	)
}

func awsInputVarsMapFromConfig(c config.Config) map[string]interface{} {
	return map[string]interface{}{
		"NetworkCIDR":            c.NetworkCIDR,
		"PublicCIDR":             c.PublicCIDR,
		"PrivateCIDR":            c.PrivateCIDR,
		"AllowIPs":               c.AllowIPs,
		"AvailabilityZone":       c.AvailabilityZone,
		"ConfigBucket":           c.ConfigBucket,
		"Deployment":             c.Deployment,
		"HostedZoneID":           c.HostedZoneID,
		"HostedZoneRecordPrefix": c.HostedZoneRecordPrefix,
		"Namespace":              c.Namespace,
		"Project":                c.Project,
		"PublicKey":              c.PublicKey,
		"RDSDefaultDatabaseName": c.RDSDefaultDatabaseName,
		"RDSInstanceClass":       c.RDSInstanceClass,
		"RDSPassword":            c.RDSPassword,
		"RDSUsername":            c.RDSUsername,
		"Region":                 c.Region,
		"SourceAccessIP":         c.SourceAccessIP,
		"TFStatePath":            c.TFStatePath,
		"MultiAZRDS":             c.MultiAZRDS,
	}
}
