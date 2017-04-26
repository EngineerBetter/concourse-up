package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/engineerbetter/concourse-up/bosh"
	"github.com/engineerbetter/concourse-up/certs"
	"github.com/engineerbetter/concourse-up/concourse"
	"github.com/engineerbetter/concourse-up/config"
	"github.com/engineerbetter/concourse-up/terraform"
	"github.com/engineerbetter/concourse-up/util"

	"gopkg.in/urfave/cli.v1"
)

var destroy = cli.Command{
	Name:      "destroy",
	Aliases:   []string{"x"},
	Usage:     "Destroys a Concourse",
	ArgsUsage: "<name>",
	Action: func(c *cli.Context) error {
		name := c.Args().Get(0)
		if name == "" {
			return errors.New("Usage is `concourse-up destroy <name>`")
		}

		if !NonInteractiveModeEnabled() {
			confirm, err := util.CheckConfirmation(os.Stdin, os.Stdout, name)
			if err != nil {
				return err
			}

			if !confirm {
				fmt.Println("Bailing out...")
				return nil
			}
		}

		client := concourse.NewClient(
			terraform.NewClient,
			bosh.NewClient,
			certs.Generate,
			&config.Client{Project: name},
			map[string]string{},
			os.Stdout,
			os.Stderr,
		)

		return client.Destroy()
	},
}
