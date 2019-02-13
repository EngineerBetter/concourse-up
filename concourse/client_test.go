package concourse_test

import (
	"errors"
	"fmt"
	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/bosh/boshfakes"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/certs/certsfakes"
	"github.com/EngineerBetter/concourse-up/commands/deploy"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/concourse/concoursefakes"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/config/configfakes"
	"github.com/EngineerBetter/concourse-up/director"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/fly/flyfakes"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/iaas/iaasfakes"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/terraform/terraformfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/tjarratt/gcounterfeiter"
	"github.com/xenolf/lego/lego"
	"io"
	"io/ioutil"
)

var _ = Describe("client", func() {
	var buildClient func() concourse.IClient
	var buildClientOtherRegion func() concourse.IClient
	var actions []string
	var stdout *gbytes.Buffer
	var stderr *gbytes.Buffer
	var deleteBoshDirectorError error
	var args *deploy.Args
	var configInBucket, configAfterLoad config.Config
	var ipChecker func() (string, error)
	var directorStateFixture, directorCredsFixture, exampleDirectorCreds []byte
	var tfInputVarsFactory *concoursefakes.FakeTFInputVarsFactory
	var flyClient *flyfakes.FakeIClient
	var terraformCLI *terraformfakes.FakeCLIInterface
	var configClient *configfakes.FakeIClient
	var boshClient *boshfakes.FakeIClient

	var setupFakeAwsProvider = func() *iaasfakes.FakeProvider {
		provider := &iaasfakes.FakeProvider{}
		provider.DBTypeReturns("db.t2.small")
		provider.RegionReturns("eu-west-1")
		provider.IAASReturns(iaas.AWS)
		provider.CheckForWhitelistedIPStub = func(ip, securityGroup string) (bool, error) {
			actions = append(actions, "checking security group for IP")
			if ip == "1.2.3.4" {
				return false, nil
			}
			return true, nil
		}
		provider.DeleteVMsInVPCStub = func(vpcID string) ([]string, error) {
			actions = append(actions, fmt.Sprintf("deleting vms in %s", vpcID))
			return nil, nil
		}
		provider.FindLongestMatchingHostedZoneStub = func(subdomain string) (string, string, error) {
			if subdomain == "ci.google.com" {
				return "google.com", "ABC123", nil
			}

			return "", "", errors.New("hosted zone not found")
		}
		return provider
	}

	var setupFakeOtherRegionProvider = func() *iaasfakes.FakeProvider {
		otherRegionClient := &iaasfakes.FakeProvider{}
		otherRegionClient.IAASReturns(iaas.AWS)
		otherRegionClient.RegionReturns("eu-central-1")
		return otherRegionClient
	}

	var setupFakeTfInputVarsFactory = func() *concoursefakes.FakeTFInputVarsFactory {
		tfInputVarsFactory = &concoursefakes.FakeTFInputVarsFactory{}

		provider, err := iaas.New(iaas.AWS, "eu-west-1")
		Expect(err).ToNot(HaveOccurred())
		awsInputVarsFactory, err := concourse.NewTFInputVarsFactory(provider)
		Expect(err).ToNot(HaveOccurred())
		tfInputVarsFactory.NewInputVarsStub = func(i config.Config) terraform.InputVars {
			actions = append(actions, "converting config.Config to TFInputVars")
			return awsInputVarsFactory.NewInputVars(i)
		}
		return tfInputVarsFactory
	}

	var setupFakeConfigClient = func() *configfakes.FakeIClient {
		configClient = &configfakes.FakeIClient{}
		configClient.LoadStub = func() (config.Config, error) {
			actions = append(actions, "loading config file")
			return configInBucket, nil
		}
		configClient.DeleteAssetStub = func(filename string) error {
			actions = append(actions, fmt.Sprintf("deleting config asset: %s", filename))
			return nil
		}
		configClient.UpdateStub = func(config config.Config) error {
			actions = append(actions, "updating config file")
			return nil
		}
		configClient.StoreAssetStub = func(filename string, contents []byte) error {
			actions = append(actions, fmt.Sprintf("storing config asset: %s", filename))
			return nil
		}
		configClient.DeleteAllStub = func(config config.Config) error {
			actions = append(actions, "deleting config")
			return nil
		}
		configClient.ConfigExistsStub = func() (bool, error) {
			actions = append(actions, "checking to see if config exists")
			return true, nil
		}
		return configClient
	}

	var setupFakeTerraformCLI = func(terraformOutputs terraform.AWSOutputs) *terraformfakes.FakeCLIInterface {
		terraformCLI = &terraformfakes.FakeCLIInterface{}
		terraformCLI.ApplyStub = func(inputVars terraform.InputVars, dryrun bool) error {
			actions = append(actions, "applying terraform")
			return nil
		}
		terraformCLI.DestroyStub = func(conf terraform.InputVars) error {
			actions = append(actions, "destroying terraform")
			return nil
		}
		terraformCLI.BuildOutputStub = func(conf terraform.InputVars) (terraform.Outputs, error) {
			actions = append(actions, "initializing terraform outputs")
			return &terraformOutputs, nil
		}
		return terraformCLI
	}

	BeforeEach(func() {
		certGenerator := func(c func(u *certs.User) (*lego.Client, error), caName string, provider iaas.Provider, ip ...string) (*certs.Certs, error) {
			actions = append(actions, fmt.Sprintf("generating cert ca: %s, cn: %s", caName, ip))
			return &certs.Certs{
				CACert: []byte("----EXAMPLE CERT----"),
			}, nil
		}

		awsClient := setupFakeAwsProvider()
		otherRegionClient := setupFakeOtherRegionProvider()
		tfInputVarsFactory = setupFakeTfInputVarsFactory()
		configClient = setupFakeConfigClient()

		flyClient = &flyfakes.FakeIClient{}
		flyClient.SetDefaultPipelineStub = func(config config.Config, allowFlyVersionDiscrepancy bool) error {
			actions = append(actions, "setting default pipeline")
			return nil
		}

		args = &deploy.Args{
			AllowIPs:    "0.0.0.0/0",
			DBSize:      "small",
			DBSizeIsSet: false,
		}

		terraformOutputs := terraform.AWSOutputs{
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
		configInBucket = config.Config{
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
			PublicCIDR:        "192.168.0.0/24",
			PrivateCIDR:       "192.168.1.0/24",
		}

		//Mutations we expect to have been done after load
		configAfterLoad = configInBucket
		configAfterLoad.AllowIPs = "\"0.0.0.0/0\""
		configAfterLoad.SourceAccessIP = "192.0.2.0"

		exampleDirectorCreds = []byte("atc_password: s3cret")

		terraformCLI = setupFakeTerraformCLI(terraformOutputs)

		boshClientFactory := func(config config.Config, outputs terraform.Outputs, director director.IClient, stdout, stderr io.Writer, provider iaas.Provider) (bosh.IClient, error) {
			boshClient = &boshfakes.FakeIClient{}
			boshClient.DeployStub = func(stateFileBytes, credsFileBytes []byte, detach bool) ([]byte, []byte, error) {
				if detach {
					actions = append(actions, "deploying director in self-update mode")
				} else {
					actions = append(actions, "deploying director")
				}
				return nil, exampleDirectorCreds, nil
			}
			boshClient.DeleteStub = func([]byte) ([]byte, error) {
				actions = append(actions, "deleting director")
				return nil, deleteBoshDirectorError
			}
			boshClient.CleanupStub = func() error {
				actions = append(actions, "cleaning up bosh init")
				return nil
			}
			boshClient.InstancesStub = func() ([]bosh.Instance, error) {
				actions = append(actions, "listing bosh instances")
				return nil, nil
			}

			return boshClient, nil
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
				tfInputVarsFactory,
				boshClientFactory,
				func(iaas.Provider, fly.Credentials, io.Writer, io.Writer, []byte) (fly.IClient, error) {
					return flyClient, nil
				},
				certGenerator,
				configClient,
				args,
				stdout,
				stderr,
				ipChecker,
				certsfakes.NewFakeAcmeClient,
				"some version",
			)
		}

		buildClientOtherRegion = func() concourse.IClient {
			return concourse.NewClient(
				otherRegionClient,
				terraformCLI,
				tfInputVarsFactory,
				boshClientFactory,
				func(iaas.Provider, fly.Credentials, io.Writer, io.Writer, []byte) (fly.IClient, error) {
					return flyClient, nil
				},
				certGenerator,
				configClient,
				args,
				stdout,
				stderr,
				ipChecker,
				certsfakes.NewFakeAcmeClient,
				"some version",
			)
		}
	})

	Describe("Deploy", func() {
		Context("when there is an existing config and no CLI args", func() {
			BeforeEach(func() {
				var err error
				directorStateFixture, err = ioutil.ReadFile("fixtures/director-state.json")
				Expect(err).ToNot(HaveOccurred())
				directorCredsFixture, err = ioutil.ReadFile("fixtures/director-creds.yml")
				Expect(err).ToNot(HaveOccurred())

				configClient.HasAssetReturnsOnCall(0, true, nil)
				configClient.LoadAssetReturnsOnCall(0, directorStateFixture, nil)
				configClient.HasAssetReturnsOnCall(1, true, nil)
				configClient.LoadAssetReturnsOnCall(1, directorCredsFixture, nil)
			})
			It("does all the things in the right order", func() {
				client := buildClient()
				err := client.Deploy()

				terraformInputVars := &terraform.AWSInputVars{
					NetworkCIDR:            configAfterLoad.NetworkCIDR,
					PublicCIDR:             configAfterLoad.PublicCIDR,
					PrivateCIDR:            configAfterLoad.PrivateCIDR,
					AllowIPs:               configAfterLoad.AllowIPs,
					AvailabilityZone:       configAfterLoad.AvailabilityZone,
					ConfigBucket:           configAfterLoad.ConfigBucket,
					Deployment:             configAfterLoad.Deployment,
					HostedZoneID:           configAfterLoad.HostedZoneID,
					HostedZoneRecordPrefix: configAfterLoad.HostedZoneRecordPrefix,
					Namespace:              configAfterLoad.Namespace,
					Project:                configAfterLoad.Project,
					PublicKey:              configAfterLoad.PublicKey,
					RDSDefaultDatabaseName: configAfterLoad.RDSDefaultDatabaseName,
					RDSInstanceClass:       configAfterLoad.RDSInstanceClass,
					RDSPassword:            configAfterLoad.RDSPassword,
					RDSUsername:            configAfterLoad.RDSUsername,
					Region:                 configAfterLoad.Region,
					SourceAccessIP:         configAfterLoad.SourceAccessIP,
					TFStatePath:            configAfterLoad.TFStatePath,
				}

				tfInputVarsFactory.NewInputVarsReturns(terraformInputVars)

				Expect(err).ToNot(HaveOccurred())

				Expect(actions[0]).To(Equal("checking to see if config exists"))
				Expect(actions[1]).To(Equal("loading config file"))

				Expect(actions[2]).To(Equal("converting config.Config to TFInputVars"))
				Expect(tfInputVarsFactory).To(HaveReceived("NewInputVars").With(configAfterLoad))

				Expect(actions[3]).To(Equal("applying terraform"))
				Expect(terraformCLI).To(HaveReceived("Apply").With(terraformInputVars, false))

				Expect(actions[4]).To(Equal("initializing terraform outputs"))
				Expect(terraformCLI).To(HaveReceived("BuildOutput").With(terraformInputVars))

				Expect(actions[5]).To(Equal("updating config file"))
				Expect(configClient).To(HaveReceived("Update").With(configAfterLoad))
				Expect(actions[6]).To(Equal("generating cert ca: concourse-up-happymeal, cn: [99.99.99.99 192.168.0.6]"))
				Expect(actions[7]).To(Equal("generating cert ca: concourse-up-happymeal, cn: [77.77.77.77]"))

				Expect(actions[8]).To(Equal("deploying director"))
				Expect(configClient).To(HaveReceived("HasAsset").With("director-state.json"))
				Expect(configClient.HasAssetArgsForCall(0)).To(Equal("director-state.json"))
				Expect(configClient).To(HaveReceived("LoadAsset").With("director-state.json"))
				Expect(configClient.LoadAssetArgsForCall(0)).To(Equal("director-state.json"))

				Expect(configClient).To(HaveReceived("HasAsset").With("director-creds.yml"))
				Expect(configClient.HasAssetArgsForCall(1)).To(Equal("director-creds.yml"))
				Expect(configClient).To(HaveReceived("LoadAsset").With("director-creds.yml"))
				Expect(configClient.LoadAssetArgsForCall(1)).To(Equal("director-creds.yml"))

				Expect(boshClient).To(HaveReceived("Deploy").With(directorStateFixture, directorCredsFixture, false))

				Expect(actions[9]).To(Equal("storing config asset: director-state.json"))
				Expect(actions[10]).To(Equal("storing config asset: director-creds.yml"))
				Expect(actions[11]).To(Equal("cleaning up bosh init"))
				Expect(actions[12]).To(Equal("setting default pipeline"))
				Expect(actions[13]).To(Equal("updating config file"))
				Expect(len(actions)).To(Equal(14))
			})
		})

		It("Prints a warning about changing the sourceIP", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr).To(gbytes.Say("WARNING: allowing access from local machine"))
		})

		Context("When a custom domain is required", func() {
			It("Prints a warning about adding a DNS record", func() {
				configInBucket.Domain = "ci.google.com"

				client := buildClient()
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				Expect(stderr).To(gbytes.Say("WARNING: adding record ci.google.com to DNS zone google.com with name ABC123"))
			})
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
				configInBucket.Domain = "ci.google.com"

				client := buildClient()
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				Expect(actions).To(ContainElement("generating cert ca: concourse-up-happymeal, cn: [ci.google.com]"))
			})
		})

		Context("When running in self-update mode and the concourse is already deployed", func() {
			It("Sets the default pipeline, before deploying the bosh director", func() {
				flyClient.CanConnectStub = func() (bool, error) {
					return true, nil
				}
				args.SelfUpdate = true

				client := buildClient()
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				Expect(actions).To(ContainElement("deploying director in self-update mode"))
			})
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
				configInBucket.Domain = "ci.google.com"
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
			Expect(tfInputVarsFactory).To(HaveReceived("NewInputVars").With(configInBucket))
		})
		It("Loads terraform output", func() {
			client := buildClient()
			err := client.Destroy()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("initializing terraform outputs"))
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
		BeforeEach(func() {
			directorCredsFixture, err := ioutil.ReadFile("fixtures/director-creds.yml")
			Expect(err).ToNot(HaveOccurred())

			configClient.HasAssetReturnsOnCall(0, true, nil)
			configClient.LoadAssetReturnsOnCall(0, directorCredsFixture, nil)
		})
		It("Loads the config file", func() {
			client := buildClient()
			_, err := client.FetchInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("loading config file"))
		})
		It("calls TFInputVarsFactory, having populated AllowIPs and SourceAccessIPs", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())
			Expect(tfInputVarsFactory).To(HaveReceived("NewInputVars").With(configAfterLoad))
		})

		It("Loads terraform output", func() {
			client := buildClient()
			_, err := client.FetchInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(actions).To(ContainElement("initializing terraform outputs"))
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
				Expect(err).To(MatchError("Do you need to add your IP 1.2.3.4 to the concourse-up-happymeal-director security group/source range entry for director firewall (for ports 22, 6868, and 25555)?"))
			})
		})
	})
})
