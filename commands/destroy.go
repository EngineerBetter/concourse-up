package commands

import (
	"errors"
	"fmt"
	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/commands/destroy"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"
	"os"

	"gopkg.in/urfave/cli.v1"
)

const awsConst = "AWS"
const gcpConst = "GCP"

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
		Destination: &initialDestroyArgs.IAAS,
	},
	cli.StringFlag{
		Name:        "namespace",
		Usage:       "(optional) Specify a namespace for deployments in order to group them in a meaningful way",
		EnvVar:      "NAMESPACE",
		Destination: &initialDestroyArgs.Namespace,
	},
}

func destroyAction(c *cli.Context, destroyArgs destroy.Args, provider iaas.Provider) error {
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
	client, err := buildDestroyClient(name, version, destroyArgs, provider)
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

func buildDestroyClient(name, version string, destroyArgs destroy.Args, provider iaas.Provider) (*concourse.Client, error) {
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
		config.New(provider, name, destroyArgs.Namespace),
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
		iaasName, err := iaas.Assosiate(initialDestroyArgs.IAAS)
		if err != nil {
			return err
		}
		provider, err := iaas.New(iaasName, initialDestroyArgs.AWSRegion)
		if err != nil {
			return fmt.Errorf("Error creating IAAS provider on destroy: [%v]", err)
		}
		return destroyAction(c, initialDestroyArgs, provider)
	},
}
