package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/commands"
	"github.com/EngineerBetter/concourse-up/director"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/terraform"
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

	if err := checkCompileTimeArgs(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func checkCompileTimeArgs() error {
	boshCompileTimeArgs := map[string]string{
		"bosh.DirectorCPIReleaseSHA1":    bosh.DirectorCPIReleaseSHA1,
		"bosh.DirectorCPIReleaseURL":     bosh.DirectorCPIReleaseURL,
		"bosh.DirectorCPIReleaseVersion": bosh.DirectorCPIReleaseVersion,
		"bosh.DirectorReleaseSHA1":       bosh.DirectorReleaseSHA1,
		"bosh.DirectorReleaseURL":        bosh.DirectorReleaseURL,
		"bosh.DirectorReleaseVersion":    bosh.DirectorReleaseVersion,
		"bosh.DirectorStemcellSHA1":      bosh.DirectorStemcellSHA1,
		"bosh.DirectorStemcellURL":       bosh.DirectorStemcellURL,
		"bosh.DirectorStemcellVersion":   bosh.DirectorStemcellVersion,
		"director.DarwinBinaryURL":       director.DarwinBinaryURL,
		"director.LinuxBinaryURL":        director.LinuxBinaryURL,
		"director.WindowsBinaryURL":      director.WindowsBinaryURL,
		"fly.DarwinBinaryURL":            fly.DarwinBinaryURL,
		"fly.LinuxBinaryURL":             fly.LinuxBinaryURL,
		"fly.WindowsBinaryURL":           fly.WindowsBinaryURL,
		"terraform.DarwinBinaryURL":      terraform.DarwinBinaryURL,
		"terraform.LinuxBinaryURL":       terraform.LinuxBinaryURL,
		"terraform.WindowsBinaryURL":     terraform.WindowsBinaryURL,
	}

	if ConcourseUpVersion == "" || strings.HasPrefix(ConcourseUpVersion, "COMPILE_TIME_VARIABLE") {
		return errors.New("Compile-time variable main.ConcourseUpVersion not set, please build with: `go build -ldflags \"-X main.ConcourseUpVersion=0.0.0\"`")
	}

	for key, value := range boshCompileTimeArgs {
		if value == "" || strings.HasPrefix(value, "COMPILE_TIME_VARIABLE") {
			return fmt.Errorf("Compile-time variable %s not set, please build with: `go build -ldflags \"-X github.com/EngineerBetter/concourse-up/%s=SOME_VALUE\"`", key, key)
		}
	}

	return nil
}
