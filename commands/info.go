package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/commands/info"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"
	"gopkg.in/urfave/cli.v1"
	"os"
)

var initialInfoArgs info.Args

var infoFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Usage:       "(optional) AWS region",
		EnvVar:      "AWS_REGION",
		Destination: &initialInfoArgs.AWSRegion,
	},
	cli.BoolFlag{
		Name:        "json",
		Usage:       "(optional) Output as json",
		EnvVar:      "JSON",
		Destination: &initialInfoArgs.JSON,
	},
	cli.BoolFlag{
		Name:        "env",
		Usage:       "(optional) Output environment variables",
		Destination: &initialInfoArgs.Env,
	},
	cli.BoolFlag{
		Name:        "cert-expiry",
		Usage:       "(optional) Output only the expiration date of the director nats certificate",
		Destination: &initialInfoArgs.CertExpiry,
	},
	cli.StringFlag{
		Name:        "iaas",
		Usage:       "(optional) IAAS, can be AWS or GCP",
		EnvVar:      "IAAS",
		Value:       "AWS",
		Destination: &initialInfoArgs.IAAS,
	},
	cli.StringFlag{
		Name:        "namespace",
		Usage:       "(optional) Specify a namespace for deployments in order to group them in a meaningful way",
		EnvVar:      "NAMESPACE",
		Destination: &initialInfoArgs.Namespace,
	},
}

func infoAction(c *cli.Context, infoArgs info.Args, provider iaas.Provider) error {
	name := c.Args().Get(0)
	if name == "" {
		return errors.New("Usage is `concourse-up info <name>`")
	}

	version := c.App.Version

	err := infoArgs.MarkSetFlags(c)
	if err != nil {
		return err
	}

	client, err := buildInfoClient(name, version, infoArgs, provider)
	if err != nil {
		return err
	}
	i, err := client.FetchInfo()
	if err != nil {
		return err
	}
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
	case infoArgs.CertExpiry:
		os.Stdout.WriteString(i.CertExpiry)
		return nil
	default:
		_, err := fmt.Fprint(os.Stdout, i)
		return err
	}
	//this will never run
	return nil
}

func buildInfoClient(name, version string, infoArgs info.Args, provider iaas.Provider) (*concourse.Client, error) {
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
		config.New(provider, name, infoArgs.Namespace),
		nil,
		os.Stdout,
		os.Stderr,
		util.FindUserIP,
		certs.NewAcmeClient,
		util.GeneratePasswordWithLength,
		util.EightRandomLetters,
		util.GenerateSSHKeyPair,
		version,
	)

	return client, nil
}

var infoCmd = cli.Command{
	Name:      "info",
	Aliases:   []string{"i"},
	Usage:     "Fetches information on a deployed environment",
	ArgsUsage: "<name>",
	Flags:     infoFlags,
	Action: func(c *cli.Context) error {
		iaasName, err := iaas.Assosiate(initialInfoArgs.IAAS)
		if err != nil {
			return err
		}
		provider, err := iaas.New(iaasName, initialInfoArgs.AWSRegion)
		if err != nil {
			return fmt.Errorf("Error creating IAAS provider on info: [%v]", err)
		}
		return infoAction(c, initialInfoArgs, provider)
	},
}
