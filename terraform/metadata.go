package terraform

// MetadataStringValue is a terraform output string variable
type MetadataStringValue struct {
	Value string `json:"value"`
}

// Metadata represents the terraform output variables
type Metadata struct {
	DirectorKeyPair          MetadataStringValue `json:"director_key_pair"`
	DirectorPublicIP         MetadataStringValue `json:"director_public_ip"`
	DirectorSecurityGroupID  MetadataStringValue `json:"director_security_group_id"`
	VMsSecurityGroupID       MetadataStringValue `json:"vms_security_group_id"`
	DirectorSubnetID         MetadataStringValue `json:"director_subnet_id"`
	BlobstoreBucket          MetadataStringValue `json:"blobstore_bucket"`
	BlobstoreUserAccessKeyID MetadataStringValue `json:"blobstore_user_access_key_id"`
	BlobstoreSecretAccessKey MetadataStringValue `json:"blobstore_user_secret_access_key"`
	BoshUserAccessKeyID      MetadataStringValue `json:"bosh_user_access_key_id"`
	BoshSecretAccessKey      MetadataStringValue `json:"bosh_user_secret_access_key"`
	BoshDBUsername           MetadataStringValue `json:"bosh_db_username"`
	BoshDBPassword           MetadataStringValue `json:"bosh_db_password"`
	BoshDBPort               MetadataStringValue `json:"bosh_db_port"`
	BoshDBAddress            MetadataStringValue `json:"bosh_db_address"`
}
