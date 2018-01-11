package certs

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/square/certstrap/pkix"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/providers/dns/route53"
)

// Certs contains certificates and keys
type Certs struct {
	CACert []byte
	Key    []byte
	Cert   []byte
}
type user struct {
	k crypto.PrivateKey
	r *acme.RegistrationResource
	sync.Once
}

func (u *user) GetEmail() string {
	return "nobody@example.com"
}

func (u *user) GetRegistration() *acme.RegistrationResource {
	return u.r
}

func (u *user) GetPrivateKey() crypto.PrivateKey {
	u.Do(func() {
		var err error
		u.k, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}
	})
	return u.k
}

func hasIP(x []string) bool {
	for _, v := range x {
		if net.ParseIP(v) != nil {
			return true
		}
	}
	return false
}

type timeoutProvider struct {
	acme.ChallengeProvider
	timeout, interval time.Duration
}

func (t timeoutProvider) Timeout() (timeout, interval time.Duration) {
	return t.timeout, t.interval
}

func acmeURL() string {
	if u := os.Getenv("CONCOURSE_UP_ACME_URL"); u != "" {
		return u
	}
	return "https://acme-v01.api.letsencrypt.org/directory"
}

// Generate generates certs for use in a bosh director manifest
func Generate(caName string, ipOrDomains ...string) (*Certs, error) {
	if hasIP(ipOrDomains) {
		return generateSelfSigned(caName, ipOrDomains...)
	}
	u := &user{}
	c, err := acme.NewClient(acmeURL(), u, acme.RSA2048)
	if err != nil {
		return nil, err
	}
	c.ExcludeChallenges([]acme.Challenge{acme.HTTP01, acme.TLSSNI01})
	r53, err := route53.NewDNSProvider()
	if err != nil {
		return nil, err
	}
	c.SetChallengeProvider(acme.DNS01, timeoutProvider{
		r53,
		10 * time.Minute,
		1 * time.Second,
	})
	u.r, err = c.Register()
	if err != nil {
		return nil, err
	}
	c.AgreeToTOS()
	resp, errs := c.ObtainCertificate(ipOrDomains, true, nil, false)
	if len(errs) != 0 {
		return nil, fmt.Errorf("%v", errs)
	}
	return &Certs{
		CACert: resp.IssuerCertificate,
		Key:    resp.PrivateKey,
		Cert:   resp.Certificate,
	}, nil
}

// Generate generates certs for use in a bosh director manifest
func generateSelfSigned(caName string, ipOrDomains ...string) (*Certs, error) {
	caCert, caKey, err := generateCACert(caName)
	if err != nil {
		return nil, err
	}

	csr, key, err := generateCertificateSigningRequest(ipOrDomains)
	if err != nil {
		return nil, err
	}

	cert, err := signCSR(csr, caCert, caKey)
	if err != nil {
		return nil, err
	}

	caCertByte, err := caCert.Export()
	if err != nil {
		return nil, err
	}

	keyBytes, err := key.ExportPrivate()
	if err != nil {
		return nil, err
	}

	certBytes, err := cert.Export()
	if err != nil {
		return nil, err
	}

	return &Certs{
		CACert: caCertByte,
		Key:    keyBytes,
		Cert:   certBytes,
	}, nil
}

func signCSR(csr *pkix.CertificateSigningRequest, caCert *pkix.Certificate, caCertKey *pkix.Key) (*pkix.Certificate, error) {
	crtOut, err := pkix.CreateCertificateHost(caCert, caCertKey, csr, 2)
	if err != nil {
		return nil, err
	}

	return crtOut, nil
}

func generateCACert(caName string) (*pkix.Certificate, *pkix.Key, error) {
	key, err := pkix.CreateRSAKey(4096)
	if err != nil {
		return nil, nil, err
	}

	caCert, err := pkix.CreateCertificateAuthority(
		key,
		"",
		10,
		"",
		"",
		"",
		"",
		caName,
	)
	if err != nil {
		return nil, nil, err
	}

	return caCert, key, nil
}

func generateCertificateSigningRequest(ipOrDomains []string) (*pkix.CertificateSigningRequest, *pkix.Key, error) {
	var name string
	domains := []string{}
	ips := []net.IP{}

	for _, ipOrDomain := range ipOrDomains {
		if name == "" {
			name = ipOrDomain
		}

		if parsedIPs, err := pkix.ParseAndValidateIPs(ipOrDomain); err != nil {
			// if invalid IP address, assume domain instead
			if strings.Contains(err.Error(), "Invalid IP address") {
				domains = append(domains, ipOrDomain)
			} else {
				return nil, nil, err
			}
		} else {
			ips = append(ips, parsedIPs[0])
		}
	}

	key, err := pkix.CreateRSAKey(2048)
	if err != nil {
		return nil, nil, err
	}

	csr, err := pkix.CreateCertificateSigningRequest(
		key,
		"",
		ips,
		domains,
		"",
		"",
		"",
		"",
		name,
	)
	if err != nil {
		return nil, nil, err
	}

	return csr, key, nil
}
