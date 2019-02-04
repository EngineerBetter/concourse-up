package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/commands/info"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"
	cli "gopkg.in/urfave/cli.v1"
)

var initialInfoArgs info.Args
var awsClient iaas.Provider

var infoFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Usage:       "(optional) AWS region",
		EnvVar:      "AWS_REGION",
		Destination: &initialInfoArgs.AWSRegion,
	},
	cli.BoolFlag{
		Name:        "json",
		Usage:       "(optional) Output as json",
		EnvVar:      "JSON",
		Destination: &initialInfoArgs.JSON,
	},
	cli.BoolFlag{
		Name:        "env",
		Usage:       "(optional) Output environment variables",
		Destination: &initialInfoArgs.Env,
	},
	cli.BoolFlag{
		Name:        "cert-expiry",
		Usage:       "(optional) Output only the expiration date of the director nats certificate",
		Destination: &initialInfoArgs.CertExpiry,
	},
	cli.StringFlag{
		Name:        "iaas",
		Usage:       "(optional) IAAS, can be AWS or GCP",
		EnvVar:      "IAAS",
		Value:       "AWS",
		Destination: &initialInfoArgs.IAAS,
	},
	cli.StringFlag{
		Name:        "namespace",
		Usage:       "(optional) Specify a namespace for deployments in order to group them in a meaningful way",
		EnvVar:      "NAMESPACE",
		Destination: &initialInfoArgs.Namespace,
	},
}

func infoAction(c *cli.Context, infoArgs info.Args, iaasFactory iaas.Factory) error {
	name := c.Args().Get(0)
	if name == "" {
		return errors.New("Usage is `concourse-up info <name>`")
	}

	version := c.App.Version

	err := infoArgs.MarkSetFlags(c)
	if err != nil {
		return err
	}

	infoArgs = setInfoRegion(infoArgs)
	client, err := buildInfoClient(name, version, infoArgs, iaasFactory)
	if err != nil {
		return err
	}
	i, err := client.FetchInfo()
	if err != nil {
		return err
	}
	switch {
	case infoArgs.JSON:
		return json.NewEncoder(os.Stdout).Encode(i)
	case infoArgs.Env:
		env, err := i.Env()
		if err != nil {
			return err
		}
		_, err = os.Stdout.WriteString(env)
		return err
	case infoArgs.CertExpiry:
		os.Stdout.WriteString(i.CertExpiry)
		return nil
	default:
		_, err := fmt.Fprint(os.Stdout, i)
		return err
	}
	//this will never run
	return nil
}

func setInfoRegion(infoArgs info.Args) info.Args {

	if !infoArgs.AWSRegionIsSet {
		switch strings.ToUpper(infoArgs.IAAS) {
		case awsConst: //nolint
			infoArgs.AWSRegion = "eu-west-1" //nolint
		case gcpConst: //nolint
			infoArgs.AWSRegion = "europe-west1" //nolint
		}
	}
	return infoArgs
}

func buildInfoClient(name, version string, infoArgs info.Args, iaasFactory iaas.Factory) (*concourse.Client, error) {
	awsClient, err := iaasFactory(infoArgs.IAAS, infoArgs.AWSRegion)
	if err != nil {
		return nil, err
	}
	terraformClient, err := terraform.New(terraform.DownloadTerraform())
	if err != nil {
		return nil, err
	}

	maintainer, err := concourse.NewMaintainer(infoArgs.IAAS)
	if err != nil {
		return nil, err
	}

	client := concourse.NewClient(
		awsClient,
		terraformClient,
		bosh.New,
		fly.New,
		certs.Generate,
		config.New(awsClient, name, infoArgs.Namespace),
		nil,
		os.Stdout,
		os.Stderr,
		util.FindUserIP,
		certs.NewAcmeClient,
		version,
		maintainer,
	)

	return client, nil
}

var infoCmd = cli.Command{
	Name:      "info",
	Aliases:   []string{"i"},
	Usage:     "Fetches information on a deployed environment",
	ArgsUsage: "<name>",
	Flags:     infoFlags,
	Action: func(c *cli.Context) error {
		return infoAction(c, initialInfoArgs, iaas.New)
	},
}
