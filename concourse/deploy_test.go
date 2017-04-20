package concourse_test

import (
	"io"
	"os"

	. "bitbucket.org/engineerbetter/concourse-up/concourse"
	"bitbucket.org/engineerbetter/concourse-up/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deploy", func() {
	It("Generates the correct terraform infrastructure", func() {
		var appliedTFConfig []byte

		client := &FakeConfigClient{}
		client.FakeLoadOrCreate = func(deployment string) (*config.Config, error) {
			return &config.Config{
				PublicKey:   "example-public-key",
				PrivateKey:  "example-private-key",
				Region:      "eu-west-1",
				Deployment:  deployment,
				TFStatePath: "example-path",
			}, nil
		}

		applier := func(config []byte, stdout, stderr io.Writer) error {
			appliedTFConfig = config
			return nil
		}

		err := Deploy("happymeal", "eu-west-1", applier, client, os.Stdout, os.Stderr)
		Expect(err).ToNot(HaveOccurred())

		Expect(string(appliedTFConfig)).To(Equal(`
terraform {
	backend "s3" {
		bucket = "concourse-up-happymeal"
		key    = "example-path"
		region = "eu-west-1"
	}
}

provider "aws" {
	region = "eu-west-1"
}

resource "aws_key_pair" "deployer" {
	key_name_prefix = "concourse-up-happymeal-"
	public_key      = "example-public-key"
}
`))
	})
})
