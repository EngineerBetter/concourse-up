package director

import (
	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
	"bitbucket.org/engineerbetter/concourse-up/util"
)

type awsCloudConfigParams struct {
	AvailabilityZone       string
	DefaultSecurityGroupID string
	DirectorSubnetID       string
	VMsSubnetID            string
}

func generateCloudConfig(conf *config.Config, metadata *terraform.Metadata) ([]byte, error) {
	templateParams := awsCloudConfigParams{
		AvailabilityZone:       conf.AvailabilityZone,
		DefaultSecurityGroupID: metadata.DirectorSecurityGroupID.Value,
		DirectorSubnetID:       metadata.DirectorSubnetID.Value,
		VMsSubnetID:            metadata.VMsSubnetID.Value,
	}

	return util.RenderTemplate(awsCloudConfigtemplate, templateParams)
}

var awsCloudConfigtemplate = `---
azs:
- name: z1
  cloud_properties:
    availability_zone: <% .AvailabilityZone %>

vm_types:
- name: medium
  cloud_properties:
    instance_type: m3.medium
    spot_bid_price: 0.09 # on-demand price: 0.073
    spot_ondemand_fallback: true
    ephemeral_disk:
      size: 200_000
      type: gp2
    security_groups:
    - <% .DefaultSecurityGroupID %>

- name: large
  cloud_properties:
    instance_type: m3.xlarge
    spot_bid_price: 0.320 # on-demand price: 0.266
    spot_ondemand_fallback: true
    ephemeral_disk:
      size: 200_000
      type: gp2
    security_groups:
    - <% .DefaultSecurityGroupID %>

disk_types:
- name: default
  disk_size: 50_000
  cloud_properties:
    type: gp2
- name: large
  disk_size: 200_000
  cloud_properties:
    type: gp2

networks:
- name: default
  type: manual
  subnets:
  - range: 10.0.0.0/24
    gateway: 10.0.0.1
    az: z1
    static:
    - 10.0.0.6
    reserved:
    - 10.0.0.1-10.0.0.5
    dns:
    - 10.0.0.2
    cloud_properties:
      subnet: <% .DirectorSubnetID %>

- name: vms
  type: manual
  subnets:
  - range: 10.0.1.0/24
    gateway: 10.0.1.1
    az: z1
    reserved:
    - 10.0.1.1
    dns:
    - 10.0.0.2
    cloud_properties:
      subnet: <% .VMsSubnetID %>

- name: vip
  type: vip

compilation:
  workers: 5
  reuse_compilation_vms: true
  az: z1
  vm_type: large
  network: default
`
