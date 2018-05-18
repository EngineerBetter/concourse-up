package certs_test

import (
	"crypto"
	"errors"

	. "github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/util"
	"github.com/xenolf/lego/acme"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Client struct {
}

func (c *Client) SetChallengeProvider(challenge acme.Challenge, p acme.ChallengeProvider) error {
	return nil
}

func (c *Client) ExcludeChallenges(challenges []acme.Challenge) {
}

func (c *Client) Register() (*acme.RegistrationResource, error) {
	return nil, nil
}

func (c *Client) AgreeToTOS() error {
	return nil
}

func (c *Client) ObtainCertificate(domains []string, bundle bool, privKey crypto.PrivateKey, mustStaple bool) (acme.CertificateResource, map[string]error) {
	if domains[0] == "google.com" {
		errs := make(map[string]error)
		errs["error"] = errors.New("this is an error")
		return acme.CertificateResource{}, errs
	}
	return acme.CertificateResource{
		PrivateKey:        []byte("BEGIN RSA PRIVATE KEY"),
		Certificate:       []byte("BEGIN CERTIFICATE"),
		IssuerCertificate: []byte("BEGIN CERTIFICATE"),
	}, nil
}

func makeFakeClient() *Client {
	return &Client{}
}

var _ = Describe("Certs", func() {
	It("Generates a cert for an IP address", func() {
		c := makeFakeClient()
		certs, err := Generate(c, "concourse-up-mole", "99.99.99.99")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(certs.CACert)).To(ContainSubstring("BEGIN CERTIFICATE"))
		Expect(string(certs.Key)).To(ContainSubstring("BEGIN RSA PRIVATE KEY"))
		Expect(string(certs.Cert)).To(ContainSubstring("BEGIN CERTIFICATE"))
	})
	It("Generates a cert for a domain", func() {
		c := makeFakeClient()
		certs, err := Generate(c, "concourse-up-mole", "concourse-up-test-"+util.GeneratePasswordWithLength(10)+".engineerbetter.com")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(certs.CACert)).To(ContainSubstring("BEGIN CERTIFICATE"))
		Expect(string(certs.Key)).To(ContainSubstring("BEGIN RSA PRIVATE KEY"))
		Expect(string(certs.Cert)).To(ContainSubstring("BEGIN CERTIFICATE"))
	})
	It("Can't generate a cert for google.com", func() {
		c := makeFakeClient()
		_, err := Generate(c, "concourse-up-mole", "google.com")
		Expect(err).To(HaveOccurred())
	})
})
