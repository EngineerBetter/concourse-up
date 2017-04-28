package concourse

import (
	"fmt"

	"github.com/engineerbetter/concourse-up/bosh"
	"github.com/engineerbetter/concourse-up/config"
	"github.com/engineerbetter/concourse-up/terraform"
)

// Info represents the terraform output and concourse-up config files
type Info struct {
	Terraform *terraform.Metadata `json:"terraform"`
	Config    *config.Config      `json:"config"`
	Instances []bosh.Instance     `json:"instances"`
}

// FetchInfo fetches and builds the info
func (client *Client) FetchInfo() (*Info, error) {
	config, err := client.configClient.Load()
	if err != nil {
		return nil, err
	}

	terraformClient, err := client.buildTerraformClient(config)
	if err != nil {
		return nil, err
	}

	metadata, err := terraformClient.Output()
	if err != nil {
		return nil, err
	}
	defer terraformClient.Cleanup()

	boshClient, err := client.buildBoshClient(config, metadata)
	if err != nil {
		return nil, err
	}
	defer boshClient.Cleanup()

	instances, err := boshClient.Instances()
	if err != nil {
		return nil, err
	}

	return &Info{
		Terraform: metadata,
		Config:    config,
		Instances: instances,
	}, nil
}

func (info *Info) String() string {
	str := "\n"
	str += fmt.Sprintf("Concourse credentials:\n\tusername: %s\n\tpassword: %s\n\tURL: https://%s\n\n", info.Config.ConcourseUsername, info.Config.ConcoursePassword, info.Config.Domain)
	str += fmt.Sprintf("Bosh credentials:\n\tusername: %s\n\tpassword: %s\n\tIP: %s\n\n", info.Config.DirectorUsername, info.Config.DirectorPassword, info.Terraform.DirectorPublicIP.Value)
	str += fmt.Sprintf("Workers:\n\tCount: %d\n\tSize: %s\n\n", info.Config.ConcourseWorkerCount, info.Config.ConcourseWorkerSize)
	str += fmt.Sprintf("Deployment:\n\tIAAS: aws\n\tRegion: %s\n\n", info.Config.Region)
	str += "Instances:"
	for _, instance := range info.Instances {
		str += fmt.Sprintf("\n\t%s\t%s\t%s", instance.Name, instance.IP, instance.State)
	}
	str += "\n\n"

	return str
}
