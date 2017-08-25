package concourse

import (
	"fmt"
	"strings"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/terraform"
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

	terraformClient, err := client.terraformClientFactory(config, client.stdout, client.stderr)
	if err != nil {
		return nil, err
	}
	defer terraformClient.Cleanup()

	metadata, err := terraformClient.Output()
	if err != nil {
		return nil, err
	}

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
	boshCACert := strings.Join(strings.Split(info.Config.DirectorCACert, "\n"), "\n\t\t")

	str := "\n"
	str += fmt.Sprintf("Deployment:\n\tIAAS:   aws\n\tRegion: %s\n\n", info.Config.Region)
	str += fmt.Sprintf("Workers:\n\tCount:              %d\n\tSize:               %s\n\tOutbound Public IP: %s\n\n", info.Config.ConcourseWorkerCount, info.Config.ConcourseWorkerSize, info.Terraform.AWS.NatGatewayIP.Value)
	str += "Instances:"
	for _, instance := range info.Instances {
		str += fmt.Sprintf("\n\t%s\t%s\t%s", instance.Name, instance.IP, instance.State)
	}
	str += "\n\n"
	str += fmt.Sprintf("Concourse credentials:\n\tusername: %s\n\tpassword: %s\n\tURL:      https://%s\n\n", info.Config.ConcourseUsername, info.Config.ConcoursePassword, info.Config.Domain)
	str += fmt.Sprintf("Bosh credentials:\n\tusername: %s\n\tpassword: %s\n\tIP:       %s\n\tCA Cert:\n\t\t%s\n", info.Config.DirectorUsername, info.Config.DirectorPassword, info.Terraform.AWS.DirectorPublicIP.Value, boshCACert)

	return str
}
