package director

import (
	"crypto/x509"
	"encoding/pem"
	gonet "net"
	gourl "net/url"
	"strconv"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Config struct {
	Host string
	Port int

	// CA certificate is not required
	CACert string

	Client       string
	ClientSecret string

	TokenFunc func(bool) (string, error)
}

func NewConfigFromURL(url string) (Config, error) {
	if len(url) == 0 {
		return Config{}, bosherr.Error("Expected non-empty Director URL")
	}

	parsedURL, err := gourl.Parse(url)
	if err != nil {
		return Config{}, bosherr.WrapErrorf(err, "Parsing Director URL '%s'", url)
	}

	host := parsedURL.Host
	port := 25555

	if len(host) == 0 {
		host = url
	}

	if strings.Contains(host, ":") {
		var portStr string

		host, portStr, err = gonet.SplitHostPort(host)
		if err != nil {
			return Config{}, bosherr.WrapErrorf(
				err, "Extracting host/port from URL '%s'", parsedURL.Host)
		}

		port, err = strconv.Atoi(portStr)
		if err != nil {
			return Config{}, bosherr.WrapErrorf(
				err, "Extracting port from URL '%s'", parsedURL.Host)
		}
	}

	if len(host) == 0 {
		return Config{}, bosherr.Errorf("Expected to extract host from URL '%s'", url)
	}

	return Config{Host: host, Port: port}, nil
}

func (c Config) Validate() error {
	if len(c.Host) == 0 {
		return bosherr.Error("Missing 'Host'")
	}

	if c.Port == 0 {
		return bosherr.Error("Missing 'Port'")
	}

	if _, err := c.CACertPool(); err != nil {
		return err
	}

	// Don't validate credentials since Info call does not require authentication.

	return nil
}

func (c Config) CACertPool() (*x509.CertPool, error) {
	if len(c.CACert) == 0 {
		return nil, nil
	}

	certPool := x509.NewCertPool()

	block, _ := pem.Decode([]byte(c.CACert))
	if block == nil {
		return nil, bosherr.Error("Parsing CA certificate: Missing PEM block")
	}

	if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
		return nil, bosherr.Error("Parsing CA certificate: Not a certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, bosherr.WrapError(err, "Parsing CA certificate")
	}

	certPool.AddCert(cert)

	return certPool, nil
}
