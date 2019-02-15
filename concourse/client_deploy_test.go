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
	var certGenerationActions []string
	var stdout *gbytes.Buffer
	var stderr *gbytes.Buffer
	var deleteBoshDirectorError error
	var args *deploy.Args
	var configInBucket config.Config
	var terraformOutputs terraform.AWSOutputs

	var directorStateFixture, directorCredsFixture []byte

	var buildClient func() concourse.IClient
	var buildClientOtherRegion func() concourse.IClient
	var ipChecker func() (string, error)
	var tfInputVarsFactory *concoursefakes.FakeTFInputVarsFactory
	var flyClient *flyfakes.FakeIClient
	var terraformCLI *terraformfakes.FakeCLIInterface
	var configClient *configfakes.FakeIClient
	var boshClient *boshfakes.FakeIClient

	var setupFakeAwsProvider = func() *iaasfakes.FakeProvider {
		provider := &iaasfakes.FakeProvider{}
		provider.DBTypeStub = func(size string) string {
			return "db.t2." + size
		}
		provider.RegionReturns("eu-west-1")
		provider.IAASReturns(iaas.AWS)
		provider.CheckForWhitelistedIPStub = func(ip, securityGroup string) (bool, error) {
			if ip == "1.2.3.4" {
				return false, nil
			}
			return true, nil
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
			return awsInputVarsFactory.NewInputVars(i)
		}
		return tfInputVarsFactory
	}

	var setupFakeTerraformCLI = func(terraformOutputs terraform.AWSOutputs) *terraformfakes.FakeCLIInterface {
		terraformCLI = &terraformfakes.FakeCLIInterface{}
		terraformCLI.BuildOutputReturns(&terraformOutputs, nil)
		return terraformCLI
	}

	BeforeEach(func() {
		var err error
		directorStateFixture, err = ioutil.ReadFile("fixtures/director-state.json")
		Expect(err).ToNot(HaveOccurred())
		directorCredsFixture, err = ioutil.ReadFile("fixtures/director-creds.yml")
		Expect(err).ToNot(HaveOccurred())

		//At the time of writing, these are defaults from the CLI flags
		args = &deploy.Args{
			AllowIPs:         "0.0.0.0/0",
			AllowIPsIsSet:    false,
			DBSize:           "small",
			DBSizeIsSet:      false,
			IAAS:             "AWS",
			IAASIsSet:        false,
			Preemptible:      true,
			Spot:             true,
			SpotIsSet:        false,
			WebSize:          "small",
			WebSizeIsSet:     false,
			WorkerCount:      1,
			WorkerCountIsSet: false,
			WorkerSize:       "xlarge",
			WorkerSizeIsSet:  false,
			WorkerType:       "m4",
			WorkerTypeIsSet:  false,
		}

		terraformOutputs = terraform.AWSOutputs{
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
		certGenerationActions = []string{}

		// Initial config in bucket from an existing deployment
		configInBucket = config.Config{
			ConcoursePassword: "s3cret",
			ConcourseUsername: "admin",
			Deployment:        "concourse-up-happymeal",
			DirectorPassword:  "secret123",
			DirectorUsername:  "admin",
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
			Project:          "happymeal",
			PublicKey:        "example-public-key",
			RDSInstanceClass: "db.t2.medium",
			RDSPassword:      "s3cret",
			RDSUsername:      "admin",
			Region:           "eu-west-1",
			TFStatePath:      "example-path",
			//These come from fixtures/director-creds.yml
			CredhubUsername:          "credhub-cli",
			CredhubPassword:          "f4b12bc0166cad1bc02b050e4e79ac4c",
			CredhubAdminClientSecret: "hxfgb56zny2yys6m9wjx",
			CredhubCACert:            "-----BEGIN CERTIFICATE-----\nMIIEXTCCAsWgAwIBAgIQSmhcetyHDHLOYGaqMnJ0QTANBgkqhkiG9w0BAQsFADA4\nMQwwCgYDVQQGEwNVU0ExFjAUBgNVBAoTDUNsb3VkIEZvdW5kcnkxEDAOBgNVBAMM\nB2Jvc2hfY2EwHhcNMTkwMjEzMTAyNTM0WhcNMjAwMjEzMTAyNTM0WjA4MQwwCgYD\nVQQGEwNVU0ExFjAUBgNVBAoTDUNsb3VkIEZvdW5kcnkxEDAOBgNVBAMMB2Jvc2hf\nY2EwggGiMA0GCSqGSIb3DQEBAQUAA4IBjwAwggGKAoIBgQC+0bA9T4awlJYSn6aq\nun6Hylu47b2UiZpFZpvPomKWPay86QaJ0vC9SK8keoYI4gWwsZSAMXp2mSCkXKRi\n+rVc+sKnzv9VgPoVY5eYIYCtJvl7KCJQE02dGoxuGOaWlBiHuD6TzY6lI9fNxkAW\neMGR3UylJ7ET0NvgAZWS1daov2GfiKkaYUCdbY8DtfhMyFhJ381VNHwoP6xlZbSf\nTInO/2TS8xpW2BcMNhFAu9MJVtC5pDHtJtkXHXep027CkrPjtFQWpzvIMvPAtZ68\n9t46nS9Ix+RmeN3v+sawNzbZscnsslhB+m4GrpL9M8g8sbweMw9yxf241z1qkiNJ\nto3HRqqyNyGsvI9n7OUrZ4D5oAfY7ze1TF+nxnkmJp14y21FEdG7t76N0J5dn6bJ\n/lroojig/PqabRsyHbmj6g8N832PEQvwsPptihEwgrRmY6fcBbMUaPCpNuVTJVa5\ng0KdBGDYDKTMlEn4xaj8P1wRbVjtXVMED2l4K4tS/UiDIb8CAwEAAaNjMGEwDgYD\nVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFHii4fiqAwJS\nnNhi6C+ibr/4OOTyMB8GA1UdIwQYMBaAFHii4fiqAwJSnNhi6C+ibr/4OOTyMA0G\nCSqGSIb3DQEBCwUAA4IBgQAGXDTlsQWIJHfvU3zy9te35adKOUeDwk1lSe4NYvgW\nFJC0w2K/1ZldmQ2leHmiXSukDJAYmROy9Y1qkUazTzjsdvHGhUF2N1p7fIweNj8e\ncsR+T21MjPEwD99m5+xLvnMRMuqzH9TqVbFIM3lmCDajh8n9cp4KvGkQmB+X7DE1\nR6AXG4EN9xn91TFrqmFFNOrFtoAjtag05q/HoqMhFFVeg+JTpsPshFjlWIkzwqKx\npn68KG2ztgS0KeDraGKwItTKengTCr/VkgorXnhKcI1C6C5iRXZp3wREu8RO+wRe\nKSGbsYIHaFxd3XwW4JnsW+hes/W5MZX01wkwOLrktf85FjssBZBavxBbyFag/LvS\n8oULOZRLYUkuElM+0Wzf8ayB574Fd97gzCVzWoD0Ei982jAdbEfk77PV1TvMNmEn\n3M6ktB7GkjuD9OL12iNzxmbQe7p1WkYYps9hK4r0pbyxZPZlPMmNNZo579rywDjF\nwEW5QkylaPEkbVDhJWeR1I8=\n-----END CERTIFICATE-----\n",
		}
	})

	JustBeforeEach(func() {
		certGenerator := func(c func(u *certs.User) (*lego.Client, error), caName string, provider iaas.Provider, ip ...string) (*certs.Certs, error) {
			certGenerationActions = append(certGenerationActions, fmt.Sprintf("generating cert ca: %s, cn: %s", caName, ip))
			return &certs.Certs{
				CACert: []byte("----EXAMPLE CERT----"),
			}, nil
		}

		flyClient = &flyfakes.FakeIClient{}
		awsClient := setupFakeAwsProvider()
		otherRegionClient := setupFakeOtherRegionProvider()
		tfInputVarsFactory = setupFakeTfInputVarsFactory()
		configClient = &configfakes.FakeIClient{}
		terraformCLI = setupFakeTerraformCLI(terraformOutputs)

		boshClientFactory := func(config config.Config, outputs terraform.Outputs, stdout, stderr io.Writer, provider iaas.Provider, versionFile []byte) (bosh.IClient, error) {
			boshClient = &boshfakes.FakeIClient{}
			boshClient.DeployReturns(directorStateFixture, directorCredsFixture, nil)
			boshClient.DeleteReturns(nil, deleteBoshDirectorError)
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
				func(size int) string { return fmt.Sprintf("generatedPassword%d", size) },
				func() string { return "8letters" },
				func() ([]byte, []byte, string, error) { return []byte("private"), []byte("public"), "fingerprint", nil },
				"some version",
			)
		}
	})

	Describe("Deploy", func() {
		Context("when there is an existing config", func() {
			var configAfterLoad, configAfterCreateEnv, configAfterConcourseDeploy config.Config
			var terraformInputVars *terraform.AWSInputVars

			Context("and no CLI args were provided", func() {
				BeforeEach(func() {
					//Mutations we expect to have been done after load
					configAfterLoad = configInBucket
					configAfterLoad.AllowIPs = "\"0.0.0.0/0\""
					configAfterLoad.SourceAccessIP = "192.0.2.0"
					configAfterLoad.NetworkCIDR = "10.0.0.0/16"
					configAfterLoad.PublicCIDR = "10.0.0.0/24"
					configAfterLoad.PrivateCIDR = "10.0.1.0/24"
					configAfterLoad.RDS1CIDR = "10.0.4.0/24"
					configAfterLoad.RDS2CIDR = "10.0.5.0/24"

					terraformInputVars = &terraform.AWSInputVars{
						AllowIPs:               configAfterLoad.AllowIPs,
						AvailabilityZone:       configAfterLoad.AvailabilityZone,
						ConfigBucket:           configAfterLoad.ConfigBucket,
						Deployment:             configAfterLoad.Deployment,
						HostedZoneID:           configAfterLoad.HostedZoneID,
						HostedZoneRecordPrefix: configAfterLoad.HostedZoneRecordPrefix,
						Namespace:              configAfterLoad.Namespace,
						NetworkCIDR:            configAfterLoad.NetworkCIDR,
						PrivateCIDR:            configAfterLoad.PrivateCIDR,
						Project:                configAfterLoad.Project,
						PublicCIDR:             configAfterLoad.PublicCIDR,
						PublicKey:              configAfterLoad.PublicKey,
						RDS1CIDR:               configAfterLoad.RDS1CIDR,
						RDS2CIDR:               configAfterLoad.RDS2CIDR,
						RDSDefaultDatabaseName: configAfterLoad.RDSDefaultDatabaseName,
						RDSInstanceClass:       configAfterLoad.RDSInstanceClass,
						RDSPassword:            configAfterLoad.RDSPassword,
						RDSUsername:            configAfterLoad.RDSUsername,
						Region:                 configAfterLoad.Region,
						SourceAccessIP:         configAfterLoad.SourceAccessIP,
						TFStatePath:            configAfterLoad.TFStatePath,
					}

					//Mutations we expect to have been done after deploying the director
					configAfterCreateEnv = configAfterLoad
					configAfterCreateEnv.ConcourseCACert = "----EXAMPLE CERT----"
					configAfterCreateEnv.DirectorCACert = "----EXAMPLE CERT----"
					configAfterCreateEnv.DirectorPublicIP = "99.99.99.99"
					configAfterCreateEnv.Domain = "77.77.77.77"
					configAfterCreateEnv.Tags = []string{"concourse-up-version=some version"}
					configAfterCreateEnv.Version = "some version"

					// Mutations we expect to have been done after deploying Concourse
					configAfterConcourseDeploy = configAfterCreateEnv
					configAfterConcourseDeploy.CredhubAdminClientSecret = "hxfgb56zny2yys6m9wjx"
					configAfterConcourseDeploy.CredhubCACert = `-----BEGIN CERTIFICATE-----
MIIEXTCCAsWgAwIBAgIQSmhcetyHDHLOYGaqMnJ0QTANBgkqhkiG9w0BAQsFADA4
MQwwCgYDVQQGEwNVU0ExFjAUBgNVBAoTDUNsb3VkIEZvdW5kcnkxEDAOBgNVBAMM
B2Jvc2hfY2EwHhcNMTkwMjEzMTAyNTM0WhcNMjAwMjEzMTAyNTM0WjA4MQwwCgYD
VQQGEwNVU0ExFjAUBgNVBAoTDUNsb3VkIEZvdW5kcnkxEDAOBgNVBAMMB2Jvc2hf
Y2EwggGiMA0GCSqGSIb3DQEBAQUAA4IBjwAwggGKAoIBgQC+0bA9T4awlJYSn6aq
un6Hylu47b2UiZpFZpvPomKWPay86QaJ0vC9SK8keoYI4gWwsZSAMXp2mSCkXKRi
+rVc+sKnzv9VgPoVY5eYIYCtJvl7KCJQE02dGoxuGOaWlBiHuD6TzY6lI9fNxkAW
eMGR3UylJ7ET0NvgAZWS1daov2GfiKkaYUCdbY8DtfhMyFhJ381VNHwoP6xlZbSf
TInO/2TS8xpW2BcMNhFAu9MJVtC5pDHtJtkXHXep027CkrPjtFQWpzvIMvPAtZ68
9t46nS9Ix+RmeN3v+sawNzbZscnsslhB+m4GrpL9M8g8sbweMw9yxf241z1qkiNJ
to3HRqqyNyGsvI9n7OUrZ4D5oAfY7ze1TF+nxnkmJp14y21FEdG7t76N0J5dn6bJ
/lroojig/PqabRsyHbmj6g8N832PEQvwsPptihEwgrRmY6fcBbMUaPCpNuVTJVa5
g0KdBGDYDKTMlEn4xaj8P1wRbVjtXVMED2l4K4tS/UiDIb8CAwEAAaNjMGEwDgYD
VR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFHii4fiqAwJS
nNhi6C+ibr/4OOTyMB8GA1UdIwQYMBaAFHii4fiqAwJSnNhi6C+ibr/4OOTyMA0G
CSqGSIb3DQEBCwUAA4IBgQAGXDTlsQWIJHfvU3zy9te35adKOUeDwk1lSe4NYvgW
FJC0w2K/1ZldmQ2leHmiXSukDJAYmROy9Y1qkUazTzjsdvHGhUF2N1p7fIweNj8e
csR+T21MjPEwD99m5+xLvnMRMuqzH9TqVbFIM3lmCDajh8n9cp4KvGkQmB+X7DE1
R6AXG4EN9xn91TFrqmFFNOrFtoAjtag05q/HoqMhFFVeg+JTpsPshFjlWIkzwqKx
pn68KG2ztgS0KeDraGKwItTKengTCr/VkgorXnhKcI1C6C5iRXZp3wREu8RO+wRe
KSGbsYIHaFxd3XwW4JnsW+hes/W5MZX01wkwOLrktf85FjssBZBavxBbyFag/LvS
8oULOZRLYUkuElM+0Wzf8ayB574Fd97gzCVzWoD0Ei982jAdbEfk77PV1TvMNmEn
3M6ktB7GkjuD9OL12iNzxmbQe7p1WkYYps9hK4r0pbyxZPZlPMmNNZo579rywDjF
wEW5QkylaPEkbVDhJWeR1I8=
-----END CERTIFICATE-----
`
					configAfterConcourseDeploy.CredhubPassword = "f4b12bc0166cad1bc02b050e4e79ac4c"
					configAfterConcourseDeploy.CredhubURL = "https://77.77.77.77:8844/"
					configAfterConcourseDeploy.CredhubUsername = "credhub-cli"
				})

				JustBeforeEach(func() {
					configClient.LoadReturns(configInBucket, nil)
					configClient.ConfigExistsReturns(true, nil)
					configClient.HasAssetReturnsOnCall(0, true, nil)
					configClient.LoadAssetReturnsOnCall(0, directorStateFixture, nil)
					configClient.HasAssetReturnsOnCall(1, true, nil)
					configClient.LoadAssetReturnsOnCall(1, directorCredsFixture, nil)
				})

				It("does all the things in the right order", func() {
					client := buildClient()
					err := client.Deploy()
					Expect(err).ToNot(HaveOccurred())

					tfInputVarsFactory.NewInputVarsReturns(terraformInputVars)

					Expect(configClient).To(HaveReceived("ConfigExists"))
					Expect(configClient).To(HaveReceived("Load"))
					Expect(tfInputVarsFactory).To(HaveReceived("NewInputVars").With(configAfterLoad))
					Expect(terraformCLI).To(HaveReceived("Apply").With(terraformInputVars, false))
					Expect(terraformCLI).To(HaveReceived("BuildOutput").With(terraformInputVars))
					Expect(configClient).To(HaveReceived("Update").With(configAfterLoad))

					Expect(certGenerationActions[0]).To(Equal("generating cert ca: concourse-up-happymeal, cn: [99.99.99.99 10.0.0.6]"))
					Expect(certGenerationActions[1]).To(Equal("generating cert ca: concourse-up-happymeal, cn: [77.77.77.77]"))

					Expect(configClient).To(HaveReceived("HasAsset").With("director-state.json"))
					Expect(configClient.HasAssetArgsForCall(0)).To(Equal("director-state.json"))
					Expect(configClient).To(HaveReceived("LoadAsset").With("director-state.json"))
					Expect(configClient.LoadAssetArgsForCall(0)).To(Equal("director-state.json"))
					Expect(configClient).To(HaveReceived("HasAsset").With("director-creds.yml"))
					Expect(configClient.HasAssetArgsForCall(1)).To(Equal("director-creds.yml"))
					Expect(configClient).To(HaveReceived("LoadAsset").With("director-creds.yml"))
					Expect(configClient.LoadAssetArgsForCall(1)).To(Equal("director-creds.yml"))
					Expect(boshClient).To(HaveReceived("Deploy").With(directorStateFixture, directorCredsFixture, false))

					Expect(configClient).To(HaveReceived("StoreAsset").With("director-state.json", directorStateFixture))
					Expect(configClient).To(HaveReceived("StoreAsset").With("director-creds.yml", directorCredsFixture))
					Expect(boshClient).To(HaveReceived("Cleanup"))
					Expect(flyClient).To(HaveReceived("SetDefaultPipeline").With(configAfterCreateEnv, false))
					Expect(configClient).To(HaveReceived("Update").With(configAfterConcourseDeploy))
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

				It("Notifies the user", func() {
					client := buildClient()
					err := client.Deploy()
					Expect(err).ToNot(HaveOccurred())

					Eventually(stdout).Should(gbytes.Say("USING PREVIOUS DEPLOYMENT CONFIG"))
				})
			})

			Context("and all the CLI args were provided", func() {
				BeforeEach(func() {
					// Set all changeable arguments (IE, not IAAS, Region, Namespace, AZ, et al)
					args.AllowIPs = "88.98.225.40"
					args.AllowIPsIsSet = true
					args.DBSize = "4xlarge"
					args.DBSizeIsSet = true
					args.Domain = "ci.google.com"
					args.DomainIsSet = true
					args.GithubAuthClientID = "github-client-id"
					args.GithubAuthClientIDIsSet = true
					args.GithubAuthClientSecret = "github-client-secret"
					args.GithubAuthClientSecretIsSet = true
					args.GithubAuthIsSet = true
					args.Preemptible = false
					args.Spot = false
					args.SpotIsSet = true
					args.Tags = []string{"env=prod", "team=foo"}
					args.TagsIsSet = true
					args.TLSCert = "i-am-a-tls-cert"
					args.TLSCertIsSet = true
					args.TLSKey = "i-am-a-tls-key"
					args.TLSKeyIsSet = true
					args.WebSize = "2xlarge"
					args.WebSizeIsSet = true
					args.WorkerCount = 2
					args.WorkerCountIsSet = true
					args.WorkerSize = "4xlarge"
					args.WorkerSizeIsSet = true
					args.WorkerType = "m5"
					args.WorkerTypeIsSet = true

					configAfterLoad = configInBucket
					configAfterLoad.AllowIPs = "\"88.98.225.40/32\""
					configAfterLoad.ConcourseWebSize = args.WebSize
					configAfterLoad.ConcourseWorkerCount = args.WorkerCount
					configAfterLoad.ConcourseWorkerSize = args.WorkerSize
					configAfterLoad.Domain = args.Domain
					configAfterLoad.GithubAuthIsSet = true
					configAfterLoad.GithubClientID = args.GithubAuthClientID
					configAfterLoad.GithubClientSecret = args.GithubAuthClientSecret
					configAfterLoad.HostedZoneID = "ABC123"
					configAfterLoad.HostedZoneRecordPrefix = "ci"
					configAfterLoad.NetworkCIDR = "10.0.0.0/16"
					configAfterLoad.PrivateCIDR = "10.0.1.0/24"
					configAfterLoad.PublicCIDR = "10.0.0.0/24"
					configAfterLoad.RDS1CIDR = "10.0.4.0/24"
					configAfterLoad.RDS2CIDR = "10.0.5.0/24"
					configAfterLoad.RDSInstanceClass = "db.t2.4xlarge"
					configAfterLoad.SourceAccessIP = "192.0.2.0"
					configAfterLoad.Spot = false
					configAfterLoad.Tags = args.Tags
					configAfterLoad.WorkerType = args.WorkerType

					terraformInputVars = &terraform.AWSInputVars{
						AllowIPs:               configAfterLoad.AllowIPs,
						AvailabilityZone:       configAfterLoad.AvailabilityZone,
						ConfigBucket:           configAfterLoad.ConfigBucket,
						Deployment:             configAfterLoad.Deployment,
						HostedZoneID:           configAfterLoad.HostedZoneID,
						HostedZoneRecordPrefix: configAfterLoad.HostedZoneRecordPrefix,
						Namespace:              configAfterLoad.Namespace,
						NetworkCIDR:            configAfterLoad.NetworkCIDR,
						PrivateCIDR:            configAfterLoad.PrivateCIDR,
						Project:                configAfterLoad.Project,
						PublicCIDR:             configAfterLoad.PublicCIDR,
						PublicKey:              configAfterLoad.PublicKey,
						RDS1CIDR:               configAfterLoad.RDS1CIDR,
						RDS2CIDR:               configAfterLoad.RDS2CIDR,
						RDSDefaultDatabaseName: configAfterLoad.RDSDefaultDatabaseName,
						RDSInstanceClass:       configAfterLoad.RDSInstanceClass,
						RDSPassword:            configAfterLoad.RDSPassword,
						RDSUsername:            configAfterLoad.RDSUsername,
						Region:                 configAfterLoad.Region,
						SourceAccessIP:         configAfterLoad.SourceAccessIP,
						TFStatePath:            configAfterLoad.TFStatePath,
					}

					configAfterCreateEnv = configAfterLoad
					configAfterCreateEnv.ConcourseCert = args.TLSCert
					configAfterCreateEnv.ConcourseKey = args.TLSKey
					configAfterCreateEnv.ConcourseUserProvidedCert = true
					configAfterCreateEnv.DirectorCACert = "----EXAMPLE CERT----"
					configAfterCreateEnv.DirectorPublicIP = "99.99.99.99"
					configAfterCreateEnv.Tags = append([]string{"concourse-up-version=some version"}, args.Tags...)
					configAfterCreateEnv.Version = "some version"

					configAfterConcourseDeploy = configAfterCreateEnv
					configAfterConcourseDeploy.CredhubURL = "https://ci.google.com:8844/"
				})

				JustBeforeEach(func() {
					configClient.LoadReturns(configInBucket, nil)
					configClient.ConfigExistsReturns(true, nil)
					configClient.HasAssetReturnsOnCall(0, true, nil)
					configClient.LoadAssetReturnsOnCall(0, directorStateFixture, nil)
					configClient.HasAssetReturnsOnCall(1, true, nil)
					configClient.LoadAssetReturnsOnCall(1, directorCredsFixture, nil)
				})

				It("updates config and calls collaborators with the current arguments", func() {
					client := buildClient()
					err := client.Deploy()
					Expect(err).ToNot(HaveOccurred())

					Expect(configClient).To(HaveReceived("ConfigExists"))
					Expect(configClient).To(HaveReceived("Load"))
					Expect(tfInputVarsFactory).To(HaveReceived("NewInputVars").With(configAfterLoad))

					Expect(terraformCLI).To(HaveReceived("Apply").With(terraformInputVars, false))
					Expect(terraformCLI).To(HaveReceived("BuildOutput").With(terraformInputVars))
					Expect(configClient).To(HaveReceived("Update").With(configAfterLoad))

					Expect(configClient).To(HaveReceived("HasAsset").With("director-state.json"))
					Expect(configClient.HasAssetArgsForCall(0)).To(Equal("director-state.json"))
					Expect(configClient).To(HaveReceived("LoadAsset").With("director-state.json"))
					Expect(configClient.LoadAssetArgsForCall(0)).To(Equal("director-state.json"))
					Expect(configClient).To(HaveReceived("HasAsset").With("director-creds.yml"))
					Expect(configClient.HasAssetArgsForCall(1)).To(Equal("director-creds.yml"))
					Expect(configClient).To(HaveReceived("LoadAsset").With("director-creds.yml"))
					Expect(configClient.LoadAssetArgsForCall(1)).To(Equal("director-creds.yml"))
					Expect(boshClient).To(HaveReceived("Deploy").With(directorStateFixture, directorCredsFixture, false))

					Expect(configClient).To(HaveReceived("StoreAsset").With("director-state.json", directorStateFixture))
					Expect(configClient).To(HaveReceived("StoreAsset").With("director-creds.yml", directorCredsFixture))
					Expect(boshClient).To(HaveReceived("Cleanup"))
					Expect(flyClient).To(HaveReceived("SetDefaultPipeline").With(configAfterCreateEnv, false))
					Expect(configClient).To(HaveReceived("Update").With(configAfterConcourseDeploy))
				})
			})
		})

		Context("a new deployment with no CLI args", func() {
			var defaultGeneratedConfig, configAfterLoad, configAfterCreateEnv, configAfterConcourseDeploy config.Config
			BeforeEach(func() {
				// Config generated by default for a new deployment
				defaultGeneratedConfig = config.Config{
					AllowIPs:                 "\"0.0.0.0/0\"",
					AvailabilityZone:         "",
					ConcoursePassword:        "",
					ConcourseUsername:        "",
					ConcourseWebSize:         "small",
					ConcourseWorkerCount:     1,
					ConcourseWorkerSize:      "xlarge",
					ConfigBucket:             "concourse-up-initial-deployment-eu-west-1-config",
					DirectorHMUserPassword:   "generatedPassword20",
					DirectorMbusPassword:     "generatedPassword20",
					DirectorNATSPassword:     "generatedPassword20",
					Deployment:               "concourse-up-initial-deployment",
					DirectorPassword:         "generatedPassword20",
					DirectorRegistryPassword: "generatedPassword20",
					DirectorUsername:         "admin",
					EncryptionKey:            "generatedPassword32",
					IAAS:                     "AWS",
					NetworkCIDR:              "10.0.0.0/16",
					PrivateCIDR:              "10.0.1.0/24",
					PrivateKey:               "private",
					Project:                  "initial-deployment",
					PublicCIDR:               "10.0.0.0/24",
					PublicKey:                "public",
					RDS1CIDR:                 "10.0.4.0/24",
					RDS2CIDR:                 "10.0.5.0/24",
					RDSDefaultDatabaseName:   "bosh_8letters",
					RDSInstanceClass:         "db.t2.small",
					RDSPassword:              "generatedPassword20",
					RDSUsername:              "admingeneratedPassword7",
					Region:                   "eu-west-1",
					SourceAccessIP:           "192.0.2.0",
					Spot:                     true,
					TFStatePath:              "terraform.tfstate",
					WorkerType:               "m4",
				}

				//Mutations we expect to have been done after load
				configAfterLoad = defaultGeneratedConfig
				configAfterLoad.AllowIPs = "\"0.0.0.0/0\""
				configAfterLoad.SourceAccessIP = "192.0.2.0"

				//Mutations we expect to have been done after deploying the director
				configAfterCreateEnv = configAfterLoad
				configAfterCreateEnv.ConcourseCACert = "----EXAMPLE CERT----"
				configAfterCreateEnv.DirectorCACert = "----EXAMPLE CERT----"
				configAfterCreateEnv.DirectorPublicIP = "99.99.99.99"
				configAfterCreateEnv.Domain = "77.77.77.77"
				configAfterCreateEnv.Tags = []string{"concourse-up-version=some version"}
				configAfterCreateEnv.Version = "some version"

				// Mutations we expect to have been done after deploying Concourse
				configAfterConcourseDeploy = configAfterCreateEnv
				configAfterConcourseDeploy.ConcourseUsername = "admin"
				configAfterConcourseDeploy.CredhubAdminClientSecret = "hxfgb56zny2yys6m9wjx"
				configAfterConcourseDeploy.CredhubCACert = `-----BEGIN CERTIFICATE-----
MIIEXTCCAsWgAwIBAgIQSmhcetyHDHLOYGaqMnJ0QTANBgkqhkiG9w0BAQsFADA4
MQwwCgYDVQQGEwNVU0ExFjAUBgNVBAoTDUNsb3VkIEZvdW5kcnkxEDAOBgNVBAMM
B2Jvc2hfY2EwHhcNMTkwMjEzMTAyNTM0WhcNMjAwMjEzMTAyNTM0WjA4MQwwCgYD
VQQGEwNVU0ExFjAUBgNVBAoTDUNsb3VkIEZvdW5kcnkxEDAOBgNVBAMMB2Jvc2hf
Y2EwggGiMA0GCSqGSIb3DQEBAQUAA4IBjwAwggGKAoIBgQC+0bA9T4awlJYSn6aq
un6Hylu47b2UiZpFZpvPomKWPay86QaJ0vC9SK8keoYI4gWwsZSAMXp2mSCkXKRi
+rVc+sKnzv9VgPoVY5eYIYCtJvl7KCJQE02dGoxuGOaWlBiHuD6TzY6lI9fNxkAW
eMGR3UylJ7ET0NvgAZWS1daov2GfiKkaYUCdbY8DtfhMyFhJ381VNHwoP6xlZbSf
TInO/2TS8xpW2BcMNhFAu9MJVtC5pDHtJtkXHXep027CkrPjtFQWpzvIMvPAtZ68
9t46nS9Ix+RmeN3v+sawNzbZscnsslhB+m4GrpL9M8g8sbweMw9yxf241z1qkiNJ
to3HRqqyNyGsvI9n7OUrZ4D5oAfY7ze1TF+nxnkmJp14y21FEdG7t76N0J5dn6bJ
/lroojig/PqabRsyHbmj6g8N832PEQvwsPptihEwgrRmY6fcBbMUaPCpNuVTJVa5
g0KdBGDYDKTMlEn4xaj8P1wRbVjtXVMED2l4K4tS/UiDIb8CAwEAAaNjMGEwDgYD
VR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFHii4fiqAwJS
nNhi6C+ibr/4OOTyMB8GA1UdIwQYMBaAFHii4fiqAwJSnNhi6C+ibr/4OOTyMA0G
CSqGSIb3DQEBCwUAA4IBgQAGXDTlsQWIJHfvU3zy9te35adKOUeDwk1lSe4NYvgW
FJC0w2K/1ZldmQ2leHmiXSukDJAYmROy9Y1qkUazTzjsdvHGhUF2N1p7fIweNj8e
csR+T21MjPEwD99m5+xLvnMRMuqzH9TqVbFIM3lmCDajh8n9cp4KvGkQmB+X7DE1
R6AXG4EN9xn91TFrqmFFNOrFtoAjtag05q/HoqMhFFVeg+JTpsPshFjlWIkzwqKx
pn68KG2ztgS0KeDraGKwItTKengTCr/VkgorXnhKcI1C6C5iRXZp3wREu8RO+wRe
KSGbsYIHaFxd3XwW4JnsW+hes/W5MZX01wkwOLrktf85FjssBZBavxBbyFag/LvS
8oULOZRLYUkuElM+0Wzf8ayB574Fd97gzCVzWoD0Ei982jAdbEfk77PV1TvMNmEn
3M6ktB7GkjuD9OL12iNzxmbQe7p1WkYYps9hK4r0pbyxZPZlPMmNNZo579rywDjF
wEW5QkylaPEkbVDhJWeR1I8=
-----END CERTIFICATE-----
`
				configAfterConcourseDeploy.CredhubPassword = "f4b12bc0166cad1bc02b050e4e79ac4c"
				configAfterConcourseDeploy.CredhubURL = "https://77.77.77.77:8844/"
				configAfterConcourseDeploy.CredhubUsername = "credhub-cli"
			})

			JustBeforeEach(func() {
				configClient.NewConfigReturns(config.Config{
					ConfigBucket: "concourse-up-initial-deployment-eu-west-1-config",
					Deployment:   "concourse-up-initial-deployment",
					Namespace:    "",
					Project:      "initial-deployment",
					Region:       "eu-west-1",
					TFStatePath:  "terraform.tfstate",
				})
				configClient.HasAssetReturnsOnCall(0, false, nil)
				configClient.HasAssetReturnsOnCall(1, false, nil)
			})

			It("does the right things in the right order", func() {
				client := buildClient()
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				terraformInputVars := &terraform.AWSInputVars{
					NetworkCIDR:            defaultGeneratedConfig.NetworkCIDR,
					PublicCIDR:             defaultGeneratedConfig.PublicCIDR,
					PrivateCIDR:            defaultGeneratedConfig.PrivateCIDR,
					AllowIPs:               defaultGeneratedConfig.AllowIPs,
					AvailabilityZone:       defaultGeneratedConfig.AvailabilityZone,
					ConfigBucket:           defaultGeneratedConfig.ConfigBucket,
					Deployment:             defaultGeneratedConfig.Deployment,
					HostedZoneID:           defaultGeneratedConfig.HostedZoneID,
					HostedZoneRecordPrefix: defaultGeneratedConfig.HostedZoneRecordPrefix,
					Namespace:              defaultGeneratedConfig.Namespace,
					Project:                defaultGeneratedConfig.Project,
					PublicKey:              defaultGeneratedConfig.PublicKey,
					RDS1CIDR:               defaultGeneratedConfig.RDS1CIDR,
					RDS2CIDR:               defaultGeneratedConfig.RDS2CIDR,
					RDSDefaultDatabaseName: defaultGeneratedConfig.RDSDefaultDatabaseName,
					RDSInstanceClass:       defaultGeneratedConfig.RDSInstanceClass,
					RDSPassword:            defaultGeneratedConfig.RDSPassword,
					RDSUsername:            defaultGeneratedConfig.RDSUsername,
					Region:                 defaultGeneratedConfig.Region,
					SourceAccessIP:         defaultGeneratedConfig.SourceAccessIP,
					TFStatePath:            defaultGeneratedConfig.TFStatePath,
				}

				tfInputVarsFactory.NewInputVarsReturns(terraformInputVars)

				Expect(configClient).To(HaveReceived("ConfigExists"))
				Expect(configClient).ToNot(HaveReceived("Load"))
				Expect(tfInputVarsFactory).To(HaveReceived("NewInputVars").With(defaultGeneratedConfig))
				Expect(configClient).To(HaveReceived("Update").With(defaultGeneratedConfig))
				Expect(terraformCLI).To(HaveReceived("Apply").With(terraformInputVars, false))
				Expect(terraformCLI).To(HaveReceived("BuildOutput").With(terraformInputVars))
				Expect(configClient).To(HaveReceived("Update").With(configAfterLoad))

				Expect(certGenerationActions[0]).To(Equal("generating cert ca: concourse-up-initial-deployment, cn: [99.99.99.99 10.0.0.6]"))
				Expect(certGenerationActions[1]).To(Equal("generating cert ca: concourse-up-initial-deployment, cn: [77.77.77.77]"))

				Expect(configClient).To(HaveReceived("HasAsset").With("director-state.json"))
				Expect(configClient.HasAssetArgsForCall(0)).To(Equal("director-state.json"))
				Expect(configClient).To(HaveReceived("HasAsset").With("director-creds.yml"))
				Expect(configClient.HasAssetArgsForCall(1)).To(Equal("director-creds.yml"))
				Expect(boshClient).To(HaveReceived("Deploy").With([]byte{}, []byte{}, false))

				Expect(configClient).To(HaveReceived("StoreAsset").With("director-state.json", directorStateFixture))
				Expect(configClient).To(HaveReceived("StoreAsset").With("director-creds.yml", directorCredsFixture))
				Expect(boshClient).To(HaveReceived("Cleanup"))
				Expect(flyClient).To(HaveReceived("SetDefaultPipeline").With(configAfterCreateEnv, false))
				Expect(configClient).To(HaveReceived("Update").With(configAfterConcourseDeploy))
			})
		})

		It("Prints a warning about changing the sourceIP", func() {
			client := buildClient()
			err := client.Deploy()
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr).To(gbytes.Say("WARNING: allowing access from local machine"))
		})

		Context("When a custom domain was previously configured", func() {
			BeforeEach(func() {
				configInBucket.Domain = "ci.google.com"
			})

			JustBeforeEach(func() {
				configClient.LoadReturns(configInBucket, nil)
				configClient.ConfigExistsReturns(true, nil)
			})

			It("Prints a warning about adding a DNS record", func() {
				client := buildClient()
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				Expect(stderr).To(gbytes.Say("WARNING: adding record ci.google.com to DNS zone google.com with name ABC123"))
			})

			It("Generates certificates for that domain and not the public IP", func() {
				client := buildClient()
				err := client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				Expect(certGenerationActions).To(ContainElement("generating cert ca: concourse-up-happymeal, cn: [ci.google.com]"))
			})

			Context("and a custom cert is provided", func() {
				BeforeEach(func() {
					args.TLSCert = "--- CERTIFICATE ---"
					args.TLSKey = "--- KEY ---"
				})

				It("Prints the correct domain and not suggest using --insecure", func() {
					client := buildClient()
					err := client.Deploy()
					Expect(err).ToNot(HaveOccurred())
					Eventually(stdout).Should(gbytes.Say("DEPLOY SUCCESSFUL"))
					Eventually(stdout).Should(gbytes.Say("fly --target happymeal login --concourse-url https://ci.google.com --username admin --password s3cret"))
				})
			})
		})

		Context("When the user tries to change the region of an existing deployment", func() {
			BeforeEach(func() {
				args.Region = "eu-central-1"
			})

			JustBeforeEach(func() {
				configClient.LoadReturns(configInBucket, nil)
				configClient.ConfigExistsReturns(true, nil)
			})
			It("Returns a meaningful error message", func() {
				client := buildClientOtherRegion()
				err := client.Deploy()
				Expect(err).To(MatchError("found previous deployment in eu-west-1. Refusing to deploy to eu-central-1 as changing regions for existing deployments is not supported"))
			})
		})

		Context("When a custom DB instance size is not provided", func() {
			BeforeEach(func() {
				args.DBSize = "small"
				args.DBSizeIsSet = false
			})

			JustBeforeEach(func() {
				configClient.LoadReturns(configInBucket, nil)
				configClient.ConfigExistsReturns(true, nil)
			})
			It("Does not override the existing DB size", func() {
				provider, err := iaas.New(iaas.AWS, "eu-west-1")
				Expect(err).ToNot(HaveOccurred())
				awsInputVarsFactory, err := concourse.NewTFInputVarsFactory(provider)
				Expect(err).ToNot(HaveOccurred())

				var passedDBSize string
				tfInputVarsFactory.NewInputVarsStub = func(config config.Config) terraform.InputVars {
					passedDBSize = config.RDSInstanceClass
					return awsInputVarsFactory.NewInputVars(config)
				}

				client := buildClient()
				err = client.Deploy()
				Expect(err).ToNot(HaveOccurred())

				Expect(passedDBSize).To(Equal(configInBucket.RDSInstanceClass))
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

				Expect(boshClient).To(HaveReceived("Deploy").With([]byte{}, []byte{}, true))
			})
		})
	})
})
