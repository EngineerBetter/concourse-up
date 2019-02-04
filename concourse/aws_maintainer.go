package concourse

import (
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/terraform"
)

type AWSMaintainer struct {
}

func NewAWSMaintainer() *AWSMaintainer {
	return &AWSMaintainer{}
}

func (m *AWSMaintainer) GetTFInputVars(client *Client, conf config.Config, environment terraform.InputVars) (terraform.InputVars, error) {
	err := environment.Build(awsInputVarsMapFromConfig(conf))
	return environment, err
}
