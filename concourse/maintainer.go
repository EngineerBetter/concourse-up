package concourse

import (
	"fmt"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/terraform"
)

type Maintainer interface {
	GetTFInputVars(client *Client, conf config.Config, environment terraform.InputVars) (terraform.InputVars, error)
}

func NewMaintainer(iaas string) (Maintainer, error) {
	switch iaas {
	case awsConst:
		return NewAWSMaintainer(), nil
	case gcpConst:
		return NewGCPMaintainer(), nil
	}

	return nil, fmt.Errorf("Unrecognised IAAS [%s]", iaas)
}
