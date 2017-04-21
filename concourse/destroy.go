package concourse

import (
	"fmt"
	"io"

	"bitbucket.org/engineerbetter/concourse-up/config"

	"bitbucket.org/engineerbetter/concourse-up/director"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
)

// Destroy destroys a concourse instance
func (client *Client) Destroy() (err error) {
	var conf *config.Config
	conf, err = client.configClient.Load()
	if err != nil {
		return
	}

	var terraformClient terraform.IClient
	terraformClient, err = client.buildTerraformClient(conf)
	if err != nil {
		return
	}
	defer func() { err = terraformClient.Cleanup() }()

	var metadata *terraform.Metadata
	metadata, err = terraformClient.Output()
	if err != nil {
		return
	}

	var boshInitClient director.IBoshInitClient
	boshInitClient, err = client.buildBoshInitClient(conf, metadata)
	if err != nil {
		return
	}
	defer func() { err = boshInitClient.Cleanup() }()

	if err = boshInitClient.Delete(); err != nil {
		if err = writeDeleteBoshDirectorErrorWarning(client.stderr, err.Error()); err != nil {
			return
		}
	}

	err = terraformClient.Destroy()
	return
}

func writeDeleteBoshDirectorErrorWarning(stderr io.Writer, message string) error {
	_, err := stderr.Write([]byte(fmt.Sprintf(
		"Warning error deleting bosh director. Continuing with terraform deletion.\n\t%s", message)))

	return err
}
