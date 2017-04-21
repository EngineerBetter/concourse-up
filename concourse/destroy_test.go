package concourse_test

import (
	"io"
	"os"

	. "bitbucket.org/engineerbetter/concourse-up/concourse"
	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/director"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Destroy", func() {
	It("Destroys the terraform infrastructure", func() {
		var destroyedTFConfig []byte

		client := &FakeConfigClient{
			FakeLoad: func() (*config.Config, error) {
				return &config.Config{
					PublicKey:   "example-public-key",
					PrivateKey:  "example-private-key",
					Region:      "eu-west-1",
					Deployment:  "concourse-up-happymeal",
					Project:     "happymeal",
					TFStatePath: "example-path",
				}, nil
			},
			FakeHasAsset: func(filename string) (bool, error) {
				return false, nil
			},
		}

		destroyedTerraform := false
		cleanedUpTerraform := false

		clientFactory := func(config []byte, stdout, stderr io.Writer) (terraform.IClient, error) {
			destroyedTFConfig = config
			return &FakeTerraformClient{
				FakeDestroy: func() error {
					destroyedTerraform = true
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
					cleanedUpTerraform = true
					return nil
				},
			}, nil
		}

		boshInitDeleted := false
		cleanedUpBoshInit := false
		boshInitClientFactory := func(manifestBytes, stateFileBytes, keyfileBytes []byte, stdout, stderr io.Writer) (director.IBoshInitClient, error) {
			return &FakeBoshInitClient{
				FakeDelete: func() error {
					boshInitDeleted = true
					return nil
				},
				FakeCleanup: func() error {
					cleanedUpBoshInit = true
					return nil
				},
			}, nil
		}

		concourseClient := NewClient(clientFactory, boshInitClientFactory, client, os.Stdout, os.Stderr)
		err := concourseClient.Destroy()
		Expect(err).ToNot(HaveOccurred())

		Expect(string(destroyedTFConfig)).To(ContainSubstring("concourse-up-happymeal"))
		Expect(destroyedTerraform).To(BeTrue())
		Expect(cleanedUpTerraform).To(BeTrue())
		Expect(boshInitDeleted).To(BeTrue())
		Expect(cleanedUpBoshInit).To(BeTrue())
	})
})
