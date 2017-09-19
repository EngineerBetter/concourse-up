package config_test

import (
	. "github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/testsupport"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var iaasClient *testsupport.FakeAWSClient
	var client *Client
	var deployArgs *DeployArgs

	BeforeEach(func() {
		iaasClient = &testsupport.FakeAWSClient{
			FakeRegion: func() string {
				return "eu-west-1"
			},
			FakeEnsureBucketExists: func(name string) error {
				return nil
			},
			FakeEnsureFileExists: func(bucket, path string, defaultContents []byte) ([]byte, bool, error) {
				return defaultContents, true, nil
			},
		}
		client = New(iaasClient, "test-project")

		deployArgs = &DeployArgs{
			IAAS:        "AWS",
			AWSRegion:   "eu-west-1",
			WorkerCount: 1,
			WorkerSize:  "xlarge",
			DBSize:      "medium",
			DBSizeIsSet: false,
		}
	})

	Describe("LoadOrCreate", func() {
		Context("When there is no existing config file", func() {
			It("Sets the RDS instance size in the default config to the given size", func() {
				conf, createdANewFile, err := client.LoadOrCreate(deployArgs)
				Expect(err).To(Succeed())

				Expect(conf.RDSInstanceClass).To(Equal("db.t2.medium"))
				Expect(createdANewFile).To(BeTrue())
			})
		})
	})
})
