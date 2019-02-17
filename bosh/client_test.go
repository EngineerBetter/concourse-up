package bosh_test

import (
	"errors"
	"io"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/bosh/internal/boshcli/boshclifakes"
	"github.com/EngineerBetter/concourse-up/bosh/internal/workingdir/workingdirfakes"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/iaas/iaasfakes"
	"github.com/EngineerBetter/concourse-up/terraform/terraformfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Client", func() {
	var buildClient func() bosh.IClient
	var boshCLI *boshclifakes.FakeICLI
	var directorClient *workingdirfakes.FakeIClient
	var stdout, stderr *gbytes.Buffer
	var configInput config.Config

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

	BeforeEach(func() {
		configInput = config.Config{
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
			PublicKey: "example-public-key",
		}
	})

	Describe("Instances", func() {
		Context("When on AWS", func() {
			JustBeforeEach(func() {
				boshCLI = &boshclifakes.FakeICLI{}
				directorClient = &workingdirfakes.FakeIClient{}
				outputs := &terraformfakes.FakeOutputs{}
				provider := setupFakeAwsProvider()

				stdout = gbytes.NewBuffer()
				stderr = gbytes.NewBuffer()

				buildClient = func() bosh.IClient {
					client, err := bosh.NewAWSClient(configInput, outputs, directorClient, stdout, stderr, provider, boshCLI)
					Expect(err).ToNot(HaveOccurred())
					return client
				}
			})
			Context("When instances are found", func() {
				JustBeforeEach(func() {
					boshCLI.RunAuthenticatedCommandStub = func(action, ip, password, ca string, detach bool, stdout io.Writer, flags ...string) error {
						stdout.Write([]byte("{\"Tables\":[{\"Rows\": [{\"instance\": \"foo\",\"ips\": \"1.2.3.4\", \"process_state\": \"bar\"}]}]}"))
						return nil
					}
				})
				It("returns them", func() {
					expectedInstance := bosh.Instance{
						Name:  "foo",
						IP:    "1.2.3.4",
						State: "bar",
					}

					client := buildClient()
					instances, err := client.Instances()
					Expect(err).ToNot(HaveOccurred())

					Expect(instances).To(Equal([]bosh.Instance{expectedInstance}))
				})
			})
		})
	})
})
