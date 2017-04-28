package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/engineerbetter/concourse-up/aws"
	"github.com/engineerbetter/concourse-up/bosh"
	"github.com/engineerbetter/concourse-up/certs"
	"github.com/engineerbetter/concourse-up/concourse"
	"github.com/engineerbetter/concourse-up/config"
	"github.com/engineerbetter/concourse-up/terraform"

	"encoding/json"

	"gopkg.in/urfave/cli.v1"
)

var infoFlagJSON bool

var info = cli.Command{
	Name:      "info",
	Aliases:   []string{"i"},
	Usage:     "Fetches information on a deployed environment",
	ArgsUsage: "<name>",
	Flags: []cli.Flag{cli.BoolFlag{
		Name:        "json",
		Usage:       "Output as json",
		EnvVar:      "JSON",
		Destination: &infoFlagJSON,
	}},
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
			&config.Client{Project: name},
			nil,
			os.Stdout,
			os.Stderr,
		)

		info, err := client.FetchInfo()
		if err != nil {
			return err
		}

		if infoFlagJSON {
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
