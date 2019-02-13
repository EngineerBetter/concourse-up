package concourse

import (
	"fmt"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"
)

type TFInputVarsFactory interface {
	NewInputVars(conf config.Config) terraform.InputVars
}

func NewTFInputVarsFactory(provider iaas.Provider) (TFInputVarsFactory, error) {
	if provider.IAAS() == iaas.AWS {
		return &AWSInputVarsFactory{}, nil
	} else if provider.IAAS() == iaas.GCP {
		credentialsPath, err := provider.Attr("credentials_path")
		if err != nil {
			return &GCPInputVarsFactory{}, fmt.Errorf("Error finding attribute [credentials_path]: [%v]", err)
		}

		project, err := provider.Attr("project")
		if err != nil {
			return &GCPInputVarsFactory{}, fmt.Errorf("Error finding attribute [project]: [%v]", err)
		}

		return &GCPInputVarsFactory{
			credentialsPath: credentialsPath,
			project:         project,
			region:          provider.Region(),
			zone:            provider.Zone(""),
		}, nil
	}

	return nil, fmt.Errorf("IAAS not supported [%s]", provider.IAAS())
}

type AWSInputVarsFactory struct{}

func (f *AWSInputVarsFactory) NewInputVars(c config.Config) terraform.InputVars {
	return &terraform.AWSInputVars{
		NetworkCIDR:            c.NetworkCIDR,
		PublicCIDR:             c.PublicCIDR,
		PrivateCIDR:            c.PrivateCIDR,
		AllowIPs:               c.AllowIPs,
		AvailabilityZone:       c.AvailabilityZone,
		ConfigBucket:           c.ConfigBucket,
		Deployment:             c.Deployment,
		HostedZoneID:           c.HostedZoneID,
		HostedZoneRecordPrefix: c.HostedZoneRecordPrefix,
		Namespace:              c.Namespace,
		Project:                c.Project,
		PublicKey:              c.PublicKey,
		RDSDefaultDatabaseName: c.RDSDefaultDatabaseName,
		RDSInstanceClass:       c.RDSInstanceClass,
		RDSPassword:            c.RDSPassword,
		RDSUsername:            c.RDSUsername,
		Rds1CIDR:               c.Rds1CIDR,
		Rds2CIDR:               c.Rds2CIDR,
		Region:                 c.Region,
		SourceAccessIP:         c.SourceAccessIP,
		TFStatePath:            c.TFStatePath,
	}
}

type GCPInputVarsFactory struct {
	credentialsPath string
	project         string
	region          string
	zone            string
}

func (f *GCPInputVarsFactory) NewInputVars(c config.Config) terraform.InputVars {
	return &terraform.GCPInputVars{
		AllowIPs:           c.AllowIPs,
		ConfigBucket:       c.ConfigBucket,
		DBName:             c.RDSDefaultDatabaseName,
		DBPassword:         c.RDSPassword,
		DBTier:             c.RDSInstanceClass,
		DBUsername:         c.RDSUsername,
		Deployment:         c.Deployment,
		DNSManagedZoneName: c.HostedZoneID,
		DNSRecordSetPrefix: c.HostedZoneRecordPrefix,
		ExternalIP:         c.SourceAccessIP,
		GCPCredentialsJSON: f.credentialsPath,
		Namespace:          c.Namespace,
		Project:            f.project,
		Region:             f.region,
		Tags:               "",
		Zone:               f.zone,
		PublicCIDR:         c.PublicCIDR,
		PrivateCIDR:        c.PrivateCIDR,
	}
}
