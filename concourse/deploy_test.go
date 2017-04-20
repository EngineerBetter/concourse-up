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

var _ = Describe("Deploy", func() {
	It("Generates the correct terraform infrastructure", func() {
		var appliedTFConfig []byte
		storedAssetPaths := []string{}

		client := &FakeConfigClient{
			FakeLoadOrCreate: func(project string) (*config.Config, error) {
				return &config.Config{
					PublicKey:   "example-public-key",
					PrivateKey:  "example-private-key",
					Region:      "eu-west-1",
					Deployment:  fmt.Sprintf("concourse-up-%s", project),
					Project:     project,
					TFStatePath: "example-path",
				}, nil
			},
			FakeStoreAsset: func(project, filename string, contents []byte) error {
				storedAssetPaths = append(storedAssetPaths, filename)
				return nil
			},
		}

		applied := false
		cleanedUp := false

		terraformClientFactory := func(config []byte, stdout, stderr io.Writer) (terraform.IClient, error) {
			appliedTFConfig = config
			return &FakeTerraformClient{
				FakeApply: func() error {
					applied = true
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

		boshInitDeployed := false
		boshInitClientFactory := func(manifestPath string, stdout, stderr io.Writer) bosh.IBoshInitClient {
			return &FakeBoshInitClient{
				FakeDeploy: func() ([]byte, error) {
					boshInitDeployed = true
					return []byte{}, nil
				},
			}
		}

		err := Deploy("happymeal", "eu-west-1", terraformClientFactory, boshInitClientFactory, client, os.Stdout, os.Stderr)
		Expect(err).ToNot(HaveOccurred())

		Expect(string(appliedTFConfig)).To(ContainSubstring("concourse-up-happymeal"))
		Expect(applied).To(BeTrue())
		Expect(cleanedUp).To(BeTrue())
		Expect(boshInitDeployed).To(BeTrue())
		Expect(storedAssetPaths).To(ContainElement("director.yml"))
		Expect(storedAssetPaths).To(ContainElement("director-state.json"))
	})
})
