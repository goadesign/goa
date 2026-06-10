package openapi

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestSpecs(t *testing.T) {
	cases := []struct {
		name  string
		meta  expr.MetaExpr
		specs []Spec
		err   string
	}{{
		name: "default",
		specs: []Spec{
			{Version: Version20, Path: "http/openapi"},
			{Version: Version30, Path: "http/openapi3"},
			{Version: Version32, Path: "http/openapi3.2"},
		},
	}, {
		name: "subset",
		meta: expr.MetaExpr{"openapi:versions": {"3.2"}},
		specs: []Spec{
			{Version: Version32, Path: "http/openapi3.2"},
		},
	}, {
		name: "canonical order",
		meta: expr.MetaExpr{"openapi:versions": {"3.2", "2.0"}},
		specs: []Spec{
			{Version: Version20, Path: "http/openapi"},
			{Version: Version32, Path: "http/openapi3.2"},
		},
	}, {
		name: "duplicate values",
		meta: expr.MetaExpr{"openapi:versions": {"3.0", "3.0"}},
		specs: []Spec{
			{Version: Version30, Path: "http/openapi3"},
		},
	}, {
		name: "unknown version",
		meta: expr.MetaExpr{"openapi:versions": {"3.1"}},
		err:  `invalid value "3.1" for meta "openapi:versions": valid values are "2.0", "3.0" and "3.2"`,
	}, {
		name: "path override",
		meta: expr.MetaExpr{"openapi:path:3.2": {"docs/openapi"}},
		specs: []Spec{
			{Version: Version20, Path: "http/openapi"},
			{Version: Version30, Path: "http/openapi3"},
			{Version: Version32, Path: "docs/openapi"},
		},
	}, {
		name: "path override of unselected version is ignored",
		meta: expr.MetaExpr{
			"openapi:versions": {"3.0"},
			"openapi:path:3.2": {"docs/openapi"},
		},
		specs: []Spec{
			{Version: Version30, Path: "http/openapi3"},
		},
	}, {
		name: "empty path",
		meta: expr.MetaExpr{"openapi:path:3.2": {""}},
		err:  `invalid value for meta "openapi:path:3.2": path cannot be empty`,
	}, {
		name: "absolute path",
		meta: expr.MetaExpr{"openapi:path:3.2": {"/etc/openapi"}},
		err:  `invalid value "/etc/openapi" for meta "openapi:path:3.2": path must be relative to the gen directory`,
	}, {
		name: "escaping path",
		meta: expr.MetaExpr{"openapi:path:3.2": {"../openapi"}},
		err:  `invalid value "../openapi" for meta "openapi:path:3.2": path must not escape the gen directory`,
	}, {
		name: "nested escaping path",
		meta: expr.MetaExpr{"openapi:path:3.2": {"docs/../../openapi"}},
		err:  `invalid value "docs/../../openapi" for meta "openapi:path:3.2": path must not escape the gen directory`,
	}, {
		name: "path with extension",
		meta: expr.MetaExpr{"openapi:path:3.2": {"docs/openapi.json"}},
		err:  `invalid value "docs/openapi.json" for meta "openapi:path:3.2": path must not include an extension, both docs/openapi.json.json and docs/openapi.json.yaml are generated`,
	}, {
		name: "duplicate paths",
		meta: expr.MetaExpr{"openapi:path:3.2": {"http/openapi3"}},
		err:  `OpenAPI versions 3.0 and 3.2 use the same output path "http/openapi3": override one with meta "openapi:path:3.2"`,
	}, {
		name: "unknown path key version",
		meta: expr.MetaExpr{"openapi:path:3.0.3": {"docs/openapi"}},
		err:  `invalid meta key "openapi:path:3.0.3": valid OpenAPI versions are "2.0", "3.0" and "3.2"`,
	}}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			specs, err := Specs(c.meta)
			if c.err != "" {
				require.EqualError(t, err, c.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.specs, specs)
		})
	}
}
