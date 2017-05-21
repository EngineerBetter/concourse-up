package terraform_test

import (
	"fmt"

	"github.com/EngineerBetter/concourse-up/aws"
	. "github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Apply & Destroy", func() {
	var bucket string

	configTemplate := `
terraform {
  backend "s3" {
    bucket = "<% .Bucket %>"
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

	BeforeEach(func() {
		bucket = fmt.Sprintf("concourse-up-integration-tests-%s", util.GeneratePassword())

		err := aws.EnsureBucketExists(bucket, "eu-west-1")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := aws.DeleteVersionedBucket(bucket, "eu-west-1")
		Expect(err).ToNot(HaveOccurred())
	})

	It("Works", func() {
		params := struct {
			Bucket string
		}{
			Bucket: bucket,
		}

		config, err := util.RenderTemplate(configTemplate, &params)
		Expect(err).ToNot(HaveOccurred())

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
