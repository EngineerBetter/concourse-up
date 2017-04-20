package concourse_test

import (
	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConcourse(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Concourse Suite")
}

type FakeConfigClient struct {
	FakeLoad         func(project string) (*config.Config, error)
	FakeLoadOrCreate func(project string) (*config.Config, error)
	FakeStoreAsset   func(project, filename string, contents []byte) error
}

func (client *FakeConfigClient) Load(project string) (*config.Config, error) {
	return client.FakeLoad(project)
}

func (client *FakeConfigClient) LoadOrCreate(project string) (*config.Config, error) {
	return client.FakeLoadOrCreate(project)
}

func (client *FakeConfigClient) StoreAsset(project, filename string, contents []byte) error {
	return client.FakeStoreAsset(project, filename, contents)
}

type FakeTerraformClient struct {
	FakeOutput  func() (*terraform.Metadata, error)
	FakeApply   func() error
	FakeDestroy func() error
	FakeCleanup func() error
}

func (client *FakeTerraformClient) Output() (*terraform.Metadata, error) {
	return client.FakeOutput()
}

func (client *FakeTerraformClient) Apply() error {
	return client.FakeApply()
}

func (client *FakeTerraformClient) Destroy() error {
	return client.FakeDestroy()
}

func (client *FakeTerraformClient) Cleanup() error {
	return client.FakeCleanup()
}

type FakeBoshInitClient struct {
	FakeDeploy func() ([]byte, error)
	FakeDelete func() error
}

func (client *FakeBoshInitClient) Deploy() ([]byte, error) {
	return client.FakeDeploy()
}

func (client *FakeBoshInitClient) Delete() error {
	return client.FakeDelete()
}
