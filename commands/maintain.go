package commands

import (
	"errors"
	"os"
	"strings"

	"github.com/EngineerBetter/concourse-up/commands/maintain"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"
	cli "gopkg.in/urfave/cli.v1"
)

var initialMaintainArgs maintain.Args
var provider iaas.Provider

var maintainFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Usage:       "(optional) AWS region",
		EnvVar:      "AWS_REGION",
		Destination: &initialMaintainArgs.AWSRegion,
	},
	cli.BoolFlag{
		Name:        "renew-nats-cert",
		Usage:       "(optional) Rotate nats certificate",
		Destination: &initialMaintainArgs.RenewNatsCert,
	},
	cli.StringFlag{
		Name:        "iaas",
		Usage:       "(optional) IAAS, can be AWS or GCP",
		EnvVar:      "IAAS",
		Value:       "AWS",
		Hidden:      true,
		Destination: &initialMaintainArgs.IAAS,
	},
	cli.StringFlag{
		Name:        "namespace",
		Usage:       "(optional) Specify a namespace for deployments in order to group them in a meaningful way",
		EnvVar:      "NAMESPACE",
		Destination: &initialMaintainArgs.Namespace,
	},
	cli.IntFlag{
		Name:        "state",
		Usage:       "(optional) Set the desired state for nats rotation tasks",
		EnvVar:      "STATE",
		Destination: &initialMaintainArgs.State,
	},
}

func maintainAction(c *cli.Context, maintainArgs maintain.Args, iaasFactory iaas.Factory) error {
	name := c.Args().Get(0)
	if name == "" {
		return errors.New("Usage is `concourse-up maintain <name>`")
	}

	version := c.App.Version

	err := maintainArgs.MarkSetFlags(c)
	if err != nil {
		return err
	}

	maintainArgs = setMaintainRegion(maintainArgs)
	client, err := buildMaintainClient(name, version, maintainArgs, iaasFactory)
	if err != nil {
		return err
	}
	err = client.Maintain(maintainArgs)
	if err != nil {
		return err
	}
	//this will never run
	return nil
}

func setMaintainRegion(maintainArgs maintain.Args) maintain.Args {

	if !maintainArgs.AWSRegionIsSet {
		switch strings.ToUpper(maintainArgs.IAAS) {
		case awsConst: //nolint
			maintainArgs.AWSRegion = "eu-west-1" //nolint
		case gcpConst: //nolint
			maintainArgs.AWSRegion = "europe-west1" //nolint
		}
	}
	return maintainArgs
}

func buildMaintainClient(name, version string, maintainArgs maintain.Args, iaasFactory iaas.Factory) (*concourse.Client, error) {
	provider, err := iaasFactory(maintainArgs.IAAS, maintainArgs.AWSRegion)
	if err != nil {
		return nil, err
	}
	terraformClient, err := terraform.New(terraform.DownloadTerraform())
	if err != nil {
		return nil, err
	}
	client := concourse.NewClient(
		provider,
		terraformClient,
		bosh.New,
		fly.New,
		certs.Generate,
		config.New(provider, name, maintainArgs.Namespace),
		nil,
		os.Stdout,
		os.Stderr,
		util.FindUserIP,
		certs.NewAcmeClient,
		version,
	)

	return client, nil
}

var maintainCmd = cli.Command{
	Name:      "maintain",
	Aliases:   []string{"m"},
	Usage:     "Handles maintainenace operations in concourse-up",
	ArgsUsage: "<name>",
	Flags:     maintainFlags,
	Action: func(c *cli.Context) error {
		return maintainAction(c, initialMaintainArgs, iaas.New)
	},
}
