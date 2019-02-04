package concoursefakes

import (
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/terraform"
)

type FakeMaintainer struct {
	fakeGetTFInputVars getTFInputVarsFn
}

type getTFInputVarsFn = func(client *concourse.Client, conf config.Config, environment terraform.InputVars) (terraform.InputVars, error)

func NewFakeMaintainer() *FakeMaintainer {
	impl := func(client *concourse.Client, conf config.Config, environment terraform.InputVars) (terraform.InputVars, error) {
		return environment, nil
	}
	return &FakeMaintainer{fakeGetTFInputVars: impl}
}

func (m *FakeMaintainer) GetTFInputVars(client *concourse.Client, conf config.Config, environment terraform.InputVars) (terraform.InputVars, error) {
	return m.fakeGetTFInputVars(client, conf, environment)
}
