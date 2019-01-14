package yaml

import (
	"strings"

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

// Interpolate returns an interpolated string using vars
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

// Path might find what you are looking for
func Path(b []byte, path string) (string, error) {
	t := template.NewTemplate(b)
	op, err := patch.NewOpsFromDefinitions([]patch.OpDefinition{})
	if err != nil {
		return "", err
	}
	tokens := []patch.Token{patch.RootToken{}}
	for _, token := range strings.Split(path, "/") {
		tokens = append(tokens, patch.KeyToken{Key: token})
	}
	findOp := patch.FindOp{
		Path: patch.NewPointer(tokens),
	}

	x, err := t.Evaluate(template.StaticVariables(make(map[string]interface{})), op, template.EvaluateOpts{
		PostVarSubstitutionOp: findOp,
	})
	return string(x), err
}
