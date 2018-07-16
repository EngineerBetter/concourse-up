package iaas_test

import (
	"os"

	"github.com/EngineerBetter/concourse-up/iaas"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client#CheckForWhitelistedIP", func() {

	BeforeEach(func() {
		os.Setenv("AWS_SECRET_ACCESS_KEY", "123")
		os.Setenv("AWS_ACCESS_KEY_ID", "123")
	})

	var fakeEC2ClientCreator = func() (iaas.IEC2, error) {
		return &fakeEC2Client{}, nil
	}

	Context("When the IP is not found", func() {
		It("returns false", func() {
			awsClient, err := iaas.New("AWS", "eu-west-1")
			Expect(err).To(Succeed())
			ip := "0.0.0.0"
			securityGroup := "something"
			result, err := awsClient.CheckForWhitelistedIP(ip, securityGroup, fakeEC2ClientCreator)
			Expect(err).To(Succeed())
			Expect(result).To(BeFalse())
		})
	})

	Context("When the IP is whitelisted", func() {
		It("returns true", func() {
			awsClient, err := iaas.New("AWS", "eu-west-1")
			Expect(err).To(Succeed())
			ip := "1.2.3.4"
			securityGroup := "something"
			result, err := awsClient.CheckForWhitelistedIP(ip, securityGroup, fakeEC2ClientCreator)
			Expect(err).To(Succeed())
			Expect(result).To(BeTrue())
		})
	})

	Context("When the IP is not whitelisted on all ports", func() {
		It("returns true", func() {
			awsClient, err := iaas.New("AWS", "eu-west-1")
			Expect(err).To(Succeed())
			ip := "5.6.7.8"
			securityGroup := "something"
			result, err := awsClient.CheckForWhitelistedIP(ip, securityGroup, fakeEC2ClientCreator)
			Expect(err).To(Succeed())
			Expect(result).To(BeFalse())
		})
	})
})
