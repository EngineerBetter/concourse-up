package concourse

import (
	"fmt"
	"io"

	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
	"bitbucket.org/engineerbetter/concourse-up/util"
)

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

func Deploy(name, region string, terraformApplier terraform.Applier, configLoader config.Loader, stdout, stderr io.Writer) error {
	deployment := fmt.Sprintf("engineerbetter-concourseup-%s", name)

	config, err := configLoader(deployment, region)
	if err != nil {
		return err
	}

	terraformConfig, err := util.RenderTemplate(template, config)
	if err != nil {
		return err
	}
	return terraformApplier(terraformConfig, stdout, stderr)
}
