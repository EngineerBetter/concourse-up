package config

import (
	"encoding/json"
	"html/template"

	"strings"

	"bitbucket.org/engineerbetter/concourse-up/aws"
	"bitbucket.org/engineerbetter/concourse-up/util"
)

const terraformStateFileName = "terraform.tfstate"

type Config struct {
	PublicKey   template.HTML `json:"public_key"`
	PrivateKey  template.HTML `json:"private_key"`
	Region      string        `json:"region"`
	Deployment  string        `json:"deployment"`
	TFStatePath string        `json:"tf_state_path"`
}

type Loader func(deployment, region string) (*Config, error)

func Load(deployment, region string) (*Config, error) {
	defaultConfigBytes, err := generateDefaultConfig(deployment, region)
	if err != nil {
		return nil, err
	}

	if err := aws.EnsureBucketExists(deployment, region); err != nil {
		return nil, err
	}

	configBytes, err := aws.EnsureFileExists(deployment, "config.json", region, defaultConfigBytes)
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
