package director

import (
	boshcmd "github.com/cloudfoundry/bosh-cli/cmd"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

// https://github.com/cloudfoundry/bosh-cli/blob/master/main.go
func (client *BoshInitClient) runBoshCommand(args ...string) error {
	logger := boshlog.NewWriterLogger(boshInitLogLevel, client.stdout, client.stderr)

	ui := boshui.NewConfUI(logger)
	defer ui.Flush()

	cmdFactory := boshcmd.NewFactory(boshcmd.NewBasicDeps(ui, logger))

	cmd, err := cmdFactory.New(args)
	if err != nil {
		return err
	}

	return cmd.Execute()
}
