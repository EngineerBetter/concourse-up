package concourse_test

import (
	"errors"
	"fmt"
	"io"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/director"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/testsupport"
	"github.com/aws/aws-sdk-go/service/route53"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Client", func() {
	var buildClient func() concourse.IClient
	var actions []string
	var stdout *gbytes.Buffer
	var stderr *gbytes.Buffer
	var deleteBoshDirectorError error
	var terraformMetadata *terraform.Metadata
	var args *config.DeployArgs
	var exampleConfig *config.Config
	var ipChecker func() (string, error)

	certGenerator := func(c func(u *certs.User) (certs.AcmeClient, error), caName string, ip ...string) (*certs.Certs, error) {
		actions = append(actions, fmt.Sprintf("generating cert ca: %s, cn: %s", caName, ip))
		return &certs.Certs{
			CACert: []byte("----EXAMPLE CERT----"),
		}, nil
	}

	awsClient := &testsupport.FakeAWSClient{
		FakeFindLongestMatchingHostedZone: func(subdomain string, listHostedZones func() ([]*route53.HostedZone, error)) (string, string, error) {
			if subdomain == "ci.google.com" {
				return "google.com", "ABC123", nil
			}

			return "", "", errors.New("hosted zone not found")
		},
		FakeDeleteVMsInVPC: func(vpcID string) ([]*string, error) {
			actions = append(actions, fmt.Sprintf("deleting vms in %s", vpcID))
			return nil, nil
		},
		FakeDeleteVolumes: func(volumesToDelete []*string, deleteVolume func(ec2Client iaas.IEC2, volumeID *string) error, newEC2Client func() (iaas.IEC2, error)) error {
			return nil
		},
	}

	fakeFlyClient := &testsupport.FakeFlyClient{
		FakeSetDefaultPipeline: func(deployArgs *config.DeployArgs, config *config.Config, allowFlyVersionDiscrepancy bool) error {
			actions = append(actions, "setting default pipeline")
			return nil
		},
		FakeCleanup: func() error {
			return nil
		},
		FakeCanConnect: func() (bool, error) {
			return false, nil
		},
	}

	BeforeEach(func() {
		args = &config.DeployArgs{
			AWSRegion:   "eu-west-1",
			DBSize:      "small",
			DBSizeIsSet: false,
		}

		terraformMetadata = &terraform.Metadata{
			ATCPublicIP:              terraform.MetadataStringValue{Value: "77.77.77.77"},
			ATCSecurityGroupID:       terraform.MetadataStringValue{Value: "sg-999"},
			BlobstoreBucket:          terraform.MetadataStringValue{Value: "blobs.aws.com"},
			BlobstoreSecretAccessKey: terraform.MetadataStringValue{Value: "abc123"},
			BlobstoreUserAccessKeyID: terraform.MetadataStringValue{Value: "abc123"},
			BoshDBAddress:            terraform.MetadataStringValue{Value: "rds.aws.com"},
			BoshDBPort:               terraform.MetadataStringValue{Value: "5432"},
			BoshSecretAccessKey:      terraform.MetadataStringValue{Value: "abc123"},
			BoshUserAccessKeyID:      terraform.MetadataStringValue{Value: "abc123"},
			DirectorKeyPair:          terraform.MetadataStringValue{Value: "-- KEY --"},
			DirectorPublicIP:         terraform.MetadataStringValue{Value: "99.99.99.99"},
			DirectorSecurityGroupID:  terraform.MetadataStringValue{Value: "sg-123"},
			NatGatewayIP:             terraform.MetadataStringValue{Value: "88.88.88.88"},
			PrivateSubnetID:          terraform.MetadataStringValue{Value: "sn-private-123"},
			PublicSubnetID:           terraform.MetadataStringValue{Value: "sn-public-123"},
			VMsSecurityGroupID:       terraform.MetadataStringValue{Value: "sg-456"},
			VPCID:                    terraform.MetadataStringValue{Value: "vpc-112233"},
		}

		deleteBoshDirectorError = nil
		actions = []string{}
		exampleConfig = &config.Config{
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
			Region:            "eu-west-1",
			Deployment:        "concourse-up-happymeal",
			Project:           "happymeal",
			TFStatePath:       "example-path",
			DirectorUsername:  "admin",
			DirectorPassword:  "secret123",
			RDSUsername:       "admin",
			RDSPassword:       "s3cret",
			ConcoursePassword: "s3cret",
			ConcourseUsername: "admin",
			RDSInstanceClass:  "db.t2.medium",
		}

		configClient := &testsupport.FakeConfigClient{
			FakeLoadOrCreate: func(deployArgs *config.DeployArgs) (*config.Config, bool, error) {
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
			FakeDeleteAll: func(config *config.Config) error {
				actions = append(actions, "deleting config")
				return nil
			},
		}

		terraformClientFactory := func(iaas string, config *config.Config, stdout, stderr io.Writer) (terraform.IClient, error) {
			return &testsupport.FakeTerraformClient{
				FakeApply: func(dryrun bool) error {
					Expect(dryrun).To(BeFalse())
					actions = append(actions, fmt.Sprintf("applying terraform, db size: %s", config.RDSInstanceClass))
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

		boshClientFactory := func(config *config.Config, metadata *terraform.Metadata, director director.IClient, stdout, stderr io.Writer) (bosh.IClient, error) {
			return &testsupport.FakeBoshClient{
				FakeDeploy: func(stateFileBytes, credsFileBytes []byte, detach bool) ([]byte, []byte, error) {
					if detach {
						actions = append(actions, "deploying director in self-update mode")
					} else {
						actions = append(actions, "deploying director")
					}
					return nil, nil, nil
				},
				FakeDelete: func([]byte) ([]byte, error) {
					actions = append(actions, "deleting director")
					return nil, deleteBoshDirectorError
				},
				FakeCleanup: func() error {
					actions = append(actions, "cleaning up bosh init")
					return nil
				},
			}, nil
		}

		ipChecker = func() (string, error) {
			return "192.0.2.0", nil
		}

		stdout = gbytes.NewBuffer()
		stderr = gbytes.NewBuffer()

		buildClient = func() concourse.IClient {
			return concourse.NewClient(
				awsClient,
				terraformClientFactory,
				boshClientFactory,
				func(fly.Credentials, io.Writer, io.Writer) (fly.IClient, error) {
					return fakeFlyClient, nil
				},
				certGenerator,
				configClient,
				args,
				stdout,
				stderr,
				ipChecker,
				testsupport.NewFakeAcmeClient,
			)
		}
	})

	Describe("Deploy", func() {
		It("Prints a warning about changing the sourceIP", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr).To(gbytes.Say("WARNING: allowing access from local machine"))
		})

		Context("When a custom domain is required", func() {
			It("Prints a warning about adding a DNS record", func() {
				args.Domain = "ci.google.com"

				client := buildClient()
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				Expect(stderr).To(gbytes.Say("WARNING: adding record ci.google.com to Route53 hosted zone google.com ID: ABC123"))
			})
		})

		It("Loads of creates config file", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("loading or creating config file"))
		})

		It("Generates the correct terraform infrastructure", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("applying terraform, db size: db.t2.medium"))
		})

		It("Cleans up the correct terraform client", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("cleaning up terraform client"))
		})

		It("Generates certificates for bosh", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("generating cert ca: concourse-up-happymeal, cn: [99.99.99.99 10.0.0.6]"))
		})

		It("Generates certificates for concourse", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("generating cert ca: concourse-up-happymeal, cn: [77.77.77.77]"))
		})

		It("Sets the director public IP on the config", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(exampleConfig.DirectorPublicIP).To(Equal("99.99.99.99"))
		})

		Context("When the user tries to change the region of an existing deployment", func() {
			It("Returns a meaningful error message", func() {
				args.AWSRegion = "eu-central-1"

				client := buildClient()
				err := client.Deploy()
				Expect(err).To(MatchError("found previous deployment in eu-west-1. Refusing to deploy to eu-central-1 as changing regions for existing deployments is not supported"))
			})
		})

		Context("When a custom DB instance size is provided", func() {
			It("Deploys that instance type to TF", func() {
				args.DBSize = "large"
				args.DBSizeIsSet = true

				client := buildClient()
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				Expect(actions).To(ContainElement("applying terraform, db size: db.m4.large"))
			})
		})

		Context("When a custom DB instance size is not provided", func() {
			It("Does not override the existing DB size", func() {
				args.DBSize = "small"
				args.DBSizeIsSet = false

				client := buildClient()
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				Expect(actions).To(ContainElement("applying terraform, db size: db.t2.medium"))
			})
		})

		Context("When a custom domain is required", func() {
			It("Generates certificates for that domain and not the public IP", func() {
				args.Domain = "ci.google.com"

				client := buildClient()
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				Expect(actions).To(ContainElement("generating cert ca: concourse-up-happymeal, cn: [ci.google.com]"))
			})
		})

		It("Updates the config", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("updating config file"))
		})

		It("Deploys the director", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("deploying director"))
		})

		It("Sets the default pipeline, after deploying the bosh director", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions[8]).To(Equal("deploying director"))
		})

		Context("When running in self-update mode and the concourse is already deployed", func() {
			It("Sets the default pipeline, before deploying the bosh director", func() {
				fakeFlyClient.FakeCanConnect = func() (bool, error) {
					return true, nil
				}
				args.SelfUpdate = true

				client := buildClient()
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				Expect(actions[8]).To(Equal("deploying director in self-update mode"))
			})
		})

		It("Saves the bosh state", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("storing config asset: director-state.json"))
		})

		It("Cleans up the director", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("cleaning up bosh init"))
		})

		It("Warns about access to local machine", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Eventually(stderr).Should(gbytes.Say("WARNING: allowing access from local machine"))
		})

		It("Prints the bosh credentials", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())
			Eventually(stdout).Should(gbytes.Say("DEPLOY SUCCESSFUL"))
			Eventually(stdout).Should(gbytes.Say("fly --target happymeal login --insecure --concourse-url https://77.77.77.77 --username admin --password s3cret"))
		})

		Context("When a custom cert is provided", func() {
			It("Prints the correct domain and not suggest using --insecure", func() {
				args.Domain = "ci.google.com"
				args.TLSCert = "--- CERTIFICATE ---"
				args.TLSKey = "--- KEY ---"

				client := buildClient()
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())
				Eventually(stdout).Should(gbytes.Say("DEPLOY SUCCESSFUL"))
				Eventually(stdout).Should(gbytes.Say("fly --target happymeal login --concourse-url https://ci.google.com --username admin --password s3cret"))
			})
		})

		Context("When an existing config is loaded", func() {
			It("Notifies the user", func() {
				client := buildClient()
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				Eventually(stdout).Should(gbytes.Say("USING PREVIOUS DEPLOYMENT CONFIG"))
			})
		})

		Context("When a metadata field is missing", func() {
			It("Returns an error", func() {
				terraformMetadata.DirectorKeyPair = terraform.MetadataStringValue{Value: ""}
				client := buildClient()
				err := client.Deploy()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("director_key_pair"))
				Expect(err.Error()).To(ContainSubstring("non zero value required"))
			})
		})
	})

	Describe("Destroy", func() {
		It("Loads the config file", func() {
			client := buildClient()
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("loading config file"))
		})

		It("Deletes the vms in the vpcs", func() {
			client := buildClient()
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("deleting vms in vpc-112233"))
		})

		It("Destroys the terraform infrastructure", func() {
			client := buildClient()
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("destroying terraform"))
		})

		It("Cleans up the terraform client", func() {
			client := buildClient()
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("cleaning up terraform client"))
		})

		It("Deletes the config", func() {
			client := buildClient()
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("deleting config"))
		})

		It("Prints a destroy success message", func() {
			client := buildClient()
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Eventually(stdout).Should(gbytes.Say("DESTROY SUCCESSFUL"))
		})

		Context("When there is an error deleting the bosh director", func() {
			BeforeEach(func() {
				deleteBoshDirectorError = errors.New("some error")
			})

			It("Continues the error", func() {
				client := buildClient()
				err := client.Destroy()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
