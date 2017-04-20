package concourse_test

import (
	"fmt"
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
		client.FakeLoad = func(project string) (*config.Config, error) {
			return &config.Config{
				PublicKey:   "example-public-key",
				PrivateKey:  "example-private-key",
				Region:      "eu-west-1",
				Deployment:  fmt.Sprintf("concourse-up-%s", project),
				Project:     project,
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
