package director

import (
	"fmt"
	"io"

	boshcmd "github.com/cloudfoundry/bosh-cli/cmd"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

const boshInitLogLevel = boshlog.LevelWarn

var defaultBoshArgs = []string{"--non-interactive", "--tty", "--no-color"}

// RunAuthenticatedCommand runs a command against the bosh director, after authenticating
func (client *Client) RunAuthenticatedCommand(stdout, stderr io.Writer, args ...string) error {
	args = append([]string{
		"--environment",
		fmt.Sprintf("https://%s", client.creds.Host),
		"--ca-cert",
		client.caCertPath,
		"--client",
		client.creds.Username,
		"--client-secret",
		client.creds.Password,
	}, args...)

	return client.RunCommand(stdout, stderr, args...)
}

// RunCommand runs a command against the bosh director
// https://github.com/cloudfoundry/bosh-cli/blob/master/main.go
func (client *Client) RunCommand(stdout, stderr io.Writer, args ...string) error {
	logger := boshlog.NewWriterLogger(boshInitLogLevel,
		stdout,
		stderr,
	)

	ui := boshui.NewConfUI(logger)
	defer ui.Flush()
	writerUI := boshui.NewWriterUI(stdout, stderr, logger)

	// NOTE SetParent is implemented manually on vendored version of bosh-cli
	ui.SetParent(writerUI)

	basicDeps := boshcmd.NewBasicDeps(ui, logger)
	cmdFactory := boshcmd.NewFactory(basicDeps)

	args = append(defaultBoshArgs, args...)

	cmd, err := cmdFactory.New(args)
	if err != nil {
		return err
	}

	if err = cmd.Execute(); err != nil {
		return err
	}

	ui.Flush()

	return nil
}
