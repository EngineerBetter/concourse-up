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
func Generate(caName, ipOrDomain string) (*Certs, error) {
	caCert, caKey, err := generateCACert(caName)
	if err != nil {
		return nil, err
	}

	csr, key, err := generateCertificateSigningRequest(ipOrDomain)
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

func generateCertificateSigningRequest(ipOrDomain string) (*pkix.CertificateSigningRequest, *pkix.Key, error) {
	var name string
	var domains []string

	ips, err := pkix.ParseAndValidateIPs(ipOrDomain)
	if err != nil {
		// if invalid IP address, assume domain instead
		if strings.Contains(err.Error(), "Invalid IP address") {
			name = ipOrDomain
			domains = []string{ipOrDomain}
			ips = []net.IP{}
		} else {
			return nil, nil, err
		}
	} else {
		name = ips[0].String()
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
