package testsupport

import (
	"github.com/xenolf/lego/lego"

	"github.com/EngineerBetter/concourse-up/certs"
)

// NewFakeAcmeClient returns a new FakeAcmeClient
func NewFakeAcmeClient(u *certs.User) (*lego.Client, error) {
	return &lego.Client{}, nil
}
