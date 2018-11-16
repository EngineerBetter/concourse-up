package gcp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/EngineerBetter/concourse-up/util"
	"github.com/asaskevich/govalidator"
)

// InputVars holds all the parameters GCP IAAS needs
type InputVars struct {
	Region             string
	Zone               string
	Tags               string
	Project            string
	GCPCredentialsJSON string
	ExternalIP         string
	Deployment         string
	ConfigBucket       string
	DBUsername         string
	DBPassword         string
	DBTier             string
	DBName             string
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
			return errors.New("terraform:gcp field " + reflectType.Field(i).Name + " should be provided")
		}
		if !env.Field(i).CanSet() {
			return errors.New("terraform:gcp field " + reflectType.Field(i).Name + " cannot be set")
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

// Metadata represents output from terraform on GCP or GCP
type Metadata struct {
	Network                    MetadataStringValue `json:"network" valid:"required"`
	PrivateSubnetworkName      MetadataStringValue `json:"private_subnetwork_name" valid:"required"`
	PublicSubnetworkName       MetadataStringValue `json:"public_subnetwork_name" valid:"required"`
	PublicSubnetworkCidr       MetadataStringValue `json:"public_subnetwork_cidr" valid:"required"`
	PrivateSubnetworkCidr      MetadataStringValue `json:"private_subnetwork_cidr" valid:"required"`
	PrivateSubnetworInternalGw MetadataStringValue `json:"private_subnetwor_internal_gw" valid:"required"`
	PublicSubnetworInternalGw  MetadataStringValue `json:"public_subnetwor_internal_gw" valid:"required"`
	ATCPublicIP                MetadataStringValue `json:"atc_public_ip" valid:"required"`
	DirectorAccountCreds       MetadataStringValue `json:"director_account_creds" valid:"required"`
	DirectorPublicIP           MetadataStringValue `json:"director_public_ip" valid:"required"`
	DBAddress                  MetadataStringValue `json:"db_address" valid:"required"`
	DBName                     MetadataStringValue `json:"db_name" valid:"required"`
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
