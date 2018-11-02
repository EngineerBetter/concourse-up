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
	Zone               string `json:"zone" valid:"required"`
	Tags               string `json:"tags" valid:"required"`
	Project            string `json:"project" valid:"required"`
	GCPCredentialsJSON string `json:"gcp_credentials_json" valid:"required"`
	ExternalIP         string `json:"external_ip" valid:"required"`
	Network            string `json:"network" valid:"required"`
	Subnetwork         string `json:"subnetwork" valid:"required"`
	InternalCIDR       string `json:"internal_cidr" valid:"required"`
	InternalGW         string `json:"internal_gw" valid:"required"`
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
