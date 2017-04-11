package main

import (
	"os"

	"gopkg.in/urfave/cli.v1" // imports as package "cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Concourse-Up"
	app.Usage = "A CLI tool to deploy Concourse CI"
	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
