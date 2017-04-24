package terraform

import "github.com/asaskevich/govalidator"

// MetadataStringValue is a terraform output string variable
type MetadataStringValue struct {
	Value string `json:"value"`
}

// Metadata represents the terraform output variables
type Metadata struct {
	DirectorKeyPair          MetadataStringValue `json:"director_key_pair" valid:"required"`
	DirectorPublicIP         MetadataStringValue `json:"director_public_ip" valid:"required"`
	DirectorSecurityGroupID  MetadataStringValue `json:"director_security_group_id" valid:"required"`
	VMsSecurityGroupID       MetadataStringValue `json:"vms_security_group_id" valid:"required"`
	ELBSecurityGroupID       MetadataStringValue `json:"elb_security_group_id" valid:"required"`
	DirectorSubnetID         MetadataStringValue `json:"director_subnet_id" valid:"required"`
	ConcourseSubnetID        MetadataStringValue `json:"concourse_subnet_id" valid:"required"`
	BlobstoreBucket          MetadataStringValue `json:"blobstore_bucket" valid:"required"`
	BlobstoreUserAccessKeyID MetadataStringValue `json:"blobstore_user_access_key_id" valid:"required"`
	BlobstoreSecretAccessKey MetadataStringValue `json:"blobstore_user_secret_access_key" valid:"required"`
	BoshUserAccessKeyID      MetadataStringValue `json:"bosh_user_access_key_id" valid:"required"`
	BoshSecretAccessKey      MetadataStringValue `json:"bosh_user_secret_access_key" valid:"required"`
	BoshDBUsername           MetadataStringValue `json:"bosh_db_username" valid:"required"`
	BoshDBPassword           MetadataStringValue `json:"bosh_db_password" valid:"required"`
	BoshDBPort               MetadataStringValue `json:"bosh_db_port" valid:"required"`
	BoshDBAddress            MetadataStringValue `json:"bosh_db_address" valid:"required"`
	ELBName                  MetadataStringValue `json:"elb_name"`
}

// AssertValid returns an error if the struct contains any missing fields
func (metadata *Metadata) AssertValid() error {
	_, err := govalidator.ValidateStruct(metadata)
	return err
}
