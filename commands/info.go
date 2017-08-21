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
	"github.com/EngineerBetter/concourse-up/terraform"

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
		Usage:       "Output as json",
		EnvVar:      "JSON",
		Destination: &infoArgs.JSON,
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

		logFile, err := os.Create("terraform.combined.log")
		if err != nil {
			return err
		}
		defer logFile.Close()

		client := concourse.NewClient(
			terraform.NewClient,
			bosh.NewClient,
			certs.Generate,
			aws.FindLongestMatchingHostedZone,
			&config.Client{Project: name, S3Region: infoArgs.AWSRegion},
			nil,
			os.Stdout,
			os.Stderr,
		)

		info, err := client.FetchInfo()
		if err != nil {
			return err
		}

		if infoArgs.JSON {
			bytes, err := json.Marshal(info)
			if err != nil {
				return err
			}
			fmt.Println(string(bytes))
		} else {
			fmt.Println(info.String())
		}

		return nil
	},
}
