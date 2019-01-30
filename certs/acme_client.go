package certs

import (
	"os"

	"github.com/xenolf/lego/lego"
)

// NewAcmeClient returns a new AcmeClient
func NewAcmeClient(u *User) (*lego.Client, error) {

	var (
		c  *lego.Config
		cl *lego.Client
	)

	c = lego.NewConfig(u)
	c.CADirURL = acmeURL()

	cl, err := lego.NewClient(c)
	if err != nil {
		return nil, err
	}

	return cl, nil
}

func acmeURL() string {
	if u := os.Getenv("CONCOURSE_UP_ACME_URL"); u != "" {
		return u
	}
	return lego.LEDirectoryProduction
}
