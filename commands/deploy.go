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
	"github.com/EngineerBetter/concourse-up/commands/deploy"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/util"

	cli "gopkg.in/urfave/cli.v1"
)

var initialDeployArgs deploy.Args

var deployFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Usage:       "(optional) AWS region",
		EnvVar:      "AWS_REGION",
		Destination: &initialDeployArgs.AWSRegion,
	},
	cli.StringFlag{
		Name:        "domain",
		Usage:       "(optional) Domain to use as endpoint for Concourse web interface (eg: ci.myproject.com)",
		EnvVar:      "DOMAIN",
		Destination: &initialDeployArgs.Domain,
	},
	cli.StringFlag{
		Name:        "tls-cert",
		Usage:       "(optional) TLS cert to use with Concourse endpoint",
		EnvVar:      "TLS_CERT",
		Destination: &initialDeployArgs.TLSCert,
	},
	cli.StringFlag{
		Name:        "tls-key",
		Usage:       "(optional) TLS private key to use with Concourse endpoint",
		EnvVar:      "TLS_KEY",
		Destination: &initialDeployArgs.TLSKey,
	},
	cli.IntFlag{
		Name:        "workers",
		Usage:       "(optional) Number of Concourse worker instances to deploy",
		EnvVar:      "WORKERS",
		Value:       1,
		Destination: &initialDeployArgs.WorkerCount,
	},
	cli.StringFlag{
		Name:        "worker-size",
		Usage:       "(optional) Size of Concourse workers. Can be medium, large, xlarge, 2xlarge, 4xlarge, 12xlarge or 24xlarge",
		EnvVar:      "WORKER_SIZE",
		Value:       "xlarge",
		Destination: &initialDeployArgs.WorkerSize,
	},
	cli.StringFlag{
		Name:        "worker-type",
		Usage:       "(optional) Specify a worker type for aws (m5 or m4)",
		EnvVar:      "WORKER_TYPE",
		Value:       "m4",
		Destination: &initialDeployArgs.WorkerType,
	},
	cli.StringFlag{
		Name:        "web-size",
		Usage:       "(optional) Size of Concourse web node. Can be small, medium, large, xlarge, 2xlarge",
		EnvVar:      "WEB_SIZE",
		Value:       "small",
		Destination: &initialDeployArgs.WebSize,
	},
	cli.StringFlag{
		Name:        "iaas",
		Usage:       "(optional) IAAS, can be AWS or GCP",
		EnvVar:      "IAAS",
		Value:       "AWS",
		Hidden:      true,
		Destination: &initialDeployArgs.IAAS,
	},
	cli.BoolFlag{
		Name:        "self-update",
		Usage:       "(optional) Causes Concourse-up to exit as soon as the BOSH deployment starts. May only be used when upgrading an existing deployment",
		EnvVar:      "SELF_UPDATE",
		Hidden:      true,
		Destination: &initialDeployArgs.SelfUpdate,
	},
	cli.StringFlag{
		Name:        "db-size",
		Usage:       "(optional) Size of Concourse RDS instance. Can be small, medium, large, xlarge, 2xlarge, or 4xlarge",
		EnvVar:      "DB_SIZE",
		Value:       "small",
		Destination: &initialDeployArgs.DBSize,
	},
	cli.BoolTFlag{
		Name:        "spot",
		Usage:       "(optional) Use spot instances for workers. Can be true/false (default: true)",
		EnvVar:      "SPOT",
		Destination: &initialDeployArgs.Spot,
	},
	cli.StringFlag{
		Name:        "allow-ips",
		Usage:       "(optional) Comma separated list of IP addresses or CIDR ranges to allow access to",
		EnvVar:      "ALLOW_IPS",
		Value:       "0.0.0.0/0",
		Destination: &initialDeployArgs.AllowIPs,
	},
	cli.StringFlag{
		Name:        "github-auth-client-id",
		Usage:       "(optional) Client ID for a github OAuth application - Used for Github Auth",
		EnvVar:      "GITHUB_AUTH_CLIENT_ID",
		Destination: &initialDeployArgs.GithubAuthClientID,
	},
	cli.StringFlag{
		Name:        "github-auth-client-secret",
		Usage:       "(optional) Client Secret for a github OAuth application - Used for Github Auth",
		EnvVar:      "GITHUB_AUTH_CLIENT_SECRET",
		Destination: &initialDeployArgs.GithubAuthClientSecret,
	},
	cli.StringSliceFlag{
		Name:  "add-tag",
		Usage: "(optional) Key=Value pair to tag EC2 instances with - Multiple tags can be applied with multiple uses of this flag",
		Value: &initialDeployArgs.Tags,
	},
	cli.StringFlag{
		Name:        "namespace",
		Usage:       "(optional) Specify a namespace for deployments in order to group them in a meaningful way",
		EnvVar:      "NAMESPACE",
		Destination: &initialDeployArgs.Namespace,
	},
	cli.StringFlag{
		Name:        "zone",
		Usage:       "(optional) Specify an availability zone",
		EnvVar:      "ZONE",
		Destination: &initialDeployArgs.Zone,
	},
}

func deployAction(c *cli.Context, deployArgs deploy.Args, iaasFactory iaas.Factory) error {
	name := c.Args().Get(0)
	if name == "" {
		return errors.New("Usage is `concourse-up deploy <name>`")
	}

	version := c.App.Version

	deployArgs, err := validateDeployArgs(c, deployArgs)
	if err != nil {
		return err
	}

	deployArgs, err = setZoneAndRegion(deployArgs)
	if err != nil {
		return err
	}

	client, err := buildClient(name, version, deployArgs, iaasFactory)
	if err != nil {
		return err
	}

	return client.Deploy()
}

func validateDeployArgs(c *cli.Context, deployArgs deploy.Args) (deploy.Args, error) {
	err := deployArgs.MarkSetFlags(c)
	if err != nil {
		return deployArgs, err
	}

	if err = deployArgs.Validate(); err != nil {
		return deployArgs, err
	}

	return deployArgs, nil
}

func setZoneAndRegion(deployArgs deploy.Args) (deploy.Args, error) {
	if !deployArgs.AWSRegionIsSet {
		switch strings.ToUpper(deployArgs.IAAS) {
		case "AWS": //nolint
			deployArgs.AWSRegion = "eu-west-1"
		case "GCP": //nolint
			deployArgs.AWSRegion = "europe-west1"
		}
	}

	if deployArgs.ZoneIsSet && deployArgs.AWSRegionIsSet {
		if err := zoneBelongsToRegion(deployArgs.Zone, deployArgs.AWSRegion); err != nil {
			return deployArgs, err
		}
	}

	if deployArgs.ZoneIsSet && !deployArgs.AWSRegionIsSet {
		region, message := regionFromZone(deployArgs.Zone)
		if region != "" {
			deployArgs.AWSRegion = region
			fmt.Print(message)
		}
	}

	return deployArgs, nil
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

func buildClient(name, version string, deployArgs deploy.Args, iaasFactory iaas.Factory) (*concourse.Client, error) {
	awsClient, err := iaasFactory(deployArgs.IAAS, deployArgs.AWSRegion)
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
		config.New(awsClient, name, deployArgs.Namespace),
		&deployArgs,
		os.Stdout,
		os.Stderr,
		util.FindUserIP,
		certs.NewAcmeClient,
		version,
	)

	return client, nil
}

var deployCmd = cli.Command{
	Name:      "deploy",
	Aliases:   []string{"d"},
	Usage:     "Deploys or updates a Concourse",
	ArgsUsage: "<name>",
	Flags:     deployFlags,
	Action: func(c *cli.Context) error {
		return deployAction(c, initialDeployArgs, iaas.New)
	},
}
