package commands

import (
	"encoding/json"
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
	cli.StringFlag{
		Name:        "namespace",
		Usage:       "(optional) Specify a namespace for deployments in order to group them in a meaningful way",
		EnvVar:      "NAMESPACE",
		Destination: &infoArgs.Namespace,
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

		provider, err := iaas.New(infoArgs.IAAS, infoArgs.AWSRegion)
		if err != nil {
			return err
		}

		terraformClient, err := terraform.New(terraform.DownloadTerraform())
		if err != nil {
			return err
		}

		client := concourse.NewClient(
			provider,
			terraformClient,
			bosh.New,
			fly.New,

			certs.Generate,
			config.New(provider, name, infoArgs.Namespace),
			nil,
			os.Stdout,
			os.Stderr,
			util.FindUserIP,
			certs.NewAcmeClient,
			c.App.Version,
		)

		i, err := client.FetchInfo()
		if err != nil {
			return err
		}

		// This is temporary and used for discovery of BOSH details
		switch provider.IAAS() {
		case "AWS":
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
			default:
				_, err := fmt.Fprint(os.Stdout, i)
				return err
			}
		case "GCP":
			// DO THE ENV "i.Env()" for GCP in order to be able to connect to the director
			env, err := i.Env()
			if err != nil {
				return err
			}
			_, err = os.Stdout.WriteString(env)
			return err
		}
		return nil
	},
}
