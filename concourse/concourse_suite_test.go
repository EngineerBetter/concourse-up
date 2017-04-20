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
	FakeLoad         func(deployment string) (*config.Config, error)
	FakeLoadOrCreate func(deployment string) (*config.Config, error)
}

func (client *FakeConfigClient) Load(deployment string) (*config.Config, error) {
	return client.FakeLoad(deployment)
}

func (client *FakeConfigClient) LoadOrCreate(deployment string) (*config.Config, error) {
	return client.FakeLoadOrCreate(deployment)
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
