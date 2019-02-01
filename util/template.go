package util

import (
	"bytes"
	"io/ioutil"
	"text/template"
)

// RenderTemplate renders a template to a string
func RenderTemplate(name string, templateStr string, params interface{}) ([]byte, error) {
	templ, err := template.New(name).Parse(templateStr)
	if err != nil {
		return nil, err
	}

	buffer := new(bytes.Buffer)

	if err = templ.Execute(buffer, params); err != nil {
		return nil, err
	}

	outputBytes, err := ioutil.ReadAll(buffer)
	if err != nil {
		return nil, err
	}

	return outputBytes, nil
}
