package config

import (
	"encoding/json"
	"html/template"

	"strings"

	"bitbucket.org/engineerbetter/concourse-up/aws"
	"bitbucket.org/engineerbetter/concourse-up/util"
)

const terraformStateFileName = "terraform.tfstate"
const configBucketS3Region = "eu-west-1"
const configFilePath = "config.json"

type Config struct {
	PublicKey   template.HTML `json:"public_key"`
	PrivateKey  template.HTML `json:"private_key"`
	Region      string        `json:"region"`
	Deployment  string        `json:"deployment"`
	TFStatePath string        `json:"tf_state_path"`
}

type IClient interface {
	Load(deployment string) (*Config, error)
	LoadOrCreate(deployment string) (*Config, error)
}

type Client struct {
	deployment string
}

func NewClient(deployment string) Client {
	return Client{
		deployment: deployment,
	}
}

func (client *Client) Load(deployment string) (*Config, error) {
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

func (client *Client) LoadOrCreate(deployment string) (*Config, error) {
	defaultConfigBytes, err := generateDefaultConfig(deployment, configBucketS3Region)
	if err != nil {
		return nil, err
	}

	if err := aws.EnsureBucketExists(deployment, configBucketS3Region); err != nil {
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

func generateDefaultConfig(deployment, region string) ([]byte, error) {
	privateKey, publicKey, err := util.MakeSSHKeyPair()
	if err != nil {
		return nil, err
	}

	conf := Config{
		PublicKey:   template.HTML(strings.TrimSpace(string(publicKey))),
		PrivateKey:  template.HTML(strings.TrimSpace(string(privateKey))),
		Deployment:  deployment,
		TFStatePath: terraformStateFileName,
		Region:      region,
	}

	return json.Marshal(&conf)
}
