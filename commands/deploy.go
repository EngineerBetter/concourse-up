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

var awsRegion string
var domain string

var deployFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Value:       "eu-west-1",
		Usage:       "AWS region",
		EnvVar:      "AWS_REGION",
		Destination: &awsRegion,
	},
	cli.StringFlag{
		Name:        "domain",
		Usage:       "Domain to use as endpoint for Concourse web interface (eg: ci.myproject.com)",
		EnvVar:      "DOMAIN",
		Destination: &domain,
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

		client := concourse.NewClient(
			terraform.NewClient,
			bosh.NewClient,
			certs.Generate,
			aws.FindLongestMatchingHostedZone,
			&config.Client{Project: name},
			map[string]string{
				"aws-region": awsRegion,
				"domain":     domain,
			},
			os.Stdout,
			os.Stderr,
		)

		return client.Deploy()
	},
}
