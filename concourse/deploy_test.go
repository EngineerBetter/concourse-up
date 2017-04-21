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
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Deploy", func() {
	It("Generates the correct terraform infrastructure", func() {
		var appliedTFConfig []byte
		storedAssetPaths := []string{}

		client := &FakeConfigClient{
			FakeLoadOrCreate: func() (*config.Config, error) {
				return &config.Config{
					PublicKey:   "example-public-key",
					PrivateKey:  "example-private-key",
					Region:      "eu-west-1",
					Deployment:  "concourse-up-happymeal",
					Project:     "happymeal",
					TFStatePath: "example-path",
				}, nil
			},
			FakeStoreAsset: func(filename string, contents []byte) error {
				storedAssetPaths = append(storedAssetPaths, filename)
				return nil
			},
			FakeHasAsset: func(filename string) (bool, error) {
				return false, nil
			},
		}

		appliedTerraform := false
		cleanedUp := false

		terraformClientFactory := func(config []byte, stdout, stderr io.Writer) (terraform.IClient, error) {
			appliedTFConfig = config
			return &FakeTerraformClient{
				FakeApply: func() error {
					appliedTerraform = true
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
		cleanedUpBoshInit := false
		boshInitClientFactory := func(manifestBytes, stateFileBytes, keyfileBytes []byte, stdout, stderr io.Writer) (director.IBoshInitClient, error) {
			return &FakeBoshInitClient{
				FakeDeploy: func() ([]byte, error) {
					boshInitDeployed = true
					return []byte{}, nil
				},
				FakeCleanup: func() error {
					cleanedUpBoshInit = true
					return nil
				},
			}, nil
		}

		stdout := gbytes.NewBuffer()

		concourseClient := NewClient(terraformClientFactory, boshInitClientFactory, client, stdout, os.Stderr)
		err := concourseClient.Deploy()
		Expect(err).ToNot(HaveOccurred())

		Expect(string(appliedTFConfig)).To(ContainSubstring("concourse-up-happymeal"))
		Expect(appliedTerraform).To(BeTrue())
		Expect(cleanedUp).To(BeTrue())
		Expect(boshInitDeployed).To(BeTrue())
		Expect(storedAssetPaths).To(HaveLen(1))
		Expect(storedAssetPaths).To(ContainElement("director-state.json"))
		Expect(cleanedUpBoshInit).To(BeTrue())

		Eventually(stdout).Should(gbytes.Say("DEPLOY SUCCESSFUL"))
	})
})
