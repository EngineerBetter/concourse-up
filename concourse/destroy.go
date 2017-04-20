package concourse

import (
	"fmt"
	"io"

	"bitbucket.org/engineerbetter/concourse-up/bosh"
	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
	"bitbucket.org/engineerbetter/concourse-up/util"
)

// Destroy destroys a concourse instance
func Destroy(name string,
	terraformClientFactory terraform.ClientFactory,
	boshInitClientFactory bosh.BoshInitClientFactory,
	configClient config.IClient, stdout, stderr io.Writer) error {
	config, err := configClient.Load(name)
	if err != nil {
		return err
	}

	terraformFile, err := util.RenderTemplate(template, config)
	if err != nil {
		return err
	}

	terraformClient, err := terraformClientFactory(terraformFile, stdout, stderr)
	if err != nil {
		return err
	}
	defer func() {
		err = terraformClient.Cleanup()
	}()

	metadata, err := terraformClient.Output()
	if err != nil {
		return err
	}

	boshInitClient, _, err := createBoshInitClient(config, metadata, boshInitClientFactory, stdout, stderr)
	if err != nil {
		return err
	}

	if err = boshInitClient.Delete(); err != nil {
		stderr.Write([]byte(fmt.Sprintf("Warning error deleting bosh director. Continuing with terraform deletion.\n\t%s", err.Error())))
	}

	err = terraformClient.Destroy()

	return err
}
