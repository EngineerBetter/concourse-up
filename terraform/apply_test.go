package terraform_test

import (
	. "bitbucket.org/engineerbetter/concourse-up/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Apply & Destroy", func() {
	It("Works", func() {
		config := `
terraform {
  backend "s3" {
    bucket = "concourse-up-integration-tests"
    key    = "apply_test"
    region = "eu-west-1"
  }
}

provider "aws" {
  region     = "eu-west-1"
}

resource "aws_iam_user" "example-user-2" {
  name = "example-2"
}
`
		stdout := gbytes.NewBuffer()
		stderr := gbytes.NewBuffer()

		err := Apply(config, stdout, stderr)
		Expect(err).ToNot(HaveOccurred())

		Eventually(stdout).Should(gbytes.Say("Apply complete! Resources: 1 added, 0 changed, 0 destroyed."))

		err = Destroy(config, stdout, stderr)
		Expect(err).ToNot(HaveOccurred())

		Eventually(stdout).Should(gbytes.Say("Destroy complete! Resources: 1 destroyed."))
	})
})
