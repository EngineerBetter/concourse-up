package concourse_test

import (
	"errors"
	"fmt"
	"io"

	"bitbucket.org/engineerbetter/concourse-up/bosh"
	"bitbucket.org/engineerbetter/concourse-up/certs"
	"bitbucket.org/engineerbetter/concourse-up/concourse"
	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/db"
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
			PublicKey: "example-public-key",
			PrivateKey: `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA2spClkDkFfy2c91Z7N3AImPf0v3o5OoqXUS6nE2NbV2bP/o7
Oa3KnpzeQ5DBmW3EW7tuvA4bAHxPuk25T9tM8jiItg0TNtMlxzFYVxFq8jMmokEi
sMVbjh9XIZptyZHbZzsJsbaP/xOGHSQNYwH/7qnszbPKN82zGwrsbrGh1hRMATbU
S+oor1XTLWGKuLs72jWJK864RW/WiN8eNfk7on1Ugqep4hnXLQjrgbOOxeX7/Pap
VEExC63c1FmZjLnOc6mLbZR07qM9jj5fmR94DzcliF8SXIvp6ERDMYtnI7gAC4XA
ZgATsS0rkb5t7dxsaUl0pHfU9HlhbMciN3bJrwIDAQABAoIBADQIWiGluRjJixKv
F83PRvxmyDpDjHm0fvLDf6Xgg7v4wQ1ME326KS/jmrBy4rf8dPBj+QfcSuuopMVn
6qRlQT1x2IGDRoiJWriusZWzXL3REGUSHI/xv75jEbO6KFYBzC4Wyk1rX3+IQyL3
Cf/738QAwYKCOZtf3jKWPHhu4lAo/rq6FY/okWMybaAXajCTF2MgJcmMm73jIgk2
6A6k9Cobs7XXNZVogAUsHU7bgnkfxYgz34UTZu0FDQRGf3MpHeWp32dhw9UAaFz7
nfoBVxU1ppqM4TCdXvezKgi8QV6imvDyD67/JNUn0B06LKMbAIK/mffA9UL8CXkc
YSj5AIECgYEA/b9MVy//iggMAh+DZf8P+fS79bblVamdHsU8GvHEDdIg0lhBl3pQ
Nrpi63sXVIMz52BONKLJ/c5/wh7xIiApOMcu2u+2VjN00dqpivasERf0WbgSdvMS
Gi+0ofG0kF94W7z8Z1o9rT4Wn9wxuqkRLLp3A5CkpjzlEnPVoW9X2I8CgYEA3LuD
ZpL2dRG5sLA6ahrJDZASk4cBaQGcYpx/N93dB3XlCTguPIJL0hbt1cwwhgCQh6cu
B0mDWsiQIMwET7bL5PX37c1QBh0rPqQsz8/T7jNEDCnbWDWQSaR8z6sGJCWEkWzo
AtzvPkTj75bDsYG0KVlYMfNJyYHZJ5ECJ08ZTOECgYEA5rLF9X7uFdC7GjMMg+8h
119qhDuExh0vfIpV2ylz1hz1OkiDWfUaeKd8yBthWrTuu64TbEeU3eyguxzmnuAe
mkB9mQ/X9wdRbnofKviZ9/CPeAKixwK3spcs4w+d2qTyCHYKBO1GpfuNFkpb7BlK
RCBDlDotd/ZlTiGCWQOiGoECgYEAmM/sQUf+/b8+ubbXSfuvMweKBL5TWJn35UEI
xemACpkw7fgJ8nQV/6VGFFxfP3YGmRNBR2Q6XtA5D6uOVI1tjN5IPUaFXyY0eRJ5
v4jW5LJzKqSTqPa0JHeOvMpe3wlmRLOLz+eabZaN4qGSa0IrMvEaoMIYVDvj1YOL
ZSFal6ECgYBDXbrmvF+G5HoASez0WpgrHxf3oZh+gP40rzwc94m9rVP28i8xTvT9
5SrvtzwjMsmQPUM/ttaBnNj1PvmOTTmRhXVw5ztAN9hhuIwVm8+mECFObq95NIgm
sWbB3FCIsym1FXB+eRnVF3Y15RwBWWKA5RfwUNpEXFxtv24tQ8jrdA==
-----END RSA PRIVATE KEY-----`,
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
			FakeDeleteAsset: func(filename string) error {
				actions = append(actions, fmt.Sprintf("deleting config asset: %s", filename))
				return nil
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

		boshClientFactory := func(config *config.Config, metadata *terraform.Metadata, director director.IClient, dbRunner db.Runner) bosh.IClient {
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

		It("Saves the bosh state", func() {
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("storing config asset: director-state.json"))
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

		It("Deletes the bosh state", func() {
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("deleting config asset: director-state.json"))
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

			It("Returns the error", func() {
				err := client.Destroy()
				Expect(err).To(MatchError("some error"))
			})
		})
	})
})
