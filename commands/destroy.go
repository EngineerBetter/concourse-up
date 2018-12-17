package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/commands/destroy"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"

	cli "gopkg.in/urfave/cli.v1"
)

//var destroyArgs config.DestroyArgs
var initialDestroyArgs destroy.Args

var destroyFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Usage:       "(optional) AWS region",
		EnvVar:      "AWS_REGION",
		Destination: &initialDestroyArgs.AWSRegion,
	},
	cli.StringFlag{
		Name:        "iaas",
		Usage:       "(optional) IAAS, can be AWS or GCP",
		EnvVar:      "IAAS",
		Value:       "AWS",
		Hidden:      true,
		Destination: &initialDestroyArgs.IAAS,
	},
	cli.StringFlag{
		Name:        "namespace",
		Usage:       "(optional) Specify a namespace for deployments in order to group them in a meaningful way",
		EnvVar:      "NAMESPACE",
		Destination: &initialDestroyArgs.Namespace,
	},
}

func destroyAction(c *cli.Context, destroyArgs destroy.Args, iaasFactory iaas.Factory) error {
	name := c.Args().Get(0)
	if name == "" {
		return errors.New("Usage is `concourse-up destroy <name>`")
	}

	if !NonInteractiveModeEnabled() {
		confirm, err := util.CheckConfirmation(os.Stdin, os.Stdout, name)
		if err != nil {
			return err
		}

		if !confirm {
			fmt.Println("Bailing out...")
			return nil
		}
	}

	version := c.App.Version

	destroyArgs, err := markSetFlags(c, destroyArgs)
	if err != nil {
		return err
	}
	destroyArgs = setRegion(destroyArgs)
	client, err := buildDestroyClient(name, version, destroyArgs, iaasFactory)
	if err != nil {
		return err
	}
	return client.Destroy()
}
func markSetFlags(c *cli.Context, destroyArgs destroy.Args) (destroy.Args, error) {
	err := destroyArgs.MarkSetFlags(c)
	if err != nil {
		return destroyArgs, err
	}
	return destroyArgs, nil
}
func setRegion(destroyArgs destroy.Args) destroy.Args {

	if !destroyArgs.AWSRegionIsSet {
		switch strings.ToUpper(destroyArgs.IAAS) {
		case "AWS": //nolint
			destroyArgs.AWSRegion = "eu-west-1"
		case "GCP": //nolint
			destroyArgs.AWSRegion = "europe-west1"
		}
	}
	return destroyArgs
}
func buildDestroyClient(name, version string, destroyArgs destroy.Args, iaasFactory iaas.Factory) (*concourse.Client, error) {
	awsClient, err := iaasFactory(destroyArgs.IAAS, destroyArgs.AWSRegion)
	if err != nil {
		return nil, err
	}
	terraformClient, err := terraform.New(terraform.DownloadTerraform())
	if err != nil {
		return nil, err
	}
	client := concourse.NewClient(
		awsClient,
		terraformClient,
		bosh.New,
		fly.New,
		certs.Generate,
		config.New(awsClient, name, destroyArgs.Namespace),
		nil,
		os.Stdout,
		os.Stderr,
		util.FindUserIP,
		certs.NewAcmeClient,
		version,
	)

	return client, nil
}

var destroyCmd = cli.Command{
	Name:      "destroy",
	Aliases:   []string{"x"},
	Usage:     "Destroys a Concourse",
	ArgsUsage: "<name>",
	Flags:     destroyFlags,
	Action: func(c *cli.Context) error {
		return destroyAction(c, initialDestroyArgs, iaas.New)
	},
}
