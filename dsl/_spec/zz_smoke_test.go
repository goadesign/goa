// This file verifies that the specs produce the expected data structures.
package dsl_test

import (
	"testing"

	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

func TestAPISpec(t *testing.T) {
	if err := eval.Register(design.Root); err != nil {
		t.Fatalf("API: failed to register DSL: %s", err)
	}
	if err := eval.RunDSL(); err != nil {
		t.Fatalf("API: failed to run DSL: %s", err)
	}
	api := design.Root.API

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
		if s.URL != "https://{param}.goa.design:443" {
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

	if len(service.Errors) != 5 {
		t.Fatalf("Service: invalid Errors count (%d)", len(service.Errors))
	}
	if service.Errors[0].Name != "name_of_error_1" {
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

	if len(service.Methods) != 2 {
		t.Fatalf("Service: invalid methods count")
	}
	if service.Methods[0].Name != "method" {
		t.Errorf("Service: invalid first method name")
	}
	if service.Methods[0].Description != "Optional description" {
		t.Errorf("Service: invalid first method description")
	}
	if service.Methods[0].Docs == nil {
		t.Errorf("Service: docs is nil")
	}
	if service.Methods[0].Docs != nil && service.Methods[0].Docs.Description != "Optional description" {
		t.Errorf("Service: invalid docs description")
	}
	if service.Methods[0].Docs != nil && service.Methods[0].Docs.URL != "https://goa.design" {
		t.Errorf("Service: invalid docs URL")
	}
	if service.Methods[0].Payload == nil {
		t.Errorf("Service: first method payload type is nil")
	}
	if service.Methods[0].Payload != nil && service.Methods[0].Payload.Kind() != design.UserTypeKind {
		t.Errorf("Service: first method payload type has invalid kind")
	}
	if ut, ok := service.Methods[0].Payload.(*design.UserTypeExpr); ok {
		if ut.Name() != "Payload" {
			t.Errorf("Service: invalid first method payload type")
		}
		if ut.Description != "Optional description" {
			t.Errorf("Service: invalid first method payload type description")
		}
		if o, ok := ut.Type.(design.Object); ok {
			if len(o) != 2 {
				t.Errorf("Service: invalid attribute count for first method payload type")
			} else {
				if att, ok := o["required"]; ok {
					if att.Type.Kind() != design.StringKind {
						t.Errorf("Service: invalid 'required' attribute type for first method payload type")
					}
				} else {
					t.Errorf("Service: missing 'required' attribute for first method payload type")
				}
				if att, ok := o["name"]; ok {
					if att.Type.Kind() != design.StringKind {
						t.Errorf("Service: invalid 'name' attribute type for first method payload type")
					}
				} else {
					t.Errorf("Service: missing 'name' attribute for first method payload type")
				}
			}
		}
		if len(ut.Validation.Required) != 2 {
			t.Errorf("Service: invalid first method payload type required attributes")
		} else {
			if ut.Validation.Required[0] != "required" {
				t.Errorf("Service: invalid first method payload type first required attribute name")
			}
			if ut.Validation.Required[1] != "name" {
				t.Errorf("Service: invalid first method payload type second required attribute name")
			}
		}
	}
	if service.Methods[0].Result == nil {
		t.Errorf("Service: first method payload type is nil")
	}
	if service.Methods[0].Result != nil && service.Methods[0].Result.Kind() != design.MediaTypeKind {
		t.Errorf("Service: first method result type has invalid kind")
	}
	if mt, ok := service.Methods[0].Result.(*design.MediaTypeExpr); ok {
		if mt.Name() != "application/vnd.goa.result" {
			t.Errorf("Service: invalid first method result media type identifier")
		}
		if mt.Description != "Optional description" {
			t.Errorf("Service: invalid first method result type description")
		}
		if o, ok := mt.Type.(design.Object); ok {
			if len(o) != 2 {
				t.Errorf("Service: invalid attribute count for first method result type")
			} else {
				if att, ok := o["required"]; ok {
					if att.Type.Kind() != design.StringKind {
						t.Errorf("Service: invalid 'required' attribute type for first method result type")
					}
				} else {
					t.Errorf("Service: missing 'required' attribute for first method result type")
				}
				if att, ok := o["name"]; ok {
					if att.Type.Kind() != design.StringKind {
						t.Errorf("Service: invalid 'name' attribute type for first method result type")
					}
				} else {
					t.Errorf("Service: missing 'name' attribute for first method result type")
				}
			}
		}
		if len(mt.Validation.Required) != 2 {
			t.Errorf("Service: invalid first method result type required attributes")
		} else {
			if mt.Validation.Required[0] != "required" {
				t.Errorf("Service: invalid first method result type first required attribute name")
			}
			if mt.Validation.Required[1] != "name" {
				t.Errorf("Service: invalid first method result type second required attribute name")
			}
		}
		if len(mt.Views) != 1 {
			t.Errorf("Service: invalid first method result media type view count")
		}
		if len(mt.Views) == 1 && mt.Views[0].Name != "default" {
			t.Errorf("Service: invalid first method result media type view name")
		}
		if len(mt.Views) == 1 {
			o := mt.Views[0].AttributeExpr.Type.(design.Object)
			if len(o) != 2 {
				t.Errorf("Service: invalid first method result media type view attribute count")
			}
			if len(o) == 2 && o["required"] == nil {
				t.Errorf("Service: missing first method result media type view attribute 'required' attribute")
			}
			if len(o) == 2 && o["name"] == nil {
				t.Errorf("Service: missing first method result media type view attribute 'name' attribute")
			}
		}
	}
	if len(service.Methods[0].Errors) != 1 {
		t.Errorf("Service: invalid first method error count")
	}
	if len(service.Methods[0].Errors) == 1 && service.Methods[0].Errors[0].Name != "method_specific_error" {
		t.Errorf("Service: invalid first method error name")
	}
	if len(service.Methods[0].Errors) == 1 && service.Methods[0].Errors[0].Type != design.ErrorMedia {
		t.Errorf("Service: invalid first method error type")
	}
	if len(service.Methods[0].Metadata) != 1 {
		t.Errorf("Service: invalid first method metadata count")
	}
	if len(service.Methods[0].Metadata) == 1 {
		if _, ok := service.Methods[0].Metadata["name"]; !ok {
			t.Errorf("Service: first method metadata is missing 'name' key")
		} else {
			if len(service.Methods[0].Metadata["name"]) != 2 {
				t.Errorf("Service: first method metadata 'name' is invalid")
			} else {
				if service.Methods[0].Metadata["name"][0] != "some value" {
					t.Errorf("Service: first method metadata 'name' first value is invalid")
				}
				if service.Methods[0].Metadata["name"][1] != "some other value" {
					t.Errorf("Service: first method metadata 'name' second value is invalid")
				}
			}
		}

	}

	if service.Methods[1].Name != "inline-object" {
		t.Errorf("Service: invalid third name")
	}
	ut, ok := service.Methods[1].Payload.(*design.UserTypeExpr)
	if !ok {
		t.Errorf("Service: invalid third method payload type")
	} else {
		if ut.Description != "Optional description" {
			t.Errorf("Service: invalid third method payload type description")
		}
		o, ok := ut.Type.(design.Object)
		if !ok {
			t.Errorf("Service: invalid third method payload inner type")
		} else {
			at, ok := o["required"]
			if !ok {
				t.Errorf("Service: third method payload inner type is missing 'required' attribute")
			} else if at.Type != design.String {
				t.Errorf("Service: third method payload type 'required' field type is invalid")
			}
			at, ok = o["optional"]
			if !ok {
				t.Errorf("Service: third method payload inner type is missing 'optional' attribute")
			} else if at.Type != design.String {
				t.Errorf("Service: third method payload type 'optional' field type is invalid")
			}
		}
		if len(ut.Validation.Required) == 0 {
			t.Errorf("Service: third method payload type is missing required field")
		}
	}
	ut, ok = service.Methods[1].Result.(*design.UserTypeExpr)
	if !ok {
		t.Errorf("Service: invalid third method result type")
	} else {
		if ut.Description != "Optional description" {
			t.Errorf("Service: invalid third method result type description")
		}
		o, ok := ut.Type.(design.Object)
		if !ok {
			t.Errorf("Service: invalid third method result inner type")
		} else {
			at, ok := o["required"]
			if !ok {
				t.Errorf("Service: third method result inner type is missing 'required' attribute")
			} else if at.Type != design.String {
				t.Errorf("Service: third method result type 'required' field type is invalid")
			}
			at, ok = o["optional"]
			if !ok {
				t.Errorf("Service: third method result inner type is missing 'optional' attribute")
			} else if at.Type != design.String {
				t.Errorf("Service: third method result type 'optional' field type is invalid")
			}
		}
		if len(ut.Validation.Required) == 0 {
			t.Errorf("Service: third method result type is missing required field")
		}
	}
}

func TestTypes(t *testing.T) {
	if len(design.Root.Types) != 10 {
		t.Fatalf("Types: invalid count (%d)", len(design.Root.Types))
	}
	b := design.Root.UserType("Name")
	if b == nil {
		t.Fatalf("Types: type 'Name' is missing")
	}
	basic, ok := b.(*design.UserTypeExpr)
	if !ok {
		t.Fatalf("Types: invalid 'Name' type")
	}
	if basic.Description != "Optional description" {
		t.Errorf("Types: invalid 'Name' type description")
	}
	if basic.Type.Kind() != design.ObjectKind {
		t.Fatalf("Types: invalid 'Name' type kind")
	}
	o := basic.Type.(design.Object)
	if len(o) != 1 {
		t.Fatalf("Types: invalid 'Name' type attribute count")
	}
	att, ok := o["an_attribute"]
	if !ok {
		t.Fatalf("Types: missing 'Name' type 'an_attribute' attribute")
	}
	if att.Type.Kind() != design.StringKind {
		t.Errorf("Types: invalid 'Name' type 'an_attribute' attribute type")
	}
	if len(basic.Validation.Required) != 1 {
		t.Fatalf("Types: invalid 'Name' type required attribute count")
	}
	if basic.Validation.Required[0] != "an_attribute" {
		t.Errorf("Types: invalid 'Name' type first required attribute")
	}

	a := design.Root.UserType("AllTypes")
	if a == nil {
		t.Fatalf("Types: type 'AllTypes' is missing")
	}
	all, ok := a.(*design.UserTypeExpr)
	if !ok {
		t.Fatalf("Types: invalid 'AllTypes' type")
	}
	if all.Description != "An object with attributes of all possible types" {
		t.Errorf("Types: invalid 'AllTypes' type description")
	}
	if all.Type.Kind() != design.ObjectKind {
		t.Fatalf("Types: invalid 'AllTypes' type kind")
	}
	o = all.Type.(design.Object)
	if len(o) != 14 {
		t.Fatalf("Types: invalid 'AllTypes' type attribute count")
	}
	cases := map[string]design.Kind{
		"string":     design.StringKind,
		"bytes":      design.BytesKind,
		"boolean":    design.BooleanKind,
		"int32":      design.Int32Kind,
		"int64":      design.Int64Kind,
		"float32":    design.Float32Kind,
		"float64":    design.Float64Kind,
		"any":        design.AnyKind,
		"object":     design.ObjectKind,
		"user":       design.UserTypeKind,
		"media":      design.MediaTypeKind,
		"collection": design.MediaTypeKind,
	}
	for n, k := range cases {
		att, ok := o[n]
		if !ok {
			t.Fatalf("Types: invalid 'AllTypes' type '%s' attribute type", n)
		}
		if att.Type.Kind() != k {
			t.Errorf("Types: invalid 'AllTypes' type '%s' attribute type", n)
		}
	}
	if att, ok := o["object"]; ok {
		if att.Description != "Inner type" {
			t.Errorf("Types: invalid 'AllTypes' type 'object' attribute description")
		}
		if len(att.Validation.Required) != 1 {
			t.Fatalf("Types: invalid 'AllTypes' type 'object' attribute required validation")
		}
		if att.Validation.Required[0] != "inner_attribute" {
			t.Fatalf("Types: invalid 'AllTypes' type 'object' attribute required validation attribute name")
		}
		o = att.Type.(design.Object)
		if len(o) != 1 {
			t.Errorf("Types: invalid 'AllTypes' type 'object' attribute inner attribute count")
		}
		if _, ok := o["inner_attribute"]; !ok {
			t.Fatalf("Types: invalid 'AllTypes' type 'object' attribute inner attribute missing")
		}
		if o["inner_attribute"].Type.Kind() != design.StringKind {
			t.Errorf("Types: invalid 'AllTypes' type 'object' attribute inner attribute type")
		}
	}

	if AArrayType.ElemType.Type.Kind() != design.StringKind {
		t.Errorf("Types: invalid 'AArrayType' element type")
	}
	if AArrayType.ElemType.Validation.Pattern != "regexp" {
		t.Fatalf("Types: invalid 'AArrayType' element type validation")
	}

	if AMapType.ElemType.Type.Kind() != design.StringKind {
		t.Errorf("Types: invalid 'AMapType' element type")
	}
	if AMapType.KeyType.Type.Kind() != design.StringKind {
		t.Errorf("Types: invalid 'AMapType' key type")
	}
	if AMapType.ElemType.Validation.Pattern != "valueregexp" {
		t.Fatalf("Types: invalid 'AMapType' element type validation")
	}
	if AMapType.KeyType.Validation.Pattern != "keyregexp" {
		t.Fatalf("Types: invalid 'AMapType' key type validation")
	}

	attrs := design.Root.UserType("Attributes")
	if attrs == nil {
		t.Fatalf("Types: type 'Attrs' is missing")
	}
	if attrs.Attribute().Type.Kind() != design.ObjectKind {
		t.Fatalf("Types: type 'Attrs' invalid kind")
	}
	o = attrs.Attribute().Type.(design.Object)
	if len(o) != 4 {
		t.Fatalf("Types: type 'Attrs' invalid attribute count")
	}
	for _, n := range []string{"name", "name_2", "name_3", "name_4"} {
		if _, ok := o[n]; !ok {
			t.Fatalf("Types: type 'Attrs' missing %s attribute", n)
		}
		if o[n].Type.Kind() != design.StringKind {
			t.Errorf("Types: type 'Attrs' attribute %s invalid kind", n)
		}
	}
	if o["name_2"].Description != "description" {
		t.Errorf("Types: type 'Attrs' invalid 'name_2' attribute description")
	}
	if o["name_3"].Validation.MinLength == nil {
		t.Fatalf("Types: type 'Attrs' missing 'name_3' attribute validation")
	}
	if *o["name_3"].Validation.MinLength != 10 {
		t.Errorf("Types: type 'Attrs' invalid 'name_3' attribute min length validation")
	}
	if o["name_4"].Description != "description" {
		t.Errorf("Types: type 'Attrs' invalid 'name_4' attribute description")
	}
	if o["name_4"].Validation.MinLength == nil {
		t.Fatalf("Types: type 'Attrs' missing 'name_4' attribute min length validation")
	}
	if *o["name_4"].Validation.MinLength != 10 {
		t.Errorf("Types: type 'Attrs' invalid 'name_4' attribute min length validation")
	}
	if o["name_4"].Validation.MaxLength == nil {
		t.Fatalf("Types: type 'Attrs' missing 'name_4' attribute max length validation")
	}
	if *o["name_4"].Validation.MaxLength != 100 {
		t.Errorf("Types: type 'Attrs' invalid 'name_4' attribute max length validation")
	}
	if o["name_4"].DefaultValue == nil {
		t.Errorf("Types: type 'Attrs' missing 'name_4' attribute default value")
	}
	if o["name_4"].DefaultValue != "default value" {
		t.Errorf("Types: type 'Attrs' invalid 'name_4' attribute default value")
	}
	if len(o["name_4"].UserExamples) == 0 {
		t.Errorf("Types: type 'Attrs' missing 'name_4' attribute example value")
	}
	if len(o["name_4"].UserExamples) != 1 || o["name_4"].UserExamples[0].Value != "example value" {
		t.Errorf("Types: type 'Attrs' invalid 'name_4' attribute example value")
	}

	rec := design.Root.UserType("Recursive")
	if rec == nil {
		t.Fatalf("Types: missing 'Recursive' type")
	}
	if rec.Attribute().Type.Kind() != design.ObjectKind {
		t.Fatalf("Types: invalid 'Recursive' type")
	}
	o = rec.Attribute().Type.(design.Object)
	if len(o) != 2 {
		t.Fatalf("Types: invalid 'Recursive' type attribute count")
	}
	if _, ok := o["recursive"]; !ok {
		t.Fatalf("Types: missing 'Recursive' type attribute 'recursive'")
	}
	if o["recursive"].Type.Kind() != design.UserTypeKind {
		t.Fatalf("Types: invalid 'Recursive' type attribute 'recursive' kind")
	}
	if o["recursive"].Type.(*design.UserTypeExpr).TypeName != "Recursive" {
		t.Errorf("Types: invalid 'Recursive' type attribute 'recursive' type")
	}
	if _, ok := o["recursives"]; !ok {
		t.Fatalf("Types: missing 'Recursive' type attribute 'recursives'")
	}
	if o["recursives"].Type.Kind() != design.ArrayKind {
		t.Fatalf("Types: invalid 'Recursive' type attribute 'recursives' kind")
	}
	if o["recursives"].Type.(*design.Array).ElemType.Type.Kind() != design.UserTypeKind {
		t.Errorf("Types: invalid 'Recursive' type attribute 'recursives' array element type")
	}
	if o["recursives"].Type.(*design.Array).ElemType.Type.(*design.UserTypeExpr).TypeName != "Recursive" {
		t.Errorf("Types: invalid 'Recursive' type attribute 'recursives' array element type name")
	}
}
