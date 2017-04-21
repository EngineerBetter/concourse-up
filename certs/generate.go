package certs

import (
	"errors"

	"github.com/square/certstrap/pkix"
)

type Certs struct {
	CACert []byte
	Key    []byte
	Cert   []byte
}

func Generate(caName, ip string) (*Certs, error) {
	caCert, caKey, err := generateCACert(caName)
	if err != nil {
		return nil, err
	}

	csr, key, err := generateCertificateSigningRequest(ip)
	if err != nil {
		return nil, err
	}

	cert, err := signCSR(csr, caCert, caKey)
	if err != nil {
		return nil, err
	}

	return &Certs{
		CACert: caCert,
		Key:    key,
		Cert:   cert,
	}, nil
}

func signCSR(csrBytes, caCertBytes, caCertKeyBytes []byte) ([]byte, error) {
	csr, err := pkix.NewCertificateSigningRequestFromPEM(csrBytes)
	if err != nil {
		return nil, err
	}
	caCert, err := pkix.NewCertificateFromPEM(caCertBytes)
	if err != nil {
		return nil, err
	}
	rawCert, err := caCert.GetRawCertificate()
	if err != nil {
		return nil, err
	}

	if !rawCert.IsCA {
		return nil, errors.New("raw is CA!")
	}

	caCertKey, err := pkix.NewKeyFromPrivateKeyPEM(caCertKeyBytes)
	if err != nil {
		return nil, err
	}

	crtOut, err := pkix.CreateCertificateHost(caCert, caCertKey, csr, 2)
	if err != nil {
		return nil, err
	}

	return crtOut.Export()
}

func generateCACert(caName string) ([]byte, []byte, error) {
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

	caCertBytes, err := caCert.Export()
	if err != nil {
		return nil, nil, err
	}

	caCertKeyBytes, err := key.ExportPrivate()
	if err != nil {
		return nil, nil, err
	}

	return caCertBytes, caCertKeyBytes, nil
}

func generateCertificateSigningRequest(ip string) ([]byte, []byte, error) {
	ips, err := pkix.ParseAndValidateIPs(ip)
	if err != nil {
		return nil, nil, err
	}

	key, err := pkix.CreateRSAKey(2048)
	if err != nil {
		return nil, nil, err
	}

	name := ips[0].String()

	csr, err := pkix.CreateCertificateSigningRequest(
		key,
		"",
		ips,
		[]string{},
		"",
		"",
		"",
		"",
		name,
	)
	if err != nil {
		return nil, nil, err
	}

	csrBytes, err := csr.Export()
	if err != nil {
		return nil, nil, err
	}

	keyBytes, err := key.ExportPrivate()
	if err != nil {
		return nil, nil, err
	}

	return csrBytes, keyBytes, nil
}
