package bosh

import (
	"fmt"
	"io"
	"os"
	"path"

	"io/ioutil"

	bicmd "github.com/cloudfoundry/bosh-init/cmd"
	biui "github.com/cloudfoundry/bosh-init/ui"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/pivotal-golang/clock"
)

type BoshInitClient struct {
	manifestPath string
	stdout       io.Writer
	stderr       io.Writer
}

type IBoshInitClient interface {
	Deploy() ([]byte, error)
	Delete() error
}

type BoshInitClientFactory func(manifestPath string, stdout, stderr io.Writer) IBoshInitClient

func NewBoshInitClient(manifestPath string, stdout, stderr io.Writer) IBoshInitClient {
	return &BoshInitClient{
		manifestPath: manifestPath,
		stdout:       stdout,
		stderr:       stderr,
	}
}

func (client *BoshInitClient) Deploy() ([]byte, error) {
	logger := boshlog.NewWriterLogger(boshlog.LevelError, client.stdout, client.stderr)

	fileSystem := boshsys.NewOsFileSystemWithStrictTempRoot(logger)
	workspaceRootPath := path.Join(os.Getenv("HOME"), ".bosh_init")
	ui := biui.NewConsoleUI(logger)
	timeService := clock.NewClock()

	cmdFactory := bicmd.NewFactory(
		fileSystem,
		ui,
		timeService,
		logger,
		boshuuid.NewGenerator(),
		workspaceRootPath,
	)

	cmdRunner := bicmd.NewRunner(cmdFactory)
	stage := biui.NewStage(ui, timeService, logger)
	err := cmdRunner.Run(stage, "deploy", client.manifestPath)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadFile(fmt.Sprintf("%s-state.json", client.manifestPath))
}

func (client *BoshInitClient) Delete() error {
	logger := boshlog.NewWriterLogger(boshlog.LevelError, client.stdout, client.stderr)

	fileSystem := boshsys.NewOsFileSystemWithStrictTempRoot(logger)
	workspaceRootPath := path.Join(os.Getenv("HOME"), ".bosh_init")
	ui := biui.NewConsoleUI(logger)
	timeService := clock.NewClock()

	cmdFactory := bicmd.NewFactory(
		fileSystem,
		ui,
		timeService,
		logger,
		boshuuid.NewGenerator(),
		workspaceRootPath,
	)

	cmdRunner := bicmd.NewRunner(cmdFactory)
	stage := biui.NewStage(ui, timeService, logger)
	return cmdRunner.Run(stage, "delete", client.manifestPath)
}
