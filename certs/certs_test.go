package certs_test

import (
	. "bitbucket.org/engineerbetter/concourse-up/certs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Certs", func() {
	It("Generates a cert", func() {
		certs, err := Generate("concourse-up-mole", "54.72.120.75")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(certs.CACert)).To(ContainSubstring("BEGIN CERTIFICATE"))
		Expect(string(certs.Key)).To(ContainSubstring("BEGIN RSA PRIVATE KEY"))
		Expect(string(certs.Cert)).To(ContainSubstring("BEGIN CERTIFICATE"))
	})
})
