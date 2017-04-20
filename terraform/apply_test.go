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
  region = "eu-west-1"
}

resource "aws_iam_user" "example-user-2" {
  name = "example-2"
}

output "director_public_ip" {
  value = "example"
}

output "director_security_group_id" {
  value = "example"
}

output "director_subnet_id" {
  value = "example"
}
`
		stdout := gbytes.NewBuffer()
		stderr := gbytes.NewBuffer()

		c, err := NewClient([]byte(config), stdout, stderr)
		Expect(err).ToNot(HaveOccurred())

		defer c.Cleanup()

		err = c.Apply()
		Expect(err).ToNot(HaveOccurred())
		Eventually(stdout).Should(gbytes.Say("Apply complete!"))

		metadata, err := c.Output()
		Expect(err).ToNot(HaveOccurred())
		Expect(metadata.DirectorPublicIP.Value).To(Equal("example"))

		err = c.Destroy()
		Expect(err).ToNot(HaveOccurred())

		Eventually(stdout).Should(gbytes.Say("Destroy complete!"))
	})
})
