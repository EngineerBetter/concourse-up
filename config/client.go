package config

import (
	"encoding/json"
	"fmt"
	"github.com/EngineerBetter/concourse-up/iaas"
)

const terraformStateFileName = "terraform.tfstate"
const configFilePath = "config.json"

//go:generate counterfeiter . IClient
type IClient interface {
	Load() (Config, error)
	DeleteAll(config Config) error
	Update(Config) error
	StoreAsset(filename string, contents []byte) error
	HasAsset(filename string) (bool, error)
	ConfigExists() (bool, error)
	LoadAsset(filename string) ([]byte, error)
	DeleteAsset(filename string) error
	NewConfig() Config
}

// Client is a client for loading the config file  from S3
type Client struct {
	Iaas         iaas.Provider
	Project      string
	Namespace    string
	BucketName   string
	BucketExists bool
	BucketError  error
}

// New instantiates a new client
func New(iaas iaas.Provider, project, namespace string) *Client {
	namespace = determineNamespace(namespace, iaas.Region())
	bucketName, exists, err := determineBucketName(iaas, namespace, project)

	if !exists && err == nil {
		err = iaas.CreateBucket(bucketName)
	}

	return &Client{
		iaas,
		project,
		namespace,
		bucketName,
		exists,
		err,
	}
}

// StoreAsset stores an associated configuration file
func (client *Client) StoreAsset(filename string, contents []byte) error {
	return client.Iaas.WriteFile(client.configBucket(),
		filename,
		contents,
	)
}

// LoadAsset loads an associated configuration file
func (client *Client) LoadAsset(filename string) ([]byte, error) {
	return client.Iaas.LoadFile(
		client.configBucket(),
		filename,
	)
}

// DeleteAsset deletes an associated configuration file
func (client *Client) DeleteAsset(filename string) error {
	return client.Iaas.DeleteFile(
		client.configBucket(),
		filename,
	)
}

// HasAsset returns true if an associated configuration file exists
func (client *Client) HasAsset(filename string) (bool, error) {
	return client.Iaas.HasFile(
		client.configBucket(),
		filename,
	)
}

// ConfigExists returns true if the configuration file exists
func (client *Client) ConfigExists() (bool, error) {
	return client.HasAsset(configFilePath)
}

// Update stores the conconcourse up config file to S3
func (client *Client) Update(config Config) error {
	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return client.Iaas.WriteFile(client.configBucket(), configFilePath, bytes)
}

// DeleteAll deletes the entire configuration bucket
func (client *Client) DeleteAll(config Config) error {
	return client.Iaas.DeleteVersionedBucket(config.ConfigBucket)
}

// Load loads an existing config file from S3
func (client *Client) Load() (Config, error) {
	if client.BucketError != nil {
		return Config{}, client.BucketError
	}

	configBytes, err := client.Iaas.LoadFile(
		client.configBucket(),
		configFilePath,
	)
	if err != nil {
		return Config{}, err
	}

	conf := Config{}
	if err := json.Unmarshal(configBytes, &conf); err != nil {
		return Config{}, err
	}

	return conf, nil
}

func (client *Client) NewConfig() Config {
	return Config{
		ConfigBucket: client.configBucket(),
		Deployment:   deployment(client.Project),
		Namespace:    client.Namespace,
		Project:      client.Project,
		Region:       client.Iaas.Region(),
		TFStatePath:  terraformStateFileName,
	}
}

func (client *Client) configBucket() string {
	return client.BucketName
}

func deployment(project string) string {
	return fmt.Sprintf("concourse-up-%s", project)
}

func createBucketName(deployment, extension string) string {
	return fmt.Sprintf("%s-%s-config", deployment, extension)
}

func determineBucketName(iaas iaas.Provider, namespace, project string) (string, bool, error) {
	regionBucketName := createBucketName(deployment(project), iaas.Region())
	namespaceBucketName := createBucketName(deployment(project), namespace)

	foundRegionNamedBucket, err := iaas.BucketExists(regionBucketName)
	if err != nil {
		return "", false, err
	}
	foundNamespacedBucket, err := iaas.BucketExists(namespaceBucketName)
	if err != nil {
		return "", false, err
	}

	foundOne := foundRegionNamedBucket || foundNamespacedBucket

	switch {
	case !foundRegionNamedBucket && foundNamespacedBucket:
		return namespaceBucketName, foundOne, nil
	case foundRegionNamedBucket && !foundNamespacedBucket:
		return regionBucketName, foundOne, nil
	default:
		return namespaceBucketName, foundOne, nil
	}
}

func determineNamespace(namespace, region string) string {
	if namespace == "" {
		return region
	}
	return namespace
}
