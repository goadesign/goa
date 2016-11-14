// This file verifies that the specs produce the expected data structures.
package dsl_test

import (
	"testing"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/eval"
)

func TestAPISpec(t *testing.T) {
	api := design.Root.API
	if !eval.Execute(api.DSL(), api) {
		t.Errorf("API: DSL execution failed: %s", eval.Context.Error())
	}

	if api.Name != "dsl_spec" {
		t.Errorf("API: invalid name")
	}

	if api.Description != "Optional API description" {
		t.Errorf("API: invalid description")
	}

	if api.Version != "1.0" {
		t.Errorf("API: invalid version")
	}

	if api.Contact == nil {
		t.Errorf("API: contact is nil")
	}
	if api.Contact != nil && api.Contact.Name != "contact name" {
		t.Errorf("API: invalid contact name")
	}
	if api.Contact != nil && api.Contact.Email != "contact@goa.design" {
		t.Errorf("API: invalid contact email")
	}
	if api.Contact != nil && api.Contact.URL != "https://goa.design" {
		t.Errorf("API: invalid contact URL")
	}

	if api.License == nil {
		t.Errorf("API: license is nil")
	}
	if api.License != nil && api.License.Name != "License name" {
		t.Errorf("API: invalid license name")
	}
	if api.License != nil && api.License.URL != "https://goa.design/license" {
		t.Errorf("API: invalid license URL")
	}

	if api.Docs == nil {
		t.Errorf("API: docs is nil")
	}
	if api.Docs != nil && api.Docs.Description != "Optional description" {
		t.Errorf("API: invalid docs description")
	}
	if api.Docs != nil && api.Docs.URL != "https://goa.design/getting-started" {
		t.Errorf("API: invalid docs URL")
	}

	if len(api.Servers) == 0 {
		t.Errorf("API: missing servers")
	}
	if len(api.Servers) > 1 {
		t.Errorf("API: too many servers")
	}
	if len(api.Servers) == 1 {
		s := api.Servers[0]
		if s.URL != "https://{param}.goa.design:443/basePath" {
			t.Errorf("API: invalid server URL")
		}
		if s.Description != "Optional description" {
			t.Errorf("API: invalid server description")
		}
		if s.Params == nil {
			t.Errorf("API: missing server params")
		}
		if s.Params != nil {
			obj := s.Params.Type.(design.Object)
			if obj == nil {
				t.Errorf("API: invalid server params type")
			}
			if len(obj) == 0 {
				t.Errorf("API: server params object is empty")
			}
			if len(obj) > 1 {
				t.Errorf("API: server params object has too many attributes")
			}
			p, ok := obj["param"]
			if !ok {
				t.Errorf("API: missing env server param")
			}
			if ok && p.Type != design.String {
				t.Errorf("API: invalid env server param type")
			}
			if ok && p.Description != "Optional description" {
				t.Errorf("API: invalid server param description")
			}
			if ok && p.DefaultValue != "default" {
				t.Errorf("API: invalid server param default value")
			}
			if ok && p.Validation.Values == nil {
				t.Errorf("API: missing server param enum values")
			}
			if ok && len(p.Validation.Values) != 2 {
				t.Errorf("API: invalid server param enum value count")
			}
			if ok && len(p.Validation.Values) == 2 && p.Validation.Values[0] != "default" {
				t.Errorf("API: invalid server param enum first value")
			}
			if ok && len(p.Validation.Values) == 2 && p.Validation.Values[1] != "other" {
				t.Errorf("API: invalid server param enum second value")
			}
		}
	}

	if api.Metadata == nil {
		t.Errorf("API: nil metadata")
	}
	if api.Metadata != nil && len(api.Metadata) == 0 {
		t.Errorf("API: empty metadata")
	}
	if len(api.Metadata) > 1 {
		t.Errorf("API: too many metadata entries")
	}
	m, ok := api.Metadata["metadata"]
	if !ok {
		t.Errorf("API: missing metadata")
	}
	if ok && len(m) != 2 {
		t.Errorf("API: invalid metadata count")
	}
	if len(m) == 2 && m[0] != "value" {
		t.Errorf("API: invalid first metadata value")
	}
	if len(m) == 2 && m[1] != "other value" {
		t.Errorf("API: invalid second metadata value")
	}
}

func TestServiceSpec(t *testing.T) {
	if len(design.Root.Services) != 1 {
		t.Fatalf("Service: invalid API services count")
	}
	service := design.Root.Services[0]
	if !eval.Execute(service.DSL(), service) {
		t.Errorf("Service: DSL execution failed: %s", eval.Context.Error())
	}

	if service.Description != "Optional service description" {
		t.Errorf("Service: invalid description")
	}

	if service.Docs == nil {
		t.Errorf("Service: docs is nil")
	}
	if service.Docs != nil && service.Docs.Description != "Optional description" {
		t.Errorf("Service: invalid docs description")
	}
	if service.Docs != nil && service.Docs.URL != "https://goa.design" {
		t.Errorf("Service: invalid docs URL")
	}

	if service.DefaultTypeName != ServiceDefaultType.Name() {
		t.Errorf("Service: invalid default type name")
	}

	if len(service.Errors) != 5 {
		t.Fatalf("Service: invalid Errors count")
	}
	if service.Errors[0].Name != "name_of_error" {
		t.Errorf("Service: invalid first error name")
	}
	if service.Errors[0].Type != design.ErrorMedia {
		t.Errorf("Service: invalid first error type")
	}
	if service.Errors[0].Description != "" {
		t.Errorf("Service: invalid first error description")
	}
	if service.Errors[1].Name != "name_of_error_2" {
		t.Errorf("Service: invalid second error name")
	}
	if service.Errors[1].Type != design.ErrorMedia {
		t.Errorf("Service: invalid second error type")
	}
	if service.Errors[1].Description != "Optional description of error" {
		t.Errorf("Service: invalid second error description")
	}
	if service.Errors[2].Name != "name_of_error_3" {
		t.Errorf("Service: invalid third error name")
	}
	if service.Errors[2].Type != AErrorMediaType {
		t.Errorf("Service: invalid third error type")
	}
	if service.Errors[2].Description != "" {
		t.Errorf("Service: invalid third error description")
	}
	if service.Errors[3].Name != "name_of_error_4" {
		t.Errorf("Service: invalid fourth error name")
	}
	if service.Errors[3].Type != AErrorType {
		t.Errorf("Service: invalid fourth error type")
	}
	if service.Errors[3].Description != "" {
		t.Errorf("Service: invalid fourth error description")
	}
	if service.Errors[4].Name != "name_of_error_5" {
		t.Errorf("Service: invalid fifth error name")
	}
	if service.Errors[4].Description != "Optional description" {
		t.Errorf("Service: invalid fifth error description")
	}
	if len(service.Errors[4].Validation.Required) != 1 {
		t.Errorf("Service: invalid fifth error validation")
	}
	obj, ok := service.Errors[4].Type.(design.Object)
	if !ok {
		t.Errorf("Service: invalid fifth error type")
	}
	if ok && len(obj) != 1 {
		t.Errorf("Service: invalid fifth error type attribute count")
	}
	if len(obj) == 1 && obj["message"] == nil {
		t.Errorf("Service: invalid fifth error type attribute name")
	}
	if len(obj) == 1 && obj["message"] != nil && obj["message"].Type != design.String {
		t.Errorf("Service: invalid fifth error type attribute type")
	}

	if len(service.Endpoints) != 6 {
		t.Fatalf("Service: invalid endpoints count")
	}
	if service.Endpoints[0].Name != "endpoint" {
		t.Errorf("Service: invalid first endpoint name")
	}
	if service.Endpoints[0].Description != "Optional description" {
		t.Errorf("Service: invalid first endpoint description")
	}
	if service.Endpoints[0].Docs == nil {
		t.Errorf("Service: docs is nil")
	}
	if service.Endpoints[0].Docs != nil && service.Endpoints[0].Docs.Description != "Optional description" {
		t.Errorf("Service: invalid docs description")
	}
	if service.Endpoints[0].Docs != nil && service.Endpoints[0].Docs.URL != "https://goa.design" {
		t.Errorf("Service: invalid docs URL")
	}

	if service.DefaultTypeName != ServiceDefaultType.Name() {
		t.Errorf("Service: invalid default type name")
	}

	if service.Endpoints[1].Name != "default-type" {
		t.Errorf("Service: invalid second endpoint name")
	}
	if service.Endpoints[2].Name != "inline-primitive" {
		t.Errorf("Service: invalid third endpoint name")
	}
	if service.Endpoints[3].Name != "inline-array" {
		t.Errorf("Service: invalid fourth endpoint name")
	}
	if service.Endpoints[4].Name != "inline-map" {
		t.Errorf("Service: invalid fifth endpoint name")
	}
	if service.Endpoints[5].Name != "inline-object" {
		t.Errorf("Service: invalid sixth name")
	}
}
