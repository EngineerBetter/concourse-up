package commands

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/EngineerBetter/concourse-up/terraform"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/util"

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
		Usage:       "(optional) Size of Concourse workers. Can be medium, large, xlarge, 2xlarge, 4xlarge, 12xlarge or 24xlarge",
		EnvVar:      "WORKER_SIZE",
		Value:       "xlarge",
		Destination: &deployArgs.WorkerSize,
	},
	cli.StringFlag{
		Name:        "worker-type",
		Usage:       "(optional) Specify a worker type for aws (m5 or m4)",
		EnvVar:      "WORKER_TYPE",
		Value:       "m4",
		Destination: &deployArgs.WorkerType,
	},
	cli.StringFlag{
		Name:        "web-size",
		Usage:       "(optional) Size of Concourse web node. Can be small, medium, large, xlarge, 2xlarge",
		EnvVar:      "WEB_SIZE",
		Value:       "small",
		Destination: &deployArgs.WebSize,
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
		Name:        "self-update",
		Usage:       "(optional) Causes Concourse-up to exit as soon as the BOSH deployment starts. May only be used when upgrading an existing deployment",
		EnvVar:      "SELF_UPDATE",
		Hidden:      true,
		Destination: &deployArgs.SelfUpdate,
	},
	cli.StringFlag{
		Name:        "db-size",
		Usage:       "(optional) Size of Concourse RDS instance. Can be small, medium, large, xlarge, 2xlarge, or 4xlarge",
		EnvVar:      "DB_SIZE",
		Value:       "small",
		Destination: &deployArgs.DBSize,
	},
	cli.BoolTFlag{
		Name:        "spot",
		Usage:       "(optional) Use spot instances for workers. Can be true/false (default: true)",
		EnvVar:      "SPOT",
		Destination: &deployArgs.Spot,
	},
	cli.StringFlag{
		Name:        "allow-ips",
		Usage:       "(optional) Comma separated list of IP addresses or CIDR ranges to allow access to",
		EnvVar:      "ALLOW_IPS",
		Value:       "0.0.0.0/0",
		Destination: &deployArgs.AllowIPs,
	},
	cli.StringFlag{
		Name:        "github-auth-client-id",
		Usage:       "(optional) Client ID for a github OAuth application - Used for Github Auth",
		EnvVar:      "GITHUB_AUTH_CLIENT_ID",
		Destination: &deployArgs.GithubAuthClientID,
	},
	cli.StringFlag{
		Name:        "github-auth-client-secret",
		Usage:       "(optional) Client Secret for a github OAuth application - Used for Github Auth",
		EnvVar:      "GITHUB_AUTH_CLIENT_SECRET",
		Destination: &deployArgs.GithubAuthClientSecret,
	},
	cli.StringSliceFlag{
		Name:  "add-tag",
		Usage: "(optional) Key=Value pair to tag EC2 instances with - Multiple tags can be applied with multiple uses of this flag",
		Value: &deployArgs.Tags,
	},
	cli.StringFlag{
		Name:        "namespace",
		Usage:       "(optional) Specify a namespace for deployments in order to group them in a meaningful way",
		EnvVar:      "NAMESPACE",
		Destination: &deployArgs.Namespace,
	},
	cli.StringFlag{
		Name:        "zone",
		Usage:       "(optional) Specify an availability zone",
		EnvVar:      "ZONE",
		Destination: &deployArgs.Zone,
	},
}

func regionFromZone(zone string) (string, string) {
	re := regexp.MustCompile(`(?m)^\w+-\w+-\d`)
	regionFound := re.FindString(zone)
	if regionFound != "" {
		return regionFound, fmt.Sprintf("No region provided, please note that your zone will be paired with a matching region.\nThis region: %s is used for deployment.\n", regionFound)
	}
	return "", ""
}

func zoneBelongsToRegion(zone, region string) error {
	if !strings.Contains(zone, region) {
		return fmt.Errorf("The region and the zones provided do not match. Please note that the zone %s needs to be within a %s region", zone, region)
	}
	return nil
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

		err := deployArgs.MarkSetFlags(c)
		if err != nil {
			return err
		}

		deployArgs.GithubAuthIsSet = c.IsSet("github-auth-client-id") && c.IsSet("github-auth-client-secret")
		if err = deployArgs.Validate(); err != nil {
			return err
		}

		if deployArgs.ZoneIsSet && deployArgs.AWSRegionIsSet {
			if err1 := zoneBelongsToRegion(deployArgs.Zone, deployArgs.AWSRegion); err1 != nil {
				return err1
			}
		}

		if deployArgs.ZoneIsSet && !deployArgs.AWSRegionIsSet {
			region, message := regionFromZone(deployArgs.Zone)
			if region != "" {
				deployArgs.AWSRegion = region
				fmt.Print(message)
			}
		}

		awsClient, err := iaas.New(deployArgs.IAAS, deployArgs.AWSRegion)
		if err != nil {
			return err
		}
		terraformClient, err := terraform.New(terraform.DownloadTerraform())
		if err != nil {
			return err
		}
		client := concourse.NewClient(
			awsClient,
			terraformClient,
			bosh.NewClient,
			fly.New,
			certs.Generate,
			config.New(awsClient, name, deployArgs.Namespace),

			&deployArgs,
			os.Stdout,
			os.Stderr,
			util.FindUserIP,
			certs.NewAcmeClient,
			c.App.Version,
		)

		return client.Deploy()
	},
}
