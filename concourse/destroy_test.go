package concourse_test

import (
	"io"
	"os"

	. "bitbucket.org/engineerbetter/concourse-up/concourse"
	"bitbucket.org/engineerbetter/concourse-up/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Destroy", func() {
	It("Destroys the terraform infrastructure", func() {
		var destroyedTFConfig []byte

		client := &FakeConfigClient{}
		client.FakeLoad = func(deployment string) (*config.Config, error) {
			return &config.Config{
				PublicKey:   "example-public-key",
				PrivateKey:  "example-private-key",
				Region:      "eu-west-1",
				Deployment:  deployment,
				TFStatePath: "example-path",
			}, nil
		}

		applier := func(config []byte, stdout, stderr io.Writer) error {
			destroyedTFConfig = config
			return nil
		}

		err := Destroy("happymeal", applier, client, os.Stdout, os.Stderr)
		Expect(err).ToNot(HaveOccurred())

		Expect(string(destroyedTFConfig)).To(Equal(`
terraform {
	backend "s3" {
		bucket = "engineerbetter-concourseup-happymeal"
		key    = "example-path"
		region = "eu-west-1"
	}
}

provider "aws" {
	region = "eu-west-1"
}

resource "aws_key_pair" "deployer" {
	key_name_prefix = "engineerbetter-concourseup-happymeal-"
	public_key      = "example-public-key"
}
`))
	})
})
