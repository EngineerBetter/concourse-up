package concourse

import (
	"io"

	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
	"bitbucket.org/engineerbetter/concourse-up/util"
)

// Deploy deploys a concourse instance
func Deploy(name, region string, terraformApplier terraform.Applier, configClient config.IClient, stdout, stderr io.Writer) error {
	config, err := configClient.LoadOrCreate(name)
	if err != nil {
		return err
	}

	terraformConfig, err := util.RenderTemplate(template, config)
	if err != nil {
		return err
	}
	return terraformApplier(terraformConfig, stdout, stderr)
}

const template = `
terraform {
	backend "s3" {
		bucket = "<% .Deployment %>"
		key    = "<% .TFStatePath %>"
		region = "<% .Region %>"
	}
}

provider "aws" {
	region = "<% .Region %>"
}

resource "aws_key_pair" "deployer" {
	key_name_prefix = "<% .Deployment %>-"
	public_key      = "<% .PublicKey %>"
}
`
