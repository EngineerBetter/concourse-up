package commands

import (
	cli "gopkg.in/urfave/cli.v1"
)

// Commands is a list of all supported CLI commands
var Commands = []cli.Command{
	deployCmd,
	destroyCmd,
	infoCmd,
	maintainCmd,
}

var nonInteractive bool

// GlobalFlags are the global CLIflags
var GlobalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:        "non-interactive, n",
		EnvVar:      "NON_INTERACTIVE",
		Usage:       "Non interactive",
		Destination: &nonInteractive,
	},
}

// NonInteractiveModeEnabled returns true if --non-interactive true has been passed in
func NonInteractiveModeEnabled() bool {
	return nonInteractive
}
