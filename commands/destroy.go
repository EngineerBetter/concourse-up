package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/EngineerBetter/concourse-up/aws"
	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"

	"gopkg.in/urfave/cli.v1"
)

var destroyArgs config.DestroyArgs

var destroyFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Value:       "eu-west-1",
		Usage:       "(optional) AWS region",
		EnvVar:      "AWS_REGION",
		Destination: &destroyArgs.AWSRegion,
	},
}

var destroy = cli.Command{
	Name:      "destroy",
	Aliases:   []string{"x"},
	Usage:     "Destroys a Concourse",
	ArgsUsage: "<name>",
	Flags:     destroyFlags,
	Action: func(c *cli.Context) error {
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

		awsClient := &aws.Client{}

		client := concourse.NewClient(
			awsClient,
			terraform.NewClient,
			bosh.NewClient,
			fly.New,
			certs.Generate,
			config.New(awsClient, name, deployArgs.AWSRegion),
			nil,
			os.Stdout,
			os.Stderr,
		)

		return client.Destroy()
	},
}
