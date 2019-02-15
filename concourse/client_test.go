package concourse_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/bosh/boshfakes"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/certs/certsfakes"
	"github.com/EngineerBetter/concourse-up/commands/deploy"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/concourse/concoursefakes"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/config/configfakes"
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
)

var _ = Describe("client", func() {
	var buildClient func() concourse.IClient
	var actions []string
	var stdout *gbytes.Buffer
	var stderr *gbytes.Buffer
	var deleteBoshDirectorError error
	var args *deploy.Args
	var configInBucket, configAfterLoad, configAfterCreateEnv config.Config
	var ipChecker func() (string, error)
	var directorStateFixture, directorCredsFixture []byte
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
		var err error
		directorStateFixture, err = ioutil.ReadFile("fixtures/director-state.json")
		Expect(err).ToNot(HaveOccurred())
		directorCredsFixture, err = ioutil.ReadFile("fixtures/director-creds.yml")
		Expect(err).ToNot(HaveOccurred())

		certGenerator := func(c func(u *certs.User) (*lego.Client, error), caName string, provider iaas.Provider, ip ...string) (*certs.Certs, error) {
			actions = append(actions, fmt.Sprintf("generating cert ca: %s, cn: %s", caName, ip))
			return &certs.Certs{
				CACert: []byte("----EXAMPLE CERT----"),
			}, nil
		}

		awsClient := setupFakeAwsProvider()
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
		configAfterLoad.NetworkCIDR = "10.0.0.0/16"
		configAfterLoad.PublicCIDR = "10.0.0.0/24"
		configAfterLoad.PrivateCIDR = "10.0.1.0/24"
		configAfterLoad.RDS1CIDR = "10.0.4.0/24"
		configAfterLoad.RDS2CIDR = "10.0.5.0/24"

		//Mutations we expect to have been done after Deploy
		configAfterCreateEnv = configAfterLoad
		configAfterCreateEnv.ConcourseCACert = "----EXAMPLE CERT----"
		configAfterCreateEnv.DirectorCACert = "----EXAMPLE CERT----"
		configAfterCreateEnv.DirectorPublicIP = "99.99.99.99"
		configAfterCreateEnv.Domain = "77.77.77.77"
		configAfterCreateEnv.Tags = []string{"concourse-up-version=some version"}
		configAfterCreateEnv.Version = "some version"

		terraformCLI = setupFakeTerraformCLI(terraformOutputs)

		boshClientFactory := func(config config.Config, outputs terraform.Outputs, stdout, stderr io.Writer, provider iaas.Provider, versionFile []byte) (bosh.IClient, error) {
			boshClient = &boshfakes.FakeIClient{}
			boshClient.DeployStub = func(stateFileBytes, credsFileBytes []byte, detach bool) ([]byte, []byte, error) {
				if detach {
					actions = append(actions, "deploying director in self-update mode")
				} else {
					actions = append(actions, "deploying director")
				}
				return directorStateFixture, directorCredsFixture, nil
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
				func(size int) string { return fmt.Sprintf("generatedPassword%d", size) },
				func() string { return "8letters" },
				func() ([]byte, []byte, string, error) { return []byte("private"), []byte("public"), "fingerprint", nil },
				"some version",
			)
		}
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
