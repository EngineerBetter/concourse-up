package commands

import (
	"errors"
	"os"

	"github.com/engineerbetter/concourse-up/aws"
	"github.com/engineerbetter/concourse-up/bosh"
	"github.com/engineerbetter/concourse-up/certs"
	"github.com/engineerbetter/concourse-up/concourse"
	"github.com/engineerbetter/concourse-up/config"
	"github.com/engineerbetter/concourse-up/terraform"

	"gopkg.in/urfave/cli.v1"
)

var deployArgs config.DeployArgs

var deployFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Value:       "eu-west-1",
		Usage:       "(optional) AWS region",
		EnvVar:      "AWS_REGION",
		Destination: &deployArgs.AWSRegion,
	},
	cli.StringFlag{
		Name:        "domain",
		Usage:       "Domain to use as endpoint for Concourse web interface (eg: ci.myproject.com)",
		EnvVar:      "DOMAIN",
		Destination: &deployArgs.Domain,
	},
	cli.StringFlag{
		Name:        "tls-cert",
		Usage:       "(optional) TLS cert to use with Concourse endpoint",
		EnvVar:      "TLS_CERT",
		Destination: &deployArgs.TLSCert,
	},
	cli.StringFlag{
		Name:        "tls-key",
		Usage:       "(optional) TLS private key to use with Concourse endpoint",
		EnvVar:      "TLS_KEY",
		Destination: &deployArgs.TLSKey,
	},
	cli.IntFlag{
		Name:        "workers",
		Usage:       "Number of Concourse worker instances to deploy",
		EnvVar:      "TLS_KEY",
		Value:       1,
		Destination: &deployArgs.WorkerCount,
	},
}

var deploy = cli.Command{
	Name:      "deploy",
	Aliases:   []string{"d"},
	Usage:     "Deploys or updates a Concourse",
	ArgsUsage: "<name>",
	Flags:     deployFlags,
	Action: func(c *cli.Context) error {
		name := c.Args().Get(0)
		if name == "" {
			return errors.New("Usage is `concourse-up deploy <name>`")
		}

		if err := deployArgs.Validate(); err != nil {
			return err
		}

		client := concourse.NewClient(
			terraform.NewClient,
			bosh.NewClient,
			certs.Generate,
			aws.FindLongestMatchingHostedZone,
			&config.Client{Project: name},
			&deployArgs,
			os.Stdout,
			os.Stderr,
		)

		return client.Deploy()
	},
}
