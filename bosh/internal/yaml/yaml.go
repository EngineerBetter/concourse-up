package yaml

import (
	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/cppforlife/go-patch/patch"
	yamlenc "github.com/ghodss/yaml"
)

func newOpsFromString(ops string) (patch.Op, error) {
	var opDefs []patch.OpDefinition
	err := yamlenc.Unmarshal([]byte(ops), &opDefs)
	if err != nil {
		return nil, err
	}
	return patch.NewOpsFromDefinitions(opDefs)
}

func Interpolate(s string, ops string, vars map[string]interface{}) (string, error) {
	t := template.NewTemplate([]byte(s))
	op, err := newOpsFromString(ops)
	if err != nil {
		return "", err
	}
	x, err := t.Evaluate(template.StaticVariables(vars), op, template.EvaluateOpts{
		// ExpectAllKeys:     true,
		// ExpectAllVarsUsed: true,
	})
	return string(x), err
}
