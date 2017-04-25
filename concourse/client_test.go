package concourse_test

import (
	"errors"
	"fmt"
	"io"

	"bitbucket.org/engineerbetter/concourse-up/bosh"
	"bitbucket.org/engineerbetter/concourse-up/certs"
	"bitbucket.org/engineerbetter/concourse-up/concourse"
	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/director"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Client", func() {
	var client concourse.IClient
	var actions []string
	var stdout *gbytes.Buffer
	var stderr *gbytes.Buffer
	var deleteBoshDirectorError error
	var terraformMetadata *terraform.Metadata

	certGenerator := func(caName string, ip string) (*certs.Certs, error) {
		actions = append(actions, fmt.Sprintf("generating cert ca: %s, ip: %s", caName, ip))
		return &certs.Certs{
			CACert: []byte("----EXAMPLE CERT----"),
		}, nil
	}

	BeforeEach(func() {
		terraformMetadata = &terraform.Metadata{
			DirectorPublicIP:         terraform.MetadataStringValue{Value: "99.99.99.99"},
			DirectorKeyPair:          terraform.MetadataStringValue{Value: "-- KEY --"},
			DirectorSecurityGroupID:  terraform.MetadataStringValue{Value: "sg-123"},
			VMsSecurityGroupID:       terraform.MetadataStringValue{Value: "sg-456"},
			DefaultSubnetID:          terraform.MetadataStringValue{Value: "sn-123"},
			BoshDBPort:               terraform.MetadataStringValue{Value: "5432"},
			BoshDBAddress:            terraform.MetadataStringValue{Value: "rds.aws.com"},
			BoshDBUsername:           terraform.MetadataStringValue{Value: "admin"},
			BoshDBPassword:           terraform.MetadataStringValue{Value: "s3cret"},
			BoshUserAccessKeyID:      terraform.MetadataStringValue{Value: "abc123"},
			BoshSecretAccessKey:      terraform.MetadataStringValue{Value: "abc123"},
			BlobstoreBucket:          terraform.MetadataStringValue{Value: "blobs.aws.com"},
			BlobstoreUserAccessKeyID: terraform.MetadataStringValue{Value: "abc123"},
			BlobstoreSecretAccessKey: terraform.MetadataStringValue{Value: "abc123"},
			ELBSecurityGroupID:       terraform.MetadataStringValue{Value: "sg-789"},
			ELBName:                  terraform.MetadataStringValue{Value: "elb-123"},
		}
		deleteBoshDirectorError = nil
		actions = []string{}
		exampleConfig := &config.Config{
			PublicKey:        "example-public-key",
			PrivateKey:       "example-private-key",
			Region:           "eu-west-1",
			Deployment:       "concourse-up-happymeal",
			Project:          "happymeal",
			TFStatePath:      "example-path",
			DirectorUsername: "admin",
			DirectorPassword: "secret123",
		}
		configClient := &FakeConfigClient{
			FakeLoadOrCreate: func() (*config.Config, bool, error) {
				actions = append(actions, "loading or creating config file")
				return exampleConfig, false, nil
			},
			FakeLoad: func() (*config.Config, error) {
				actions = append(actions, "loading config file")
				return exampleConfig, nil
			},
			FakeUpdate: func(config *config.Config) error {
				actions = append(actions, "updating config file")
				return nil
			},
			FakeStoreAsset: func(filename string, contents []byte) error {
				actions = append(actions, fmt.Sprintf("storing config asset: %s", filename))
				return nil
			},
			FakeHasAsset: func(filename string) (bool, error) {
				return false, nil
			},
		}

		terraformClientFactory := func(config []byte, stdout, stderr io.Writer) (terraform.IClient, error) {
			return &FakeTerraformClient{
				FakeApply: func() error {
					actions = append(actions, "applying terraform")
					return nil
				},
				FakeDestroy: func() error {
					actions = append(actions, "destroying terraform")
					return nil
				},
				FakeOutput: func() (*terraform.Metadata, error) {
					actions = append(actions, "fetching terraform metadata")
					return terraformMetadata, nil
				},
				FakeCleanup: func() error {
					actions = append(actions, "cleaning up terraform client")
					return nil
				},
			}, nil
		}

		boshClientFactory := func(config *config.Config, metadata *terraform.Metadata, director director.IClient) bosh.IClient {
			return &FakeBoshClient{
				FakeDeploy: func([]byte) ([]byte, error) {
					actions = append(actions, "deploying director")
					return []byte{}, nil
				},
				FakeDelete: func([]byte) ([]byte, error) {
					actions = append(actions, "deleting director")
					return []byte{}, deleteBoshDirectorError
				},
				FakeCleanup: func() error {
					actions = append(actions, "cleaning up bosh init")
					return nil
				},
			}
		}

		stdout = gbytes.NewBuffer()
		stderr = gbytes.NewBuffer()

		client = concourse.NewClient(
			terraformClientFactory,
			boshClientFactory,
			certGenerator,
			configClient,
			stdout,
			stderr,
		)
	})

	Describe("Deploy", func() {
		It("Loads of creates config file", func() {
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("loading or creating config file"))
		})

		It("Generates the correct terraform infrastructure", func() {
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("applying terraform"))
		})

		It("Cleans up the correct terraform client", func() {
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("cleaning up terraform client"))
		})

		It("Generates certificates", func() {
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("generating cert ca: concourse-up-happymeal, ip: 99.99.99.99"))
		})

		It("Updates the config", func() {
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("updating config file"))
		})

		It("Deploys the director", func() {
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("deploying director"))
		})

		It("Cleans up the director", func() {
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("cleaning up bosh init"))
		})

		It("Warns about access to local machine", func() {
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Eventually(stderr).Should(gbytes.Say("WARNING: allowing access from local machine"))
		})

		It("Prints the bosh credentials", func() {
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())
			Eventually(stdout).Should(gbytes.Say(
				"DEPLOY SUCCESSFUL. Bosh connection credentials:\n\tIP Address: 99.99.99.99\n\tUsername: admin\n\tPassword: secret123\n\tCA Cert:\n\t\t----EXAMPLE CERT----"))
		})

		Context("When an existing config is loaded", func() {
			It("Notifies the user", func() {
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				Eventually(stdout).Should(gbytes.Say("USING PREVIOUS DEPLOYMENT CONFIG"))
			})
		})

		Context("When a metadata field is missing", func() {
			It("Returns an error", func() {
				terraformMetadata.DirectorKeyPair = terraform.MetadataStringValue{Value: ""}
				err := client.Deploy()
				Expect(err.Error()).To(ContainSubstring("director_key_pair"))
				Expect(err.Error()).To(ContainSubstring("non zero value required"))
			})
		})
	})

	Describe("Destroy", func() {
		It("Loads the config file", func() {
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("loading config file"))
		})
		It("Deletes the director", func() {
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("deleting director"))
		})

		It("Cleans up the director", func() {
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("cleaning up bosh init"))
		})

		It("Destroys the terraform infrastructure", func() {
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("destroying terraform"))
		})

		It("Cleans up the terraform client", func() {
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("cleaning up terraform client"))
		})

		It("Prints a destroy success message", func() {
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Eventually(stdout).Should(gbytes.Say("DESTROY SUCCESSFUL"))
		})

		Context("When there is an error deleting the bosh director", func() {
			BeforeEach(func() {
				deleteBoshDirectorError = errors.New("some error")
			})

			It("Still attemps to destroy the terraform", func() {
				err := client.Destroy()
				Expect(err).ToNot(HaveOccurred())

				Expect(actions).To(ContainElement("destroying terraform"))
			})

			It("Prints a warning", func() {
				err := client.Destroy()
				Expect(err).ToNot(HaveOccurred())

				Eventually(stderr).Should(gbytes.Say("Warning error deleting bosh director. Continuing with terraform deletion."))
			})
		})
	})
})
