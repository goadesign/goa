package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// Enum validation compiled template
var enumValidationTemplateC *template.Template

// Generate enum validation code.
func EnumValidationCode(values []interface{}) (string, error) {
	if enumValidationTemplateC == nil {
		var err error
		enumValidationTemplateC, err = template.New("enum validation").Funcs(funcMap).Parse(enumValidationTemplate)
		if err != nil {
			return "", fmt.Errorf("failed to instantiate enum validation template: %s", err)
		}
	}
	var b bytes.Buffer
	err := enumValidationTemplateC.Execute(&b, values)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// Template generation helpers
var funcMap = template.FuncMap{
	"join": strings.Join,
}

const enumValidationTemplate = `
func validateEnum(val interface{}) error { {{range .}}
	if val == "{{.}}" {
		return nil
	}{{end}}
	return fmt.Errorf("invalid value \"%s\": allowed values are {{join . ", "}}", val)
}
`
