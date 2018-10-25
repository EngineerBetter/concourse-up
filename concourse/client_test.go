package concourse_test

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/director"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/testsupport"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Client", func() {
	var buildClient func() concourse.IClient
	var buildClientOtherRegion func() concourse.IClient
	var actions []string
	var stdout *gbytes.Buffer
	var stderr *gbytes.Buffer
	var deleteBoshDirectorError error
	var args *config.DeployArgs
	var exampleConfig config.Config
	var ipChecker func() (string, error)
	var exampleDirectorCreds []byte

	type TerraformMetadata struct {
		ATCPublicIP              string
		ATCSecurityGroupID       string
		BlobstoreBucket          string
		BlobstoreSecretAccessKey string
		BlobstoreUserAccessKeyID string
		BoshDBAddress            string
		BoshDBPort               string
		BoshSecretAccessKey      string
		BoshUserAccessKeyID      string
		DirectorKeyPair          string
		DirectorPublicIP         string
		DirectorSecurityGroupID  string
		NatGatewayIP             string
		PrivateSubnetID          string
		PublicSubnetID           string
		VMsSecurityGroupID       string
		VPCID                    string
	}
	var terraformMetadata TerraformMetadata

	certGenerator := func(c func(u *certs.User) (certs.AcmeClient, error), caName string, ip ...string) (*certs.Certs, error) {
		actions = append(actions, fmt.Sprintf("generating cert ca: %s, cn: %s", caName, ip))
		return &certs.Certs{
			CACert: []byte("----EXAMPLE CERT----"),
		}, nil
	}

	awsClient := &testsupport.FakeAWSClient{
		FakeFindLongestMatchingHostedZone: func(subdomain string) (string, string, error) {
			if subdomain == "ci.google.com" {
				return "google.com", "ABC123", nil
			}

			return "", "", errors.New("hosted zone not found")
		},
		FakeCheckForWhitelistedIP: func(ip, securityGroup string) (bool, error) {
			actions = append(actions, "checking security group for IP")
			if ip == "1.2.3.4" {
				return false, nil
			}
			return true, nil
		},
		FakeDeleteVMsInVPC: func(vpcID string) ([]*string, error) {
			actions = append(actions, fmt.Sprintf("deleting vms in %s", vpcID))
			return nil, nil
		},
		FakeDeleteVolumes: func(volumesToDelete []*string, deleteVolume func(ec2Client iaas.IEC2, volumeID *string) error) error {
			return nil
		},
		FakeRegion: func() string {
			return "eu-west-1"
		},
	}

	otherRegionClient := &testsupport.FakeAWSClient{
		FakeRegion: func() string {
			return "eu-central-1"
		},
	}

	fakeFlyClient := &testsupport.FakeFlyClient{
		FakeSetDefaultPipeline: func(config config.Config, allowFlyVersionDiscrepancy bool) error {
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

		terraformMetadata = TerraformMetadata{
			ATCPublicIP:              "77.77.77.77",
			ATCSecurityGroupID:       "sg-999",
			BlobstoreBucket:          "blobs.aws.com",
			BlobstoreSecretAccessKey: "abc123",
			BlobstoreUserAccessKeyID: "abc123",
			BoshDBAddress:            "rds.aws.com",
			BoshDBPort:               "5432",
			BoshSecretAccessKey:      "abc123",
			BoshUserAccessKeyID:      "abc123",
			DirectorKeyPair:          "-- KEY --",
			DirectorPublicIP:         "99.99.99.99",
			DirectorSecurityGroupID:  "sg-123",
			NatGatewayIP:             "88.88.88.88",
			PrivateSubnetID:          "sn-private-123",
			PublicSubnetID:           "sn-public-123",
			VMsSecurityGroupID:       "sg-456",
			VPCID:                    "vpc-112233",
		}

		deleteBoshDirectorError = nil
		actions = []string{}
		exampleConfig = config.Config{
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

		exampleDirectorCreds = []byte("atc_password: s3cret")

		configClient := &testsupport.FakeConfigClient{
			FakeLoadOrCreate: func(deployArgs *config.DeployArgs) (config.Config, bool, bool, error) {
				actions = append(actions, "loading or creating config file")
				return exampleConfig, false, false, nil
			},
			FakeLoad: func() (config.Config, error) {
				actions = append(actions, "loading config file")
				return exampleConfig, nil
			},
			FakeDeleteAsset: func(filename string) error {
				actions = append(actions, fmt.Sprintf("deleting config asset: %s", filename))
				return nil
			},
			FakeUpdate: func(config config.Config) error {
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
			FakeDeleteAll: func(config config.Config) error {
				actions = append(actions, "deleting config")
				return nil
			},
		}
		terraformCLI := &testsupport.FakeCLI{
			FakeIAAS: func(name string) (terraform.InputVars, terraform.IAASMetadata) {
				fakeInputVars := &testsupport.FakeTerraformInputVars{
					FakeBuild: func(data map[string]interface{}) error {
						actions = append(actions, "building iaas environment")
						return nil
					},
				}
				fakeMetadata := &testsupport.FakeIAASMetadata{
					FakeGet: func(key string) (string, error) {
						actions = append(actions, fmt.Sprintf("looking for key %s", key))
						mm := reflect.ValueOf(&terraformMetadata)
						m := mm.Elem()
						mv := m.FieldByName(key)
						if !mv.IsValid() {
							return "", errors.New(key + " key not found")
						}
						return mv.String(), nil
					},
				}
				return fakeInputVars, fakeMetadata
			},
			FakeApply: func(conf terraform.InputVars, dryrun bool) error {
				Expect(dryrun).To(BeFalse())
				actions = append(actions, "applying terraform")
				return nil
			},
			FakeDestroy: func(conf terraform.InputVars) error {
				actions = append(actions, "destroying terraform")
				return nil
			},
			FakeBuildOutput: func(conf terraform.InputVars, metadata terraform.IAASMetadata) error {
				actions = append(actions, "initializing terraform metadata")
				return nil
			},
		}

		boshClientFactory := func(config config.Config, metadata terraform.IAASMetadata, director director.IClient, stdout, stderr io.Writer) (bosh.IClient, error) {
			return &testsupport.FakeBoshClient{
				FakeDeploy: func(stateFileBytes, credsFileBytes []byte, detach bool) ([]byte, []byte, error) {
					if detach {
						actions = append(actions, "deploying director in self-update mode")
					} else {
						actions = append(actions, "deploying director")
					}
					return nil, exampleDirectorCreds, nil
				},
				FakeDelete: func([]byte) ([]byte, error) {
					actions = append(actions, "deleting director")
					return nil, deleteBoshDirectorError
				},
				FakeCleanup: func() error {
					actions = append(actions, "cleaning up bosh init")
					return nil
				},
				FakeInstances: func() ([]bosh.Instance, error) {
					actions = append(actions, "listing bosh instances")
					return nil, nil
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
				terraformCLI,
				boshClientFactory,
				func(fly.Credentials, io.Writer, io.Writer, []byte) (fly.IClient, error) {
					return fakeFlyClient, nil
				},
				certGenerator,
				configClient,
				args,
				stdout,
				stderr,
				ipChecker,
				testsupport.NewFakeAcmeClient,
				"some version",
			)
		}

		buildClientOtherRegion = func() concourse.IClient {
			return concourse.NewClient(
				otherRegionClient,
				terraformCLI,
				boshClientFactory,
				func(fly.Credentials, io.Writer, io.Writer, []byte) (fly.IClient, error) {
					return fakeFlyClient, nil
				},
				certGenerator,
				configClient,
				args,
				stdout,
				stderr,
				ipChecker,
				testsupport.NewFakeAcmeClient,
				"some version",
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
				exampleConfig.Domain = "ci.google.com"

				client := buildClient()
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				Expect(stderr).To(gbytes.Say("WARNING: adding record ci.google.com to Route53 hosted zone google.com ID: ABC123"))
			})
		})
		It("Builds IAAS environment", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("building iaas environment"))
		})

		It("Loads or creates config file", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("loading or creating config file"))
		})

		It("Generates the correct terraform infrastructure", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("applying terraform"))
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

		Context("When the user tries to change the region of an existing deployment", func() {
			It("Returns a meaningful error message", func() {
				args.AWSRegion = "eu-central-1"

				client := buildClientOtherRegion()
				err := client.Deploy()
				Expect(err).To(MatchError("found previous deployment in eu-west-1. Refusing to deploy to eu-central-1 as changing regions for existing deployments is not supported"))
			})
		})

		Context("When a custom DB instance size is not provided", func() {
			It("Does not override the existing DB size", func() {
				args.DBSize = "small"
				args.DBSizeIsSet = false

				client := buildClient()
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				Expect(actions).To(ContainElement("applying terraform"))
			})
		})

		Context("When a custom domain is required", func() {
			It("Generates certificates for that domain and not the public IP", func() {
				exampleConfig.Domain = "ci.google.com"

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

			Expect(testsupport.CompareActions(actions, "deploying director", "setting default pipeline")).To(BeNumerically("<", 0))
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

				Expect(actions).To(ContainElement("deploying director in self-update mode"))
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
				exampleConfig.Domain = "ci.google.com"
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
	})

	Describe("Destroy", func() {
		It("Loads the config file", func() {
			client := buildClient()
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("loading config file"))
		})
		It("Builds IAAS environment", func() {
			client := buildClient()
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("building iaas environment"))
		})
		It("Loads terraform output", func() {
			client := buildClient()
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("initializing terraform metadata"))
		})
		It("Gets the vpc ID", func() {
			client := buildClient()
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("looking for key VPCID"))
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

	Describe("FetchInfo", func() {
		It("Loads the config file", func() {
			client := buildClient()
			_, err := client.FetchInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("loading config file"))
		})
		It("Builds IAAS environment", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("building iaas environment"))
		})

		It("Loads terraform output", func() {
			client := buildClient()
			_, err := client.FetchInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("initializing terraform metadata"))
		})

		It("Gets the director public IP", func() {
			client := buildClient()
			_, err := client.FetchInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("looking for key DirectorPublicIP"))
		})
		It("Gets the nat gateway IP", func() {
			client := buildClient()
			_, err := client.FetchInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("looking for key NatGatewayIP"))
		})
		It("Gets the director security group ID", func() {
			client := buildClient()
			_, err := client.FetchInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("looking for key DirectorSecurityGroupID"))
		})
		It("Checks that the IP is whitelisted", func() {
			client := buildClient()
			_, err := client.FetchInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("checking security group for IP"))
		})

		It("Retrieves the BOSH instances", func() {
			client := buildClient()
			_, err := client.FetchInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("listing bosh instances"))
		})

		Context("When the IP address isn't properly whitelisted", func() {
			BeforeEach(func() {
				ipChecker = func() (string, error) {
					return "1.2.3.4", nil
				}
			})

			It("Returns a meaningful error", func() {
				client := buildClient()
				_, err := client.FetchInfo()
				Expect(err).To(MatchError("Do you need to add your IP 1.2.3.4 to the concourse-up-happymeal-director security group (for ports 22, 6868, and 25555)?"))
			})
		})
	})
})
