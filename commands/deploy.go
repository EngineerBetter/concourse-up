package commands

import (
	"errors"
	"os"

	"bitbucket.org/engineerbetter/concourse-up/certs"
	"bitbucket.org/engineerbetter/concourse-up/concourse"
	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/director"
	"bitbucket.org/engineerbetter/concourse-up/terraform"

	"gopkg.in/urfave/cli.v1"
)

var awsRegion string

var deployFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Value:       "eu-west-1",
		Usage:       "AWS region",
		EnvVar:      "AWS_REGION",
		Destination: &awsRegion,
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
			director.NewBoshInitClient,
			certs.Generate,
			&config.Client{Project: name},
			os.Stdout,
			os.Stderr,
		)

		return client.Deploy()
	},
}
