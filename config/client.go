package config

import (
	"encoding/json"
	"fmt"

	"github.com/EngineerBetter/concourse-up/aws"
)

const terraformStateFileName = "terraform.tfstate"
const configFilePath = "config.json"

// IClient is an interface for the config file client
type IClient interface {
	Load() (*Config, error)
	DeleteAll(config *Config) error
	LoadOrCreate(deployArgs *DeployArgs) (*Config, bool, error)
	Update(*Config) error
	StoreAsset(filename string, contents []byte) error
	HasAsset(filename string) (bool, error)
	LoadAsset(filename string) ([]byte, error)
	DeleteAsset(filename string) error
}

// Client is a client for loading the config file  from S3
type Client struct {
	aws      aws.IClient
	project  string
	s3Region string
}

// New instantiates a new client
func New(aws aws.IClient, project string, s3Region string) IClient {
	return &Client{
		aws,
		project,
		s3Region,
	}
}

// StoreAsset stores an associated configuration file
func (client *Client) StoreAsset(filename string, contents []byte) error {
	return client.aws.WriteFile(client.configBucket(),
		filename,
		client.s3Region,
		contents,
	)
}

// LoadAsset loads an associated configuration file
func (client *Client) LoadAsset(filename string) ([]byte, error) {
	return client.aws.LoadFile(
		client.configBucket(),
		filename,
		client.s3Region,
	)
}

// DeleteAsset deletes an associated configuration file
func (client *Client) DeleteAsset(filename string) error {
	return client.aws.DeleteFile(
		client.configBucket(),
		filename,
		client.s3Region,
	)
}

// HasAsset returns true if an associated configuration file exists
func (client *Client) HasAsset(filename string) (bool, error) {
	return client.aws.HasFile(
		client.configBucket(),
		filename,
		client.s3Region,
	)
}

// Update stores the conconcourse up config file to S3
func (client *Client) Update(config *Config) error {
	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return client.aws.WriteFile(client.configBucket(), configFilePath, client.s3Region, bytes)
}

// DeleteAll deletes the entire configuration bucket
func (client *Client) DeleteAll(config *Config) error {
	return client.aws.DeleteVersionedBucket(config.ConfigBucket, client.s3Region)
}

// Load loads an existing config file from S3
func (client *Client) Load() (*Config, error) {
	configBytes, err := client.aws.LoadFile(
		client.configBucket(),
		configFilePath,
		client.s3Region,
	)
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
func (client *Client) LoadOrCreate(deployArgs *DeployArgs) (*Config, bool, error) {
	defaultConfigBytes, err := generateDefaultConfig(
		deployArgs.IAAS,
		client.project,
		client.deployment(),
		client.configBucket(),
		deployArgs.AWSRegion,
	)
	if err != nil {
		return nil, false, err
	}

	if err = client.aws.EnsureBucketExists(client.configBucket(), client.s3Region); err != nil {
		return nil, false, err
	}

	configBytes, createdNewFile, err := client.aws.EnsureFileExists(
		client.configBucket(),
		configFilePath,
		client.s3Region,
		defaultConfigBytes,
	)
	if err != nil {
		return nil, false, err
	}

	conf := Config{}
	if err := json.Unmarshal(configBytes, &conf); err != nil {
		return nil, false, err
	}

	return &conf, createdNewFile, nil
}

func (client *Client) deployment() string {
	return fmt.Sprintf("concourse-up-%s", client.project)
}

func (client *Client) configBucket() string {
	return fmt.Sprintf("%s-%s-config", client.deployment(), client.s3Region)
}
