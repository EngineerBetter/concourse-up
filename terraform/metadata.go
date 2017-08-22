package terraform

import "github.com/asaskevich/govalidator"

// MetadataStringValue is a terraform output string variable
type MetadataStringValue struct {
	Value string `json:"value"`
}

// Metadata represents the terraform output variables
type Metadata struct {
	DirectorKeyPair         MetadataStringValue `json:"director_key_pair" valid:"required"`
	DirectorPublicIP        MetadataStringValue `json:"director_public_ip" valid:"required"`
	DirectorSecurityGroupID MetadataStringValue `json:"director_security_group_id" valid:"required"`
	VMsSecurityGroupID      MetadataStringValue `json:"vms_security_group_id" valid:"required"`
	ELBSecurityGroupID      MetadataStringValue `json:"elb_security_group_id" valid:"required"`
	PublicSubnetID          MetadataStringValue `json:"public_subnet_id" valid:"required"`
	PrivateSubnetID         MetadataStringValue `json:"private_subnet_id" valid:"required"`
	VPCID                   MetadataStringValue `json:"vpc_id" valid:"required"`
	NatGatewayIP            MetadataStringValue `json:"nat_gateway_ip" valid:"required"`

	BlobstoreBucket          MetadataStringValue `json:"blobstore_bucket" valid:"required"`
	BlobstoreUserAccessKeyID MetadataStringValue `json:"blobstore_user_access_key_id" valid:"required"`
	BlobstoreSecretAccessKey MetadataStringValue `json:"blobstore_user_secret_access_key" valid:"required"`
	BoshUserAccessKeyID      MetadataStringValue `json:"bosh_user_access_key_id" valid:"required"`
	BoshSecretAccessKey      MetadataStringValue `json:"bosh_user_secret_access_key" valid:"required"`
	BoshDBPort               MetadataStringValue `json:"bosh_db_port" valid:"required"`
	BoshDBAddress            MetadataStringValue `json:"bosh_db_address" valid:"required"`
	ELBName                  MetadataStringValue `json:"elb_name"`
	ELBDNSName               MetadataStringValue `json:"elb_dns_name"`
	SourceAccessIP           MetadataStringValue `json:"source_access_ip"`
}

// AssertValid returns an error if the struct contains any missing fields
func (metadata *Metadata) AssertValid() error {
	_, err := govalidator.ValidateStruct(metadata)
	return err
}
