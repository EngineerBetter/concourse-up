package commands

import (
	"gopkg.in/urfave/cli.v1"
)

// Commands is a list of all supported CLI commands
var Commands = []cli.Command{
	deploy,
}
