package commands

import (
	"errors"
	"fmt"
	"os"

	"bitbucket.org/engineerbetter/concourse-up/concourse"
	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/terraform"

	"encoding/json"

	"gopkg.in/urfave/cli.v1"
)

var info = cli.Command{
	Name:      "info",
	Aliases:   []string{"i"},
	Usage:     "Fetches information on a deployed environment",
	ArgsUsage: "<name>",
	Flags:     deployFlags,
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

		info, err := concourse.FetchInfo(name, terraform.NewClient, &config.Client{}, logFile, logFile)
		if err != nil {
			return err
		}
		bytes, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(bytes))

		return nil
	},
}
