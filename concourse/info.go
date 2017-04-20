package concourse

import (
	"io"

	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
	"bitbucket.org/engineerbetter/concourse-up/util"
)

// Info represents the terraform output and concourse-up config files
type Info struct {
	Terraform *terraform.Metadata `json:"terraform"`
	Config    *config.Config      `json:"config"`
}

// FetchInfo fetches and builds the info
func FetchInfo(name string, clientFactory terraform.ClientFactory, configClient config.IClient, stdout, stderr io.Writer) (*Info, error) {
	config, err := configClient.LoadOrCreate(name)
	if err != nil {
		return nil, err
	}

	terraformFile, err := util.RenderTemplate(template, config)
	if err != nil {
		return nil, err
	}

	terraformClient, err := clientFactory(terraformFile, stdout, stderr)
	if err != nil {
		return nil, err
	}

	defer terraformClient.Cleanup()

	err = terraformClient.Apply()
	if err != nil {
		return nil, err
	}

	metadata, err := terraformClient.Output()
	if err != nil {
		return nil, err
	}

	return &Info{
		Terraform: metadata,
		Config:    config,
	}, nil
}
