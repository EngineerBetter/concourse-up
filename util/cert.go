package util

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"time"
)

// GenerateCertificate generates a pem encoded certificate and key
func GenerateCertificate(validity time.Duration, organization string, hosts []string, caCert, caKey []byte) (cert, key []byte, err error) {
	const rsaBits = 2048
	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(validity)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{organization},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	var parsedCACert *x509.Certificate
	var parsedCAKey *rsa.PrivateKey
	if caCert == nil {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
		parsedCACert = &template
		parsedCAKey = priv
	} else {
		block, _ := pem.Decode(caCert)
		if block == nil {
			err = errors.New("failed to parse PEM block containing the ca certificate key")
			return
		}
		parsedCACert, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return
		}
		block, _ = pem.Decode(caKey)
		if block == nil {
			err = errors.New("failed to parse PEM block containing the ca private key key")
			return
		}
		parsedCAKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, parsedCACert, priv.PublicKey, parsedCAKey)
	if err != nil {
		return
	}

	var certOut bytes.Buffer
	err = pem.Encode(&certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return
	}

	var keyOut bytes.Buffer
	err = pem.Encode(&keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	cert = certOut.Bytes()
	key = keyOut.Bytes()
	return
}
