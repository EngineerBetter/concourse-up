package concourse

import (
	"io"

	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
	"bitbucket.org/engineerbetter/concourse-up/util"
)

// Destroy destroys a concourse instance
func Destroy(name string, terraformApplier terraform.Applier, configClient config.IClient, stdout, stderr io.Writer) error {
	config, err := configClient.Load(name)
	if err != nil {
		return err
	}

	terraformConfig, err := util.RenderTemplate(template, config)
	if err != nil {
		return err
	}

	return terraformApplier(terraformConfig, stdout, stderr)
}
