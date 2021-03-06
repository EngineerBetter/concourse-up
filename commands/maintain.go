package commands

import (
	"errors"
	"fmt"
	"github.com/EngineerBetter/concourse-up/commands/maintain"
	"os"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"
	"gopkg.in/urfave/cli.v1"
)

var initialMaintainArgs maintain.Args
var provider iaas.Provider

var maintainFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Usage:       "(optional) AWS region",
		EnvVar:      "AWS_REGION",
		Destination: &initialMaintainArgs.Region,
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
		Destination: &initialMaintainArgs.IAAS,
	},
	cli.StringFlag{
		Name:        "namespace",
		Usage:       "(optional) Specify a namespace for deployments in order to group them in a meaningful way",
		EnvVar:      "NAMESPACE",
		Destination: &initialMaintainArgs.Namespace,
	},
	cli.IntFlag{
		Name:        "stage",
		Usage:       "(optional) Set the desired stage for nats rotation tasks",
		EnvVar:      "STAGE",
		Destination: &initialMaintainArgs.Stage,
	},
}

func maintainAction(c *cli.Context, maintainArgs maintain.Args, provider iaas.Provider) error {
	name := c.Args().Get(0)
	if name == "" {
		return errors.New("Usage is `concourse-up maintain <name>`")
	}

	version := c.App.Version

	err := maintainArgs.MarkSetFlags(c)
	if err != nil {
		return err
	}

	client, err := buildMaintainClient(name, version, maintainArgs, provider)
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

func buildMaintainClient(name, version string, maintainArgs maintain.Args, provider iaas.Provider) (*concourse.Client, error) {
	terraformClient, err := terraform.New(provider.IAAS(), terraform.DownloadTerraform())
	if err != nil {
		return nil, err
	}

	tfInputVarsFactory, err := concourse.NewTFInputVarsFactory(provider)
	if err != nil {
		return nil, fmt.Errorf("Error creating TFInputVarsFactory [%v]", err)
	}

	client := concourse.NewClient(
		provider,
		terraformClient,
		tfInputVarsFactory,
		bosh.New,
		fly.New,
		certs.Generate,
		config.New(provider, name, maintainArgs.Namespace),
		nil,
		os.Stdout,
		os.Stderr,
		util.FindUserIP,
		certs.NewAcmeClient,
		util.GeneratePasswordWithLength,
		util.EightRandomLetters,
		util.GenerateSSHKeyPair,
		version,
	)

	return client, nil
}

var maintainCmd = cli.Command{
	Name:      "maintain",
	Aliases:   []string{"m"},
	Usage:     "Handles maintenance operations in concourse-up",
	ArgsUsage: "<name>",
	Flags:     maintainFlags,
	Action: func(c *cli.Context) error {
		iaasName, err := iaas.Assosiate(initialMaintainArgs.IAAS)
		if err != nil {
			return err
		}
		provider, err := iaas.New(iaasName, initialMaintainArgs.Region)
		if err != nil {
			return fmt.Errorf("Error creating IAAS provider on maintain: [%v]", err)
		}

		return maintainAction(c, initialMaintainArgs, provider)
	},
}
