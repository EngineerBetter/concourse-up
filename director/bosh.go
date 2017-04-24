package director

import (
	"bytes"
	"io"
	"io/ioutil"

	boshcmd "github.com/cloudfoundry/bosh-cli/cmd"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

// https://github.com/cloudfoundry/bosh-cli/blob/master/main.go
func (client *BoshInitClient) runBoshCommand(args ...string) ([]byte, error) {
	combinedOutputBuffer := bytes.NewBuffer(nil)
	stdout := io.MultiWriter(client.stdout, combinedOutputBuffer)
	stderr := io.MultiWriter(client.stderr, combinedOutputBuffer)

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
