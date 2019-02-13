package terraform

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/EngineerBetter/concourse-up/util"
	"github.com/asaskevich/govalidator"
)

// InputVars holds all the parameters GCP IAAS needs
type GCPInputVars struct {
	AllowIPs           string
	ConfigBucket       string
	DBName             string
	DBPassword         string
	DBTier             string
	DBUsername         string
	Deployment         string
	DNSManagedZoneName string
	DNSRecordSetPrefix string
	ExternalIP         string
	GCPCredentialsJSON string
	Namespace          string
	PrivateCIDR        string
	Project            string
	PublicCIDR         string
	Region             string
	Tags               string
	Zone               string
}

// ConfigureTerraform interpolates terraform contents and returns terraform config
func (v *GCPInputVars) ConfigureTerraform(terraformContents string) (string, error) {
	terraformConfig, err := util.RenderTemplate("terraform", terraformContents, v)
	if terraformConfig == nil {
		return "", err
	}
	return string(terraformConfig), err
}

// Metadata represents output from terraform on GCP or GCP
type GCPOutputs struct {
	Network                    MetadataStringValue `json:"network" valid:"required"`
	PrivateSubnetworkName      MetadataStringValue `json:"private_subnetwork_name" valid:"required"`
	PublicSubnetworkName       MetadataStringValue `json:"public_subnetwork_name" valid:"required"`
	PrivateSubnetworkInternalGw MetadataStringValue `json:"private_subnetwork_internal_gw" valid:"required"`
	PublicSubnetworkInternalGw  MetadataStringValue `json:"public_subnetwork_internal_gw" valid:"required"`
	ATCPublicIP                MetadataStringValue `json:"atc_public_ip" valid:"required"`
	DirectorAccountCreds       MetadataStringValue `json:"director_account_creds" valid:"required"`
	DirectorPublicIP           MetadataStringValue `json:"director_public_ip" valid:"required"`
	BoshDBAddress              MetadataStringValue `json:"bosh_db_address" valid:"required"`
	DBName                     MetadataStringValue `json:"db_name" valid:"required"`
	NatGatewayIP               MetadataStringValue `json:"nat_gateway_ip" valid:"required"`
	SQLServerCert              MetadataStringValue `json:"server_ca_cert" valid:"required"`
	DirectorSecurityGroupID    MetadataStringValue `json:"director_firewall_name" valid:"required"`
}

// AssertValid returns an error if the struct contains any missing fields
func (outputs *GCPOutputs) AssertValid() error {
	_, err := govalidator.ValidateStruct(outputs)
	return err
}

// Init populates outputs struct with values from the buffer
func (outputs *GCPOutputs) Init(buffer *bytes.Buffer) error {
	if err := json.NewDecoder(buffer).Decode(&outputs); err != nil {
		return err
	}

	return nil
}

// Get returns a the specified value from the outputs struct
func (outputs *GCPOutputs) Get(key string) (string, error) {
	reflectValue := reflect.ValueOf(outputs)
	reflectStruct := reflectValue.Elem()
	value := reflectStruct.FieldByName(key)
	if !value.IsValid() {
		return "", errors.New(key + " key not found")
	}

	return value.FieldByName("Value").String(), nil
}
