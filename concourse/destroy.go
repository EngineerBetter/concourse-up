package concourse

import (
	"io"

	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
	"bitbucket.org/engineerbetter/concourse-up/util"
)

// Destroy destroys a concourse instance
func Destroy(name string, clientFactory terraform.ClientFactory, configClient config.IClient, stdout, stderr io.Writer) error {
	config, err := configClient.Load(name)
	if err != nil {
		return err
	}

	terraformFile, err := util.RenderTemplate(template, config)
	if err != nil {
		return err
	}

	terraformClient, err := clientFactory(terraformFile, stdout, stderr)
	if err != nil {
		return err
	}

	defer func() {
		err = terraformClient.Cleanup()
	}()

	err = terraformClient.Destroy()

	return err
}
