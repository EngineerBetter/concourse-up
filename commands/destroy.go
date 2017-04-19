package commands

import (
	"errors"
	"fmt"
	"os"

	"bitbucket.org/engineerbetter/concourse-up/util"

	"gopkg.in/urfave/cli.v1"
)

var destroy = cli.Command{
	Name:      "destroy",
	Usage:     "Destroys a Concourse",
	ArgsUsage: "<name>",
	Action: func(c *cli.Context) error {

		if len(os.Args) < 3 {
			fmt.Println("Usage is `concourse-up destroy <name>`")
			return nil
		}
		name := os.Args[2]

		confirm, err := util.CheckConfirmation(os.Stdin, os.Stdout, name)
		if err != nil {
			return err
		}

		if !confirm {
			fmt.Println("Bailing out...")
			return nil
		}

		return errors.New("not implemented")
	},
}
