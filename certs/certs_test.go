// +build integration

package certs_test

import (
	. "github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/testsupport"
	"github.com/EngineerBetter/concourse-up/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Certs", func() {
	var constructor = testsupport.NewFakeAcmeClient
	var provider = testsupport.FakeProvider{}
	It("Generates a cert for an IP address", func() {
		certs, err := Generate(constructor, "concourse-up-mole", &provider, "99.99.99.99")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(certs.CACert)).To(ContainSubstring("BEGIN CERTIFICATE"))
		Expect(string(certs.Key)).To(ContainSubstring("BEGIN RSA PRIVATE KEY"))
		Expect(string(certs.Cert)).To(ContainSubstring("BEGIN CERTIFICATE"))
	})
	It("Generates a cert for a domain", func() {
		certs, err := Generate(constructor, "concourse-up-mole", &provider, "concourse-up-test-"+util.GeneratePasswordWithLength(10)+".engineerbetter.com")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(certs.CACert)).To(ContainSubstring("BEGIN CERTIFICATE"))
		Expect(string(certs.Key)).To(ContainSubstring("BEGIN RSA PRIVATE KEY"))
		Expect(string(certs.Cert)).To(ContainSubstring("BEGIN CERTIFICATE"))
	})
	It("Can't generate a cert for google.com", func() {
		_, err := Generate(constructor, "concourse-up-mole", &provider, "google.com")
		Expect(err).To(HaveOccurred())
	})
})
