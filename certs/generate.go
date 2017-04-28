package certs

import (
	"net"
	"strings"

	"github.com/square/certstrap/pkix"
)

// Certs contains certificates and keys
type Certs struct {
	CACert []byte
	Key    []byte
	Cert   []byte
}

// Generate generates certs for use in a bosh director manifest
func Generate(caName string, ipOrDomains ...string) (*Certs, error) {
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
