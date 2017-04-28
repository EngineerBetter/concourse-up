package bosh_test

import (
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBosh(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bosh Suite")
}

type FakeDirectorClient struct {
	FakeRunCommand              func(stdout, stderr io.Writer, args ...string) error
	FakeRunAuthenticatedCommand func(stdout, stderr io.Writer, args ...string) error
	FakeSaveFileToWorkingDir    func(path string, contents []byte) (string, error)
	FakePathInWorkingDir        func(filename string) string
	FakeCleanup                 func() error
}

func (client *FakeDirectorClient) RunCommand(stdout, stderr io.Writer, args ...string) error {
	return client.FakeRunCommand(stdout, stderr, args...)
}

func (client *FakeDirectorClient) RunAuthenticatedCommand(stdout, stderr io.Writer, args ...string) error {
	return client.FakeRunAuthenticatedCommand(stdout, stderr, args...)
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
