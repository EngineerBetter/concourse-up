package iaas_test

import (
	. "github.com/EngineerBetter/concourse-up/iaas"

	"errors"
	"github.com/EngineerBetter/concourse-up/iaas/iaasfakes"
	"github.com/aws/aws-sdk-go/service/route53"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

const AwsAccessKeyIdentifier = "AWS_ACCESS_KEY_ID"
const AwsSecretAccessKeyIdentifier = "AWS_SECRET_ACCESS_KEY"

var (
	awsKeyToRemember       string
	awsSecretKeyToRemember string
	route53Fake            *iaasfakes.FakeRoute53
)

var _ = Describe("Client#FindLongestMatchingHostedZone", func() {

	BeforeEach(func() {
		temporaryInsertTestEnvironmentValues()
		route53Fake = new(iaasfakes.FakeRoute53)
	})

	AfterEach(func() {
		resetEnvironmentValues()
	})

	Context("When the hosted zone exists", func() {
		It("Returns the hosted zone details", func() {
			awsClient, err := New("AWS", "eu-west-1")
			Expect(err).To(Succeed())

			route53Fake.ListHostedZonesPagesStub = func(input *route53.ListHostedZonesInput, callback func(output *route53.ListHostedZonesOutput, lastPage bool) (shouldContinue bool)) error {
				createFakeEnvironmentAndHandToCallback(callback)
				return nil
			}

			zoneName, zoneID, err := (awsClient).FindLongestMatchingHostedZone("integration-test.concourse-up.engineerbetter.com", route53Fake)

			Expect(err).ToNot(HaveOccurred())
			Expect(zoneName).To(Equal("concourse-up.engineerbetter.com"))
			Expect(zoneID).To(Equal("Z2NEMKRYH9QASG"))
		})
	})

	Context("When the hosted zone does not exist", func() {
		It("Returns a meaningful error", func() {
			awsClient, err := New("AWS", "eu-west-1")
			Expect(err).To(Succeed())

			fakeError := errors.New("fake error")
			route53Fake.ListHostedZonesPagesReturns(fakeError)

			_, _, err = (awsClient).FindLongestMatchingHostedZone("abc.google.com", route53Fake)
			Expect(err).To(Equal(fakeError))
		})
	})
})

func createFakeEnvironmentAndHandToCallback(callback func(output *route53.ListHostedZonesOutput, lastPage bool) (shouldContinue bool)) {
	fakeZone := new(route53.ListHostedZonesOutput)
	fakeName := "concourse-up.engineerbetter.com"
	fakeId := "Z2NEMKRYH9QASG"
	zone := route53.HostedZone{Id: &fakeId, Name: &fakeName}
	hostedZones := []*route53.HostedZone{&zone}
	fakeZone.HostedZones = hostedZones
	callback(fakeZone, true)
}

func resetEnvironmentValues() {
	os.Setenv(AwsAccessKeyIdentifier, awsKeyToRemember)
	os.Setenv(AwsSecretAccessKeyIdentifier, awsSecretKeyToRemember)
}

func temporaryInsertTestEnvironmentValues() {
	awsKeyToRemember = os.Getenv(AwsAccessKeyIdentifier)
	awsSecretKeyToRemember = os.Getenv(AwsSecretAccessKeyIdentifier)
	os.Setenv(AwsAccessKeyIdentifier, "AWS_ACCESS_KEY_ID_VALUE")
	os.Setenv(AwsSecretAccessKeyIdentifier, "AWS_SECRET_ACCESS_KEY_VALUE")
}
