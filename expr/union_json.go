// This file checks JSON mappings for Goa OneOf declarations. An object whose
// fields sit beside a discriminator must leave that name to the union, in both
// service JSON and HTTP bodies; invalid designs stop before code generation.
package expr

import (
	"strings"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/internal/codegenname"
)

// validateUnionJSON rejects flattened branches that cannot be encoded as
// objects or would overwrite the union's discriminator field.
func validateUnionJSON(union *Union, parent eval.Expression) *eval.ValidationErrors {
	if !union.Flatten {
		return nil
	}
	errors := new(eval.ValidationErrors)
	if union.ValueKey != "" {
		errors.Add(parent, "flattened OneOf %q cannot select a value field", union.TypeName)
	}
	for _, branch := range union.Values {
		object := AsObject(branch.Attribute.Type)
		if object == nil {
			errors.Add(parent, "flattened OneOf %q branch %q must be an object", union.TypeName, branch.Name)
			continue
		}
		mapped := NewMappedAttributeExpr(branch.Attribute)
		for _, field := range *AsObject(mapped.Type) {
			// HTTP mappings and service JSON tags can use different names. Neither
			// representation may take the field that the union writes itself.
			goName := codegenname.Goify(codegenname.AttributeName(field.Name, field.Attribute.Meta["struct:field:name"]), true)
			names := []string{mapped.ElemName(field.Name), goName}
			if tag, ok := field.Attribute.Meta["struct:tag:json"]; ok {
				name := strings.SplitN(strings.Join(tag, ","), ",", 2)[0]
				if name == "-" {
					continue
				}
				if name != "" {
					names = []string{name}
				}
			} else if tag, ok := field.Attribute.Meta["struct:tag:json:name"]; ok && len(tag) > 0 {
				names = []string{mapped.ElemName(field.Name), strings.Join(tag, ",")}
			}
			for _, name := range names {
				if name == union.GetTypeKey() {
					errors.Add(parent, "flattened OneOf %q branch %q field %q uses discriminator JSON name %q", union.TypeName, branch.Name, field.Name, name)
					break
				}
			}
		}
	}
	return errors
}
