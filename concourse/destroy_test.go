package concourse_test

import (
	"fmt"
	"io"
	"os"

	"bitbucket.org/engineerbetter/concourse-up/bosh"
	. "bitbucket.org/engineerbetter/concourse-up/concourse"
	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
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

		destroyed := false
		cleanedUp := false

		clientFactory := func(config []byte, stdout, stderr io.Writer) (terraform.IClient, error) {
			destroyedTFConfig = config
			return &FakeTerraformClient{
				FakeDestroy: func() error {
					destroyed = true
					return nil
				},
				FakeOutput: func() (*terraform.Metadata, error) {
					return &terraform.Metadata{
						BoshDBPort: terraform.MetadataStringValue{
							Value: "5432",
						},
					}, nil
				},
				FakeCleanup: func() error {
					cleanedUp = true
					return nil
				},
			}, nil
		}

		boshInitDeleted := false
		boshInitClientFactory := func(manifestPath string, stdout, stderr io.Writer) bosh.IBoshInitClient {
			return &FakeBoshInitClient{
				FakeDelete: func() error {
					boshInitDeleted = true
					return nil
				},
			}
		}

		err := Destroy("happymeal", clientFactory, boshInitClientFactory, client, os.Stdout, os.Stderr)
		Expect(err).ToNot(HaveOccurred())

		Expect(string(destroyedTFConfig)).To(ContainSubstring("concourse-up-happymeal"))
		Expect(destroyed).To(BeTrue())
		Expect(cleanedUp).To(BeTrue())
		Expect(boshInitDeleted).To(BeTrue())
	})
})
