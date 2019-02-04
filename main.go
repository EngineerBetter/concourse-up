package main

import (
	"fmt"
	"io"
	"os"

	"github.com/EngineerBetter/concourse-up/commands"
	"github.com/fatih/color"

	cli "gopkg.in/urfave/cli.v1"
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

	oldPrinter := cli.HelpPrinter
	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		switch data.(type) {
		case cli.Command:
			dataCommand := data.(cli.Command)
			dataWithGlobalFlags := CommandWithGlobalFlags{
				Command: cli.Command{
					Flags:              append(commands.GlobalFlags, dataCommand.Flags...),
					Name:               dataCommand.Name,
					ShortName:          dataCommand.ShortName,
					Aliases:            dataCommand.Aliases,
					Usage:              dataCommand.Usage,
					UsageText:          dataCommand.UsageText,
					Description:        dataCommand.Description,
					ArgsUsage:          dataCommand.ArgsUsage,
					Category:           dataCommand.Category,
					BashComplete:       dataCommand.BashComplete,
					Before:             dataCommand.Before,
					After:              dataCommand.After,
					Action:             dataCommand.Action,
					OnUsageError:       dataCommand.OnUsageError,
					Subcommands:        dataCommand.Subcommands,
					SkipFlagParsing:    dataCommand.SkipFlagParsing,
					SkipArgReorder:     dataCommand.SkipArgReorder,
					HideHelp:           dataCommand.HideHelp,
					Hidden:             dataCommand.Hidden,
					HelpName:           dataCommand.HelpName,
					CustomHelpTemplate: dataCommand.CustomHelpTemplate,
				},
				GlobalFlags: commands.GlobalFlags,
			}
			oldPrinter(w, templ, dataWithGlobalFlags)
		default:
			oldPrinter(w, templ, data)
		}
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type CommandWithGlobalFlags struct {
	cli.Command
	GlobalFlags []cli.Flag
}
