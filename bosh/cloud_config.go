package bosh

import (
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"
)

type awsCloudConfigParams struct {
	AvailabilityZone   string
	VMsSecurityGroupID string
	ATCSecurityGroupID string
	PublicSubnetID     string
	PrivateSubnetID    string
	Spot               bool
}

func generateCloudConfig(conf config.Config, metadata *terraform.Metadata) ([]byte, error) {
	templateParams := awsCloudConfigParams{
		AvailabilityZone:   conf.AvailabilityZone,
		VMsSecurityGroupID: metadata.VMsSecurityGroupID.Value,
		ATCSecurityGroupID: metadata.ATCSecurityGroupID.Value,
		PublicSubnetID:     metadata.PublicSubnetID.Value,
		PrivateSubnetID:    metadata.PrivateSubnetID.Value,
		Spot:               conf.Spot,
	}

	return util.RenderTemplate(awsCloudConfigtemplate, templateParams)
}

var awsCloudConfigtemplate = string(MustAsset("assets/cloud-config.yml"))
