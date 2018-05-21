package certs

import (
	"os"

	"github.com/xenolf/lego/acme"
)

// NewAcmeClient returns a new AcmeClient
func NewAcmeClient(u *User) (AcmeClient, error) {

	var c AcmeClient
	c, err := acme.NewClient(acmeURL(), u, acme.RSA2048)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func acmeURL() string {
	if u := os.Getenv("CONCOURSE_UP_ACME_URL"); u != "" {
		return u
	}
	return "https://acme-v01.api.letsencrypt.org/directory"
}
