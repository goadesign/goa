// This file formats evaluated security schemes and authorization attributes for service templates.
package service

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// BuildSchemeData builds the scheme data for the given scheme and method expr.
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
			UsernameField:    codegen.Goify(userAtt, true),
			UsernamePointer:  m.Payload.IsPrimitivePointer(userAtt, true),
			UsernameRequired: m.Payload.IsRequired(userAtt),
			PasswordAttr:     passAtt,
			PasswordField:    codegen.Goify(passAtt, true),
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
		CredField:    codegen.Goify(keyAtt, true),
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
