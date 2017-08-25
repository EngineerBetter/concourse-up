package terraform_test

import (
	"fmt"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/iaas"
	. "github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Plan", func() {
	var bucket string
	var iaasClient iaas.IClient
	var conf *config.Config

	BeforeEach(func() {
		var err error
		iaasClient, err = iaas.New("AWS", "eu-west-1")
		Expect(err).ToNot(HaveOccurred())
		bucket = fmt.Sprintf("concourse-up-integration-tests-%s", util.GeneratePassword())

		err = iaasClient.EnsureBucketExists(bucket)
		Expect(err).ToNot(HaveOccurred())

		conf = &config.Config{
			ConfigBucket:           bucket,
			AvailabilityZone:       "eu-west-1a",
			RDSInstanceClass:       "t2.small",
			SourceAccessIP:         "1.2.3.4",
			RDSDefaultDatabaseName: "bosh",
			PublicKey:              "example-public-key",
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
		}
	})

	AfterEach(func() {
		err := iaasClient.DeleteVersionedBucket(bucket)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Works", func() {
		stdout := gbytes.NewBuffer()
		stderr := gbytes.NewBuffer()

		c, err := NewClient(conf, stdout, stderr)
		Expect(err).ToNot(HaveOccurred())
		defer c.Cleanup()

		err = c.Apply(true)
		Expect(err).ToNot(HaveOccurred())
		Eventually(stdout).Should(gbytes.Say("Plan"))
	})
})
