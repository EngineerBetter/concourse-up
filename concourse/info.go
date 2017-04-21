package concourse

import (
	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
)

// Info represents the terraform output and concourse-up config files
type Info struct {
	Terraform *terraform.Metadata `json:"terraform"`
	Config    *config.Config      `json:"config"`
}

// FetchInfo fetches and builds the info
func (client *Client) FetchInfo() (info *Info, err error) {
	config, err := client.configClient.Load()
	if err != nil {
		return
	}

	terraformClient, err := client.buildTerraformClient(config)
	if err != nil {
		return
	}

	metadata, err := terraformClient.Output()
	if err != nil {
		return
	}
	defer func() { err = terraformClient.Cleanup() }()

	info = &Info{
		Terraform: metadata,
		Config:    config,
	}

	return
}
