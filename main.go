package main

import (
	"fmt"
	"os"

	"github.com/EngineerBetter/concourse-up/commands"
	"github.com/fatih/color"

	"gopkg.in/urfave/cli.v1"
)

// ConcourseUpVersion is a compile-time variable set with -ldflags
var ConcourseUpVersion = "COMPILE_TIME_VARIABLE_main_concourseUpVersion"
var blue = color.New(color.FgCyan, color.Bold).SprintfFunc()

func main() {
	app := cli.NewApp()
	app.Name = "Concourse-Up"
	app.Usage = "A CLI tool to deploy Concourse CI"
	app.Version = ConcourseUpVersion
	app.Commands = commands.Commands
	app.Flags = commands.GlobalFlags
	cli.AppHelpTemplate = fmt.Sprintf(`%s

See 'concourse-up help <command>' to read about a specific command.

Built by %s %s

`, cli.AppHelpTemplate, blue("EngineerBetter"), blue("http://engineerbetter.com"))

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
