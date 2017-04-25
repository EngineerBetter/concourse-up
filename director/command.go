package director

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	boshcmd "github.com/cloudfoundry/bosh-cli/cmd"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

const boshInitLogLevel = boshlog.LevelWarn

var defaultBoshArgs = []string{"--non-interactive", "--tty", "--no-color"}

// RunAuthenticatedCommand runs a command against the bosh director, after authenticating
func (client *Client) RunAuthenticatedCommand(args ...string) ([]byte, error) {
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

	return client.RunCommand(args...)
}

// RunCommand runs a command against the bosh director
// https://github.com/cloudfoundry/bosh-cli/blob/master/main.go
func (client *Client) RunCommand(args ...string) ([]byte, error) {
	return runBoshCLIV2Command(client.stdout, client.stderr, args...)
}

func runBoshCLIV2Command(stdout, stderr io.Writer, args ...string) ([]byte, error) {
	combinedOutputBuffer := bytes.NewBuffer(nil)
	stdout = io.MultiWriter(stdout, combinedOutputBuffer)
	stderr = io.MultiWriter(stderr, combinedOutputBuffer)

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
		return nil, err
	}

	if err = cmd.Execute(); err != nil {
		return nil, err
	}

	ui.Flush()

	stdoutBytes, err := ioutil.ReadAll(combinedOutputBuffer)
	if err != nil {
		return nil, err
	}

	return stdoutBytes, nil
}
