package commands

import (
	"gopkg.in/urfave/cli.v1"
)

// Commands is a list of all supported CLI commands
var Commands = []cli.Command{
	deploy,
	destroy,
}

var nonInteractive bool

var GlobalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:        "non-interactive, n",
		EnvVar:      "NON_INTERACTIVE",
		Usage:       "Non interactive",
		Destination: &nonInteractive,
	},
}

func NonInteractiveModeEnabled() bool {
	return nonInteractive
}
