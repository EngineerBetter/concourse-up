package util

import (
	"bytes"
	"io/ioutil"
	"text/template"
)

const leftDelim = "<%"
const rightDelim = "%>"

// RenderTemplate renders a template to a string
func RenderTemplate(templateStr string, params interface{}) ([]byte, error) {
	templ, err := template.New("template").Delims(leftDelim, rightDelim).Parse(templateStr)
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
