package terraformfakes

import (
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/terraform"
)

type FakeTFInputVarsFactory struct {
	NewInputVarsFn NewInputVarsFn
}

type NewInputVarsFn func(c config.Config) terraform.InputVars

func (f *FakeTFInputVarsFactory) NewInputVars(c config.Config) terraform.InputVars {
	return f.NewInputVarsFn(c)
}
