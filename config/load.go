package config

import (
	"encoding/json"
	"fmt"
	"html/template"

	"strings"

	"bitbucket.org/engineerbetter/concourse-up/aws"
	"bitbucket.org/engineerbetter/concourse-up/util"
)

const terraformStateFileName = "terraform.tfstate"
const configBucketS3Region = "eu-west-1"
const configFilePath = "config.json"

// Config represents a concourse-up configuration file
type Config struct {
	PublicKey   template.HTML `json:"public_key"`
	PrivateKey  template.HTML `json:"private_key"`
	Region      string        `json:"region"`
	Deployment  string        `json:"deployment"`
	TFStatePath string        `json:"tf_state_path"`
	Project     string        `json:"project"`
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

	configBytes, err := aws.LoadFile(deployment, configFilePath, configBucketS3Region)
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

	if err = aws.EnsureBucketExists(deployment, configBucketS3Region); err != nil {
		return nil, err
	}

	configBytes, err := aws.EnsureFileExists(deployment, configFilePath, configBucketS3Region, defaultConfigBytes)
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

	conf := Config{
		PublicKey:   template.HTML(strings.TrimSpace(string(publicKey))),
		PrivateKey:  template.HTML(strings.TrimSpace(string(privateKey))),
		Deployment:  deployment,
		Project:     project,
		TFStatePath: terraformStateFileName,
		Region:      region,
	}

	return json.Marshal(&conf)
}
