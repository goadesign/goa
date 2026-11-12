// This file verifies that every supported validation format renders the
// matching goa runtime format constant in generated validation code.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestValidationCodeFormatConstant(t *testing.T) {
	cases := []struct {
		format   expr.ValidationFormat
		constant string
	}{
		{expr.FormatDate, "goa.FormatDate"},
		{expr.FormatDateTime, "goa.FormatDateTime"},
		{expr.FormatUUID, "goa.FormatUUID"},
		{expr.FormatEmail, "goa.FormatEmail"},
		{expr.FormatHostname, "goa.FormatHostname"},
		{expr.FormatIPv4, "goa.FormatIPv4"},
		{expr.FormatIPv6, "goa.FormatIPv6"},
		{expr.FormatIP, "goa.FormatIP"},
		{expr.FormatURI, "goa.FormatURI"},
		{expr.FormatURIReference, "goa.FormatURIReference"},
		{expr.FormatMAC, "goa.FormatMAC"},
		{expr.FormatCIDR, "goa.FormatCIDR"},
		{expr.FormatRegexp, "goa.FormatRegexp"},
		{expr.FormatJSON, "goa.FormatJSON"},
		{expr.FormatRFC1123, "goa.FormatRFC1123"},
	}
	for _, tc := range cases {
		t.Run(string(tc.format), func(t *testing.T) {
			att := &expr.AttributeExpr{
				Type:       expr.String,
				Validation: &expr.ValidationExpr{Format: tc.format},
			}
			ctx := NewAttributeContext(false, false, false, "", NewNameScope())
			code := ValidationCode(att, nil, ctx, true, false, false, "target")
			require.Contains(t, code, `goa.ValidateFormat("target", target, `+tc.constant+")")
		})
	}
}
