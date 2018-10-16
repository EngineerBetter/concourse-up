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

// Environment holds all the parameters AWS IAAS needs
type Environment struct {
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
}

// ConfigureTerraform interpolates terraform contents and returns terraform config
func (e *Environment) ConfigureTerraform(terraformContents string) (string, error) {
	cc, err := util.RenderTemplate(terraformContents, e)
	if cc == nil {
		return "", err
	}
	return string(cc), err
}

func (e *Environment) Build(data map[string]string) error {
	e1 := *e
	env := reflect.ValueOf(e1)
	fmt.Printf("VALUE: %v\nTYPE %T", e1, e1)
	t := env.Type()
	for i := 0; i < env.NumField(); i++ {
		v, ok := data[t.Field(i).Name]
		if !ok {
			return errors.New("terraform:aws field " + t.Field(i).Name + " should be provided")
		}
		if !env.Field(i).CanSet() {
			return errors.New("terraform:aws field " + t.Field(i).Name + " cannot be set")
		}
		env.Field(i).SetString(v)
	}
	i := env.Interface()
	ie := i.(Environment)
	e = &ie
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
	m := reflect.ValueOf(metadata)
	mv := m.FieldByName(key)
	if !mv.IsValid() {
		return "", errors.New(key + " key not found")
	}
	return mv.FieldByName("Value").String(), nil
}
