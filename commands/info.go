package commands

import (
	"errors"
	"fmt"
	"os"

	"bitbucket.org/engineerbetter/concourse-up/certs"
	"bitbucket.org/engineerbetter/concourse-up/concourse"
	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/director"
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
	Action: func(c *cli.Context) (err error) {
		name := c.Args().Get(0)
		if name == "" {
			err = errors.New("Usage is `concourse-up info <name>`")
			return
		}

		var logFile *os.File
		logFile, err = os.Create("terraform.combined.log")
		if err != nil {
			return
		}
		defer func() { err = logFile.Close() }()

		client := concourse.NewClient(
			terraform.NewClient,
			director.NewBoshInitClient,
			certs.Generate,
			&config.Client{Project: name},
			os.Stdout,
			os.Stderr,
		)

		var info *concourse.Info
		info, err = client.FetchInfo()
		if err != nil {
			return
		}
		var bytes []byte
		bytes, err = json.MarshalIndent(info, "", "  ")
		if err != nil {
			return
		}

		fmt.Println(string(bytes))

		return nil
	},
}
