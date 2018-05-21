package iaas_test

import (
	"os"

	. "github.com/EngineerBetter/concourse-up/iaas"
	"github.com/aws/aws-sdk-go/service/route53"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client#FindLongestMatchingHostedZone", func() {

	BeforeEach(func() {
		os.Setenv("AWS_SECRET_ACCESS_KEY", "123")
		os.Setenv("AWS_ACCESS_KEY_ID", "123")
	})

	// var listHostedZones = iaas.ListHostedZones
	var listHostedZonesFound = func() ([]*route53.HostedZone, error) {
		zones := []*route53.HostedZone{}
		zone := &route53.HostedZone{}
		zone = zone.SetName("concourse-up.engineerbetter.com")
		zone = zone.SetId("Z2NEMKRYH9QASG")
		zones = append(zones, zone)

		return zones, nil
	}

	var listHostedZonesNotFound = func() ([]*route53.HostedZone, error) {
		return nil, nil
	}

	Context("When the hosted zone exists", func() {
		It("Returns the hosted zone details", func() {
			awsClient, err := New("AWS", "eu-west-1")
			Expect(err).To(Succeed())
			zoneName, zoneID, err := (awsClient).FindLongestMatchingHostedZone("integration-test.concourse-up.engineerbetter.com", listHostedZonesFound)
			Expect(err).ToNot(HaveOccurred())
			Expect(zoneName).To(Equal("concourse-up.engineerbetter.com"))
			Expect(zoneID).To(Equal("Z2NEMKRYH9QASG"))
		})
	})

	Context("When the hosted zone does not exist", func() {
		It("Returns a meaningful error", func() {
			awsClient, err := New("AWS", "eu-west-1")
			Expect(err).To(Succeed())
			_, _, err = (awsClient).FindLongestMatchingHostedZone("abc.google.com", listHostedZonesNotFound)
			Expect(err).To(MatchError("No matching hosted zone found for domain abc.google.com"))
		})
	})
})
