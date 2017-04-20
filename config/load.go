package config

import (
	"encoding/json"
	"fmt"

	"strings"

	"bitbucket.org/engineerbetter/concourse-up/aws"
	"bitbucket.org/engineerbetter/concourse-up/util"
)

const terraformStateFileName = "terraform.tfstate"
const configBucketS3Region = "eu-west-1"
const configFilePath = "config.json"

// Config represents a concourse-up configuration file
type Config struct {
	PublicKey                string `json:"public_key"`
	PrivateKey               string `json:"private_key"`
	Region                   string `json:"region"`
	AvailabilityZone         string `json:"availability_zone"`
	Deployment               string `json:"deployment"`
	RDSDefaultDatabaseName   string `json:"rds_default_database_name"`
	SourceAccessIP           string `json:"source_access_ip"`
	TFStatePath              string `json:"tf_state_path"`
	Project                  string `json:"project"`
	ConfigBucket             string `json:"config_bucket"`
	DirectorUsername         string `json:"director_username"`
	DirectorPassword         string `json:"director_password"`
	DirectorHMUserPassword   string `json:"director_hm_user_password"`
	DirectorMbusPassword     string `json:"director_mbus_password"`
	DirectorNATSPassword     string `json:"director_nats_password"`
	DirectorRegistryPassword string `json:"director_registry_password"`
	RDSInstanceClass         string `json:"rds_instance_class"`
	RDSUsername              string `json:"rds_username"`
	RDSPassword              string `json:"rds_password"`
}

// IClient is an interface for the config file client
type IClient interface {
	Load(deployment string) (*Config, error)
	LoadOrCreate(deployment string) (*Config, error)
}

// Client is a client for loading the config file from S3
type Client struct{}

// Load loads an existing config file from S3
func (client *Client) Load(project string) (*Config, error) {
	deployment := fmt.Sprintf("concourse-up-%s", project)
	configBucket := fmt.Sprintf("%s-config", deployment)

	configBytes, err := aws.LoadFile(configBucket, configFilePath, configBucketS3Region)
	if err != nil {
		return nil, err
	}

	conf := Config{}
	if err := json.Unmarshal(configBytes, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

// LoadOrCreate loads an existing config file from S3, or creates a default if one doesn't already exist
func (client *Client) LoadOrCreate(project string) (*Config, error) {
	deployment := fmt.Sprintf("concourse-up-%s", project)

	defaultConfigBytes, err := generateDefaultConfig(project, deployment, configBucketS3Region)
	if err != nil {
		return nil, err
	}

	configBucket := fmt.Sprintf("%s-config", deployment)

	if err = aws.EnsureBucketExists(configBucket, configBucketS3Region); err != nil {
		return nil, err
	}

	configBytes, err := aws.EnsureFileExists(configBucket, configFilePath, configBucketS3Region, defaultConfigBytes)
	if err != nil {
		return nil, err
	}

	conf := Config{}
	if err := json.Unmarshal(configBytes, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

func generateDefaultConfig(project, deployment, region string) ([]byte, error) {
	privateKey, publicKey, err := util.MakeSSHKeyPair()
	if err != nil {
		return nil, err
	}

	accessIP, err := util.FindUserIP()
	if err != nil {
		return nil, err
	}

	configBucket := fmt.Sprintf("%s-config", deployment)

	conf := Config{
		PublicKey:                strings.TrimSpace(string(publicKey)),
		PrivateKey:               strings.TrimSpace(string(privateKey)),
		Deployment:               deployment,
		ConfigBucket:             configBucket,
		RDSDefaultDatabaseName:   "bosh",
		SourceAccessIP:           accessIP,
		Project:                  project,
		TFStatePath:              terraformStateFileName,
		Region:                   region,
		AvailabilityZone:         fmt.Sprintf("%sa", region),
		DirectorUsername:         "admin",
		DirectorPassword:         util.GeneratePassword(),
		DirectorHMUserPassword:   util.GeneratePassword(),
		DirectorMbusPassword:     util.GeneratePassword(),
		DirectorNATSPassword:     util.GeneratePassword(),
		DirectorRegistryPassword: util.GeneratePassword(),
		RDSInstanceClass:         "db.t2.small",
		RDSUsername:              util.GeneratePassword(),
		RDSPassword:              util.GeneratePassword(),
	}

	return json.MarshalIndent(&conf, "", "  ")
}
