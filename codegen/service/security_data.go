// This file reads evaluated security schemes and their credential fields for
// service and transport generators. Authentication calls use the same authored
// Go field names, string types and presence rules as the generated payload.
package service

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// BuildSchemeData reads one method's annotated credentials and security scheme.
// Service and transport generators receive the payload's actual Go field names,
// requiredness and scopes, including names changed by field metadata.
func BuildSchemeData(s *expr.SchemeExpr, m *expr.MethodExpr) *SchemeData {
	if !expr.IsObject(m.Payload.Type) {
		return nil
	}
	if s.Kind == expr.BasicAuthKind {
		userAtt := expr.TaggedAttribute(m.Payload, "security:username")
		passAtt := expr.TaggedAttribute(m.Payload, "security:password")
		return &SchemeData{
			Type:             s.Kind.String(),
			SchemeName:       s.SchemeName,
			UsernameAttr:     userAtt,
			UsernameField:    codegen.GoifyAtt(m.Payload.Find(userAtt), userAtt, true),
			UsernamePointer:  m.Payload.IsPrimitivePointer(userAtt, true),
			UsernameRequired: m.Payload.IsRequired(userAtt),
			PasswordAttr:     passAtt,
			PasswordField:    codegen.GoifyAtt(m.Payload.Find(passAtt), passAtt, true),
			PasswordPointer:  m.Payload.IsPrimitivePointer(passAtt, true),
			PasswordRequired: m.Payload.IsRequired(passAtt),
			Scopes:           schemeScopes(s),
		}
	}
	// The remaining scheme kinds all carry a single credential attribute
	// identified by a kind-specific security tag on the method payload.
	var tag string
	switch s.Kind {
	case expr.APIKeyKind:
		tag = "security:apikey:" + s.SchemeName
	case expr.BearerKind:
		tag = "security:bearer"
	case expr.JWTKind:
		tag = "security:token"
	case expr.OAuth2Kind:
		tag = "security:accesstoken"
	default:
		return nil
	}
	keyAtt := expr.TaggedAttribute(m.Payload, tag)
	if keyAtt == "" {
		return nil
	}
	data := &SchemeData{
		Type:         s.Kind.String(),
		Name:         s.Name,
		SchemeName:   s.SchemeName,
		CredField:    codegen.GoifyAtt(m.Payload.Find(keyAtt), keyAtt, true),
		CredPointer:  m.Payload.IsPrimitivePointer(keyAtt, true),
		CredRequired: m.Payload.IsRequired(keyAtt),
		KeyAttr:      keyAtt,
		Scopes:       schemeScopes(s),
		In:           s.In,
	}
	if s.Kind == expr.OAuth2Kind {
		data.Flows = s.Flows
	}
	return data
}

// schemeScopes returns the authorization scope names defined by the scheme. It
// returns nil when the scheme defines none.
func schemeScopes(s *expr.SchemeExpr) []string {
	if len(s.Scopes) == 0 {
		return nil
	}
	scopes := make([]string, len(s.Scopes))
	for i, sc := range s.Scopes {
		scopes[i] = sc.Name
	}
	return scopes
}
