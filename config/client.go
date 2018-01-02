package config

import (
	"encoding/json"
	"fmt"

	"github.com/EngineerBetter/concourse-up/iaas"
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
	iaas    iaas.IClient
	project string
}

// New instantiates a new client
func New(iaas iaas.IClient, project string) *Client {
	return &Client{
		iaas,
		project,
	}
}

// StoreAsset stores an associated configuration file
func (client *Client) StoreAsset(filename string, contents []byte) error {
	return client.iaas.WriteFile(client.configBucket(),
		filename,
		contents,
	)
}

// LoadAsset loads an associated configuration file
func (client *Client) LoadAsset(filename string) ([]byte, error) {
	return client.iaas.LoadFile(
		client.configBucket(),
		filename,
	)
}

// DeleteAsset deletes an associated configuration file
func (client *Client) DeleteAsset(filename string) error {
	return client.iaas.DeleteFile(
		client.configBucket(),
		filename,
	)
}

// HasAsset returns true if an associated configuration file exists
func (client *Client) HasAsset(filename string) (bool, error) {
	return client.iaas.HasFile(
		client.configBucket(),
		filename,
	)
}

// Update stores the conconcourse up config file to S3
func (client *Client) Update(config *Config) error {
	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return client.iaas.WriteFile(client.configBucket(), configFilePath, bytes)
}

// DeleteAll deletes the entire configuration bucket
func (client *Client) DeleteAll(config *Config) error {
	return client.iaas.DeleteVersionedBucket(config.ConfigBucket)
}

// Load loads an existing config file from S3
func (client *Client) Load() (*Config, error) {
	configBytes, err := client.iaas.LoadFile(
		client.configBucket(),
		configFilePath,
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
	config, err := generateDefaultConfig(
		deployArgs.IAAS,
		client.project,
		client.deployment(),
		client.configBucket(),
		deployArgs.AWSRegion,
		DBSizes[deployArgs.DBSize],
	)
	if err != nil {
		return nil, false, err
	}
	defaultConfigBytes, err := json.Marshal(config)
	if err != nil {
		return nil, false, err
	}
	if err = client.iaas.EnsureBucketExists(client.configBucket()); err != nil {
		return nil, false, err
	}
	configBytes, createdNewFile, err := client.iaas.EnsureFileExists(
		client.configBucket(),
		configFilePath,
		defaultConfigBytes,
	)
	if err != nil {
		return nil, false, err
	}
	if err := json.Unmarshal(configBytes, config); err != nil {
		return nil, false, err
	}
	return config, createdNewFile, nil
}

func (client *Client) deployment() string {
	return fmt.Sprintf("concourse-up-%s", client.project)
}

func (client *Client) configBucket() string {
	return fmt.Sprintf("%s-%s-config", client.deployment(), client.iaas.Region())
}
