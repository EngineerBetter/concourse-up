package commands

import (
	"errors"
	"os"

	"github.com/EngineerBetter/concourse-up/aws"
	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/terraform"

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
		Usage:       "(optional) Domain to use as endpoint for Concourse web interface (eg: ci.myproject.com)",
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
		Usage:       "(optional) Number of Concourse worker instances to deploy",
		EnvVar:      "WORKERS",
		Value:       1,
		Destination: &deployArgs.WorkerCount,
	},
	cli.StringFlag{
		Name:        "worker-size",
		Usage:       "(optional) Size of Concourse workers. Can be medium, large or xlarge",
		EnvVar:      "WORKER_SIZE",
		Value:       "xlarge",
		Destination: &deployArgs.WorkerSize,
	},
	cli.StringFlag{
		Name:        "iaas",
		Usage:       "(optional) IAAS, can be AWS or GCP",
		EnvVar:      "IAAS",
		Value:       "AWS",
		Hidden:      true,
		Destination: &deployArgs.IAAS,
	},
	cli.BoolFlag{
		Name:        "detach-bosh-deployment",
		Usage:       "(optional) Causes Concourse-up to exit as soon as the BOSH deployment starts. May only be used when upgrading an existing deplomnt",
		EnvVar:      "DETACH_BOSH_DEPLOYMENT",
		Hidden:      true,
		Destination: &deployArgs.DetachBoshDeployment,
	},
	cli.BoolFlag{
		Name:        "pause-self-update",
		Usage:       "(optional) Keeps the self-update job paused",
		EnvVar:      "DETACH_BOSH_DEPLOYMENT",
		Hidden:      true,
		Destination: &deployArgs.PauseSelfUpdate,
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

		awsClient := &aws.Client{}

		client := concourse.NewClient(
			awsClient,
			terraform.NewClient,
			bosh.NewClient,
			fly.New,
			certs.Generate,
			&config.Client{Project: name, S3Region: deployArgs.AWSRegion},
			&deployArgs,
			os.Stdout,
			os.Stderr,
		)

		return client.Deploy()
	},
}
