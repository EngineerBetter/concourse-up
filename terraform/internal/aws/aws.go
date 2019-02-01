package aws

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/EngineerBetter/concourse-up/util"
	"github.com/asaskevich/govalidator"
)

// InputVars holds all the parameters AWS IAAS needs
type InputVars struct {
	NetworkCIDR            string
	PublicCIDR             string
	PrivateCIDR            string
	AllowIPs               string
	AvailabilityZone       string
	ConfigBucket           string
	Deployment             string
	HostedZoneID           string
	HostedZoneRecordPrefix string
	Namespace              string
	Project                string
	PublicKey              string
	RDSDefaultDatabaseName string
	RDSInstanceClass       string
	RDSPassword            string
	RDSUsername            string
	Region                 string
	SourceAccessIP         string
	TFStatePath            string
	MultiAZRDS             bool
}

// ConfigureTerraform interpolates terraform contents and returns terraform config
func (v *InputVars) ConfigureTerraform(terraformContents string) (string, error) {
	terraformConfig, err := util.RenderTemplate(terraformContents, v)
	if terraformConfig == nil {
		return "", err
	}
	return string(terraformConfig), err
}

// Build fills the InputVars struct with values from the map input
func (v *InputVars) Build(data map[string]interface{}) error {
	reflectValue := reflect.ValueOf(v)
	env := reflectValue.Elem()
	reflectType := env.Type()
	for i := 0; i < env.NumField(); i++ {
		value, ok := data[reflectType.Field(i).Name]
		if !ok {
			return errors.New("terraform:aws field " + reflectType.Field(i).Name + " should be provided")
		}
		if !env.Field(i).CanSet() {
			return errors.New("terraform:aws field " + reflectType.Field(i).Name + " cannot be set")
		}
		switch trueValue := value.(type) {
		case string:
			env.Field(i).SetString(trueValue)
		case bool:
			env.Field(i).SetBool(trueValue)
		default:
			return fmt.Errorf("Value: %v of type %T not supported", value, value)
		}
	}
	i := env.Interface()
	tempStruct, ok := i.(InputVars)
	if !ok {
		return errors.New("could not Build terraform input data")
	}
	v = &tempStruct // nolint
	return nil
}

// MetadataStringValue is a terraform output string variable
type MetadataStringValue struct {
	Value string `json:"value"`
}

// Metadata represents output from terraform on AWS or GCP
type Metadata struct {
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
func (metadata *Metadata) AssertValid() error {
	_, err := govalidator.ValidateStruct(metadata)
	return err
}

// Init populates metadata struct with values from the buffer
func (metadata *Metadata) Init(buffer *bytes.Buffer) error {
	if err := json.NewDecoder(buffer).Decode(&metadata); err != nil {
		return err
	}

	return nil
}

// Get returns a the specified value from the metadata struct
func (metadata *Metadata) Get(key string) (string, error) {
	reflectValue := reflect.ValueOf(metadata)
	reflectStruct := reflectValue.Elem()
	value := reflectStruct.FieldByName(key)
	if !value.IsValid() {
		return "", errors.New(key + " key not found")
	}

	return value.FieldByName("Value").String(), nil
}
