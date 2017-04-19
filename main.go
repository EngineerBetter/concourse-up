package main

import (
	"os"

	"bitbucket.org/engineerbetter/concourse-up/commands"

	"gopkg.in/urfave/cli.v1" // imports as package "cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Concourse-Up"
	app.Usage = "A CLI tool to deploy Concourse CI"
	app.Commands = commands.Commands
	app.Flags = commands.GlobalFlags
	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
