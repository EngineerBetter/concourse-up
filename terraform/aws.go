package terraform

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/EngineerBetter/concourse-up/util"
	"github.com/asaskevich/govalidator"
)

// InputVars holds all the parameters AWS IAAS needs
type AWSInputVars struct {
	AllowIPs               string
	AvailabilityZone       string
	ConfigBucket           string
	Deployment             string
	HostedZoneID           string
	HostedZoneRecordPrefix string
	Namespace              string
	NetworkCIDR            string
	PrivateCIDR            string
	Project                string
	PublicCIDR             string
	PublicKey              string
	RDSDefaultDatabaseName string
	RDSInstanceClass       string
	RDSPassword            string
	RDSUsername            string
	RDS1CIDR               string
	RDS2CIDR               string
	Region                 string
	SourceAccessIP         string
	TFStatePath            string
}

// ConfigureTerraform interpolates terraform contents and returns terraform config
func (v *AWSInputVars) ConfigureTerraform(terraformContents string) (string, error) {
	terraformConfig, err := util.RenderTemplate("terraform", terraformContents, v)
	if terraformConfig == nil {
		return "", err
	}
	return string(terraformConfig), err
}

// MetadataStringValue is a terraform output string variable
type MetadataStringValue struct {
	Value string `json:"value"`
}

// Metadata represents output from terraform on AWS or GCP
type AWSOutputs struct {
	ATCPublicIP              MetadataStringValue `json:"atc_public_ip" valid:"required"`
	ATCSecurityGroupID       MetadataStringValue `json:"atc_security_group_id" valid:"required"`
	BlobstoreBucket          MetadataStringValue `json:"blobstore_bucket" valid:"required"`
	BlobstoreSecretAccessKey MetadataStringValue `json:"blobstore_user_secret_access_key" valid:"required"`
	BlobstoreUserAccessKeyID MetadataStringValue `json:"blobstore_user_access_key_id" valid:"required"`
	BoshDBAddress            MetadataStringValue `json:"bosh_db_address" valid:"required"`
	BoshDBPort               MetadataStringValue `json:"bosh_db_port" valid:"required"`
	BoshSecretAccessKey      MetadataStringValue `json:"bosh_user_secret_access_key" valid:"required"`
	BoshUserAccessKeyID      MetadataStringValue `json:"bosh_user_access_key_id" valid:"required"`
	DirectorKeyPair          MetadataStringValue `json:"director_key_pair" valid:"required"`
	DirectorPublicIP         MetadataStringValue `json:"director_public_ip" valid:"required"`
	DirectorSecurityGroupID  MetadataStringValue `json:"director_security_group_id" valid:"required"`
	NatGatewayIP             MetadataStringValue `json:"nat_gateway_ip" valid:"required"`
	PrivateSubnetID          MetadataStringValue `json:"private_subnet_id" valid:"required"`
	PublicSubnetID           MetadataStringValue `json:"public_subnet_id" valid:"required"`
	SourceAccessIP           MetadataStringValue `json:"source_access_ip"`
	VMsSecurityGroupID       MetadataStringValue `json:"vms_security_group_id" valid:"required"`
	VPCID                    MetadataStringValue `json:"vpc_id" valid:"required"`
}

// AssertValid returns an error if the struct contains any missing fields
func (outputs *AWSOutputs) AssertValid() error {
	_, err := govalidator.ValidateStruct(outputs)
	return err
}

// Init populates outputs struct with values from the buffer
func (outputs *AWSOutputs) Init(buffer *bytes.Buffer) error {
	if err := json.NewDecoder(buffer).Decode(&outputs); err != nil {
		return err
	}

	return nil
}

// Get returns a the specified value from the outputs struct
func (outputs *AWSOutputs) Get(key string) (string, error) {
	reflectValue := reflect.ValueOf(outputs)
	reflectStruct := reflectValue.Elem()
	value := reflectStruct.FieldByName(key)
	if !value.IsValid() {
		return "", errors.New(key + " key not found")
	}

	return value.FieldByName("Value").String(), nil
}
