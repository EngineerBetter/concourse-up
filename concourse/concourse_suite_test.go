package concourse_test

import (
	"bitbucket.org/engineerbetter/concourse-up/config"
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
