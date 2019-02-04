package concourse

import (
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/terraform"
)

type GCPMaintainer struct {
}

func NewGCPMaintainer() *GCPMaintainer {
	return &GCPMaintainer{}
}

func (m *GCPMaintainer) GetTFInputVars(client *Client, conf config.Config, environment terraform.InputVars) (terraform.InputVars, error) {
	project, err1 := client.provider.Attr("project")
	if err1 != nil {
		return nil, err1
	}
	credentialsPath, err1 := client.provider.Attr("credentials_path")
	if err1 != nil {
		return nil, err1
	}
	err1 = environment.Build(gcpInputVarsMapFromConfig(conf, credentialsPath, project, client))
	if err1 != nil {
		return nil, err1
	}

	return environment, nil
}
