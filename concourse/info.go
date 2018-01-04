package concourse

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/fatih/color"
)

// Info represents the terraform output and concourse-up config files
type Info struct {
	Terraform *terraform.Metadata `json:"terraform"`
	Config    *config.Config      `json:"config"`
	Secrets   map[string]string   `json:"secrets"`
	Instances []bosh.Instance     `json:"instances"`
}

var blue = color.New(color.FgCyan, color.Bold).SprintfFunc()

func (client *Client) fetchSecrets() (map[string]string, error) {
	credsBytes, err := client.configClient.LoadAsset(bosh.CredsFilename)
	if err != nil {
		return nil, err
	}
	type certificate struct {
		CA          string `yaml:"ca"`
		Certificate string `yaml:"certificate"`
		PrivateKey  string `yaml:"private_key"`
	}
	type creds struct {
		CredhubCLIPassword string      `yaml:"credhub_cli_password"`
		CredhubCert        certificate `yaml:"credhub-tls"`
	}
	var c creds
	err = yaml.Unmarshal(credsBytes, &c)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"credhub_password": c.CredhubCLIPassword,
		"credhub_username": "credhub-cli",
		"credhub_ca_cert":  c.CredhubCert.CA,
	}, nil
}

// FetchInfo fetches and builds the info
func (client *Client) FetchInfo() (*Info, error) {
	config, err := client.configClient.Load()
	if err != nil {
		return nil, err
	}

	terraformClient, err := client.terraformClientFactory(client.iaasClient.IAAS(), config, client.stdout, client.stderr)
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

	secrets, err := client.fetchSecrets()
	if err != nil {
		return nil, err
	}

	return &Info{
		Terraform: metadata,
		Config:    config,
		Secrets:   secrets,
		Instances: instances,
	}, nil
}

func (info *Info) String() string {
	boshCACert := strings.Join(strings.Split(info.Config.DirectorCACert, "\n"), "\n\t\t")

	str := "\n"
	str += fmt.Sprintf("Deployment:\n\tIAAS:   aws\n\tRegion: %s\n\n", info.Config.Region)
	str += fmt.Sprintf("Workers:\n\tCount:              %d\n\tSize:               %s\n\tOutbound Public IP: %s\n\n", info.Config.ConcourseWorkerCount, info.Config.ConcourseWorkerSize, info.Terraform.NatGatewayIP.Value)
	str += "Instances:"
	for _, instance := range info.Instances {
		str += fmt.Sprintf("\n\t%s\t%s\t%s", instance.Name, instance.IP, instance.State)
	}
	str += "\n\n"
	str += fmt.Sprintf("Concourse credentials:\n\tusername: %s\n\tpassword: %s\n\tURL:      https://%s\n\n", info.Config.ConcourseUsername, info.Config.ConcoursePassword, info.Config.Domain)
	str += fmt.Sprintf("Grafana credentials:\n\tusername: %s\n\tpassword: %s\n\tURL:      https://%s:3000\n\n", info.Config.ConcourseUsername, info.Config.ConcoursePassword, info.Config.Domain)
	str += fmt.Sprintf("Bosh credentials:\n\tusername: %s\n\tpassword: %s\n\tIP:       %s\n\tCA Cert:\n\t\t%s\n", info.Config.DirectorUsername, info.Config.DirectorPassword, info.Terraform.DirectorPublicIP.Value, boshCACert)

	str += fmt.Sprintf("Built by %s %s\n", blue("EngineerBetter"), blue("http://engineerbetter.com"))

	return str
}
