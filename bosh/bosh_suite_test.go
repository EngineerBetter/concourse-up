package bosh_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBosh(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bosh Suite")
}

type FakeDirectorClient struct {
	FakeRunCommand              func(args ...string) ([]byte, error)
	FakeRunAuthenticatedCommand func(args ...string) ([]byte, error)
	FakeSaveFileToWorkingDir    func(path string, contents []byte) (string, error)
	FakePathInWorkingDir        func(filename string) string
	FakeCleanup                 func() error
}

func (client *FakeDirectorClient) RunCommand(args ...string) ([]byte, error) {
	return client.FakeRunCommand(args...)
}

func (client *FakeDirectorClient) RunAuthenticatedCommand(args ...string) ([]byte, error) {
	return client.FakeRunAuthenticatedCommand(args...)
}

func (client *FakeDirectorClient) SaveFileToWorkingDir(filename string, contents []byte) (string, error) {
	return client.FakeSaveFileToWorkingDir(filename, contents)
}

func (client *FakeDirectorClient) PathInWorkingDir(filename string) string {
	return client.FakePathInWorkingDir(filename)
}

func (client *FakeDirectorClient) Cleanup() error {
	return client.FakeCleanup()
}
