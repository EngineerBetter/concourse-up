package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"

	"encoding/json"

	"gopkg.in/urfave/cli.v1"
)

var infoArgs config.InfoArgs

var infoFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Value:       "eu-west-1",
		Usage:       "(optional) AWS region",
		EnvVar:      "AWS_REGION",
		Destination: &infoArgs.AWSRegion,
	},
	cli.BoolFlag{
		Name:        "json",
		Usage:       "(optional) Output as json",
		EnvVar:      "JSON",
		Destination: &infoArgs.JSON,
	},
	cli.BoolFlag{
		Name:        "env",
		Usage:       "(optional) Output environment variables",
		Destination: &infoArgs.Env,
	},
	cli.StringFlag{
		Name:        "iaas",
		Usage:       "(optional) IAAS, can be AWS or GCP",
		EnvVar:      "IAAS",
		Value:       "AWS",
		Hidden:      true,
		Destination: &infoArgs.IAAS,
	},
}

var info = cli.Command{
	Name:      "info",
	Aliases:   []string{"i"},
	Usage:     "Fetches information on a deployed environment",
	ArgsUsage: "<name>",
	Flags:     infoFlags,
	Action: func(c *cli.Context) error {
		name := c.Args().Get(0)
		if name == "" {
			return errors.New("Usage is `concourse-up info <name>`")
		}

		awsClient, err := iaas.New(infoArgs.IAAS, infoArgs.AWSRegion)
		if err != nil {
			return err
		}

		acmeClient, err := certs.NewAcmeClient()
		if err != nil {
			return err
		}

		client := concourse.NewClient(
			awsClient,
			terraform.NewClient,
			bosh.NewClient,
			fly.New,
			certs.Generate,
			config.New(awsClient, name),
			nil,
			os.Stdout,
			os.Stderr,
			util.FindUserIP,
			acmeClient,
		)

		info, err := client.FetchInfo()
		if err != nil {
			return err
		}

		switch {
		case infoArgs.JSON:
			return json.NewEncoder(os.Stdout).Encode(info)
		case infoArgs.Env:
			env, err := info.Env()
			if err != nil {
				return err
			}
			_, err = os.Stdout.WriteString(env)
			return err
		default:
			_, err := fmt.Fprint(os.Stdout, info)
			return err
		}
	},
}
