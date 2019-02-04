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
var chosenIaas string

// GlobalFlags are the global CLIflags
var GlobalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:        "non-interactive, n",
		EnvVar:      "NON_INTERACTIVE",
		Usage:       "(optional) Non interactive",
		Destination: &nonInteractive,
	},
	cli.StringFlag{
		Name:        "iaas",
		Usage:       "(optional) IAAS, can be AWS or GCP",
		EnvVar:      "IAAS",
		Value:       "AWS",
		Destination: &chosenIaas,
	},
}

// NonInteractiveModeEnabled returns true if --non-interactive true has been passed in
func NonInteractiveModeEnabled() bool {
	return nonInteractive
}
