package bosh

import (
	"io"

	"github.com/EngineerBetter/concourse-up/bosh/internal/boshenv"
	"github.com/EngineerBetter/concourse-up/bosh/internal/director"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"
)

//GCPClient is an GCP specific implementation of IClient
type GCPClient struct {
	config   config.Config
	outputs  terraform.Outputs
	director director.IClient
	stdout   io.Writer
	stderr   io.Writer
	provider iaas.Provider
	boshCLI  boshenv.IBOSHCLI
}

//NewGCPClient returns a GCP specific implementation of IClient
func NewGCPClient(config config.Config, outputs terraform.Outputs, director director.IClient, stdout, stderr io.Writer, provider iaas.Provider, boshCLI boshenv.IBOSHCLI) (IClient, error) {
	return &GCPClient{
		config:   config,
		outputs:  outputs,
		director: director,
		stdout:   stdout,
		stderr:   stderr,
		provider: provider,
		boshCLI:  boshCLI,
	}, nil
}

//Cleanup is GCP specific implementation of Cleanup
func (client *GCPClient) Cleanup() error {
	return client.director.Cleanup()
}
