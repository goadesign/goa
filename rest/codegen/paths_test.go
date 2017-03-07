package codegen

import (
	"bytes"
	"testing"

	goadesign "goa.design/goa.v2/design"
	"goa.design/goa.v2/rest/design"
)

func TestPaths(t *testing.T) {
	const (
		pathWithoutParams = `
// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath() string {
	return "/account/test"
}

`

		pathWithOneParam = `
// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(id int32) string {
	return fmt.Sprintf("/account/test/%v", id)
}

`
		pathWithMultipleParams = `
// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(id int32, view string) string {
	return fmt.Sprintf("/account/test/%v/view/%v", id, view)
}

`

		pathWithAlternatives = `
// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath() string {
	return "/account/test"
}

// ShowAccountPath2 returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath2(id int32) string {
	return fmt.Sprintf("/account/test/%v", id)
}

// ShowAccountPath3 returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath3(id int32, view string) string {
	return fmt.Sprintf("/account/test/%v/view/%v", id, view)
}

`

		pathWithStringSliceParam = `
// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(slice_string []string) string {
	encodedslice_string := make([]string, len(slice_string))
	for i, v := range slice_string {
		encodedslice_string[i] = url.QueryEscape(v)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedslice_string, ","))
}

`

		pathWithInt32SliceParam = `
// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(slice_int32 []int32) string {
	encodedslice_int32 := make([]string, len(slice_int32))
	for i, v := range slice_int32 {
		encodedslice_int32[i] = strconv.FormatInt(v, 10)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedslice_int32, ","))
}

`

		pathWithInt64SliceParam = `
// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(slice_int64 []int64) string {
	encodedslice_int64 := make([]string, len(slice_int64))
	for i, v := range slice_int64 {
		encodedslice_int64[i] = strconv.FormatInt(v, 10)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedslice_int64, ","))
}

`

		pathWithUint32SliceParam = `
// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(slice_uint32 []uint32) string {
	encodedslice_uint32 := make([]string, len(slice_uint32))
	for i, v := range slice_uint32 {
		encodedslice_uint32[i] = strconv.FormatUint(v, 10)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedslice_uint32, ","))
}

`

		pathWithUint64SliceParam = `
// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(slice_uint64 []uint64) string {
	encodedslice_uint64 := make([]string, len(slice_uint64))
	for i, v := range slice_uint64 {
		encodedslice_uint64[i] = strconv.FormatUint(v, 10)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedslice_uint64, ","))
}

`

		pathWithFloat32SliceParam = `
// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(slice_float32 []float32) string {
	encodedslice_float32 := make([]string, len(slice_float32))
	for i, v := range slice_float32 {
		encodedslice_float32[i] = strconv.FormatFloat(v, 'f', -1, 32)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedslice_float32, ","))
}

`

		pathWithFloat64SliceParam = `
// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(slice_float64 []float64) string {
	encodedslice_float64 := make([]string, len(slice_float64))
	for i, v := range slice_float64 {
		encodedslice_float64[i] = strconv.FormatFloat(v, 'f', -1, 64)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedslice_float64, ","))
}

`

		pathWithBoolSliceParam = `
// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(slice_bool []bool) string {
	encodedslice_bool := make([]string, len(slice_bool))
	for i, v := range slice_bool {
		encodedslice_bool[i] = strconv.FormatBool(v)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedslice_bool, ","))
}

`

		pathWithInterfaceSliceParam = `
// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(slice_interface []interface{}) string {
	encodedslice_interface := make([]string, len(slice_interface))
	for i, v := range slice_interface {
		encodedslice_interface[i] = url.QueryEscape(fmt.Sprintf("%v", v))
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedslice_interface, ","))
}

`
	)
	var (
		service = goadesign.ServiceExpr{
			Name: "Account",
		}

		endpoint = goadesign.EndpointExpr{
			Name:    "Show",
			Service: &service,
		}

		resource = design.ResourceExpr{
			Path: "/account",
		}

		setParams = func(a *goadesign.AttributeExpr) {
			a.Type = goadesign.Object{
				"id":              {Type: goadesign.Int32},
				"view":            {Type: goadesign.String},
				"slice_string":    {Type: &goadesign.Array{ElemType: &goadesign.AttributeExpr{Type: goadesign.String}}},
				"slice_int32":     {Type: &goadesign.Array{ElemType: &goadesign.AttributeExpr{Type: goadesign.Int32}}},
				"slice_int64":     {Type: &goadesign.Array{ElemType: &goadesign.AttributeExpr{Type: goadesign.Int64}}},
				"slice_uint32":    {Type: &goadesign.Array{ElemType: &goadesign.AttributeExpr{Type: goadesign.UInt32}}},
				"slice_uint64":    {Type: &goadesign.Array{ElemType: &goadesign.AttributeExpr{Type: goadesign.UInt64}}},
				"slice_float32":   {Type: &goadesign.Array{ElemType: &goadesign.AttributeExpr{Type: goadesign.Float32}}},
				"slice_float64":   {Type: &goadesign.Array{ElemType: &goadesign.AttributeExpr{Type: goadesign.Float64}}},
				"slice_bool":      {Type: &goadesign.Array{ElemType: &goadesign.AttributeExpr{Type: goadesign.Boolean}}},
				"slice_interface": {Type: &goadesign.Array{ElemType: &goadesign.AttributeExpr{Type: goadesign.Any}}},
			}
		}

		action = func(paths ...string) *design.ActionExpr {
			routes := make([]*design.RouteExpr, len(paths))
			for i, path := range paths {
				routes[i] = &design.RouteExpr{Path: path}
			}

			a := &design.ActionExpr{
				EndpointExpr: &endpoint,
				Resource:     &resource,
				Routes:       routes,
			}

			for _, r := range a.Routes {
				r.Action = a
			}
			setParams(a.Params())

			return a
		}
	)

	cases := map[string]struct {
		Action   *design.ActionExpr
		Expected string
	}{
		"single-path-no-param":            {Action: action("/test"), Expected: pathWithoutParams},
		"single-path-one-param":           {Action: action("/test/:id"), Expected: pathWithOneParam},
		"single-path-multiple-params":     {Action: action("/test/:id/view/:view"), Expected: pathWithMultipleParams},
		"alternative-paths":               {Action: action("/test", "/test/:id", "/test/:id/view/:view"), Expected: pathWithAlternatives},
		"path-with-string-slice-param":    {Action: action("/test/:slice_string"), Expected: pathWithStringSliceParam},
		"path-with-int32-slice-param":     {Action: action("/test/:slice_int32"), Expected: pathWithInt32SliceParam},
		"path-with-int64-slice-param":     {Action: action("/test/:slice_int64"), Expected: pathWithInt64SliceParam},
		"path-with-uint32-slice-param":    {Action: action("/test/:slice_uint32"), Expected: pathWithUint32SliceParam},
		"path-with-uint64-slice-param":    {Action: action("/test/:slice_uint64"), Expected: pathWithUint64SliceParam},
		"path-with-float32-slice-param":   {Action: action("/test/:slice_float32"), Expected: pathWithFloat32SliceParam},
		"path-with-float64-slice-param":   {Action: action("/test/:slice_float64"), Expected: pathWithFloat64SliceParam},
		"path-with-bool-slice-param":      {Action: action("/test/:slice_bool"), Expected: pathWithBoolSliceParam},
		"path-with-interface-slice-param": {Action: action("/test/:slice_interface"), Expected: pathWithInterfaceSliceParam},
	}

	for k, tc := range cases {
		buf := new(bytes.Buffer)
		s := Path(tc.Action)
		e := s.Render(buf)
		actual := buf.String()

		if e != nil {
			t.Errorf("%s: failed to execute template, error %s", k, e)
		} else if actual != tc.Expected {
			t.Errorf("%s: got %v, expected %v", k, actual, tc.Expected)
		}
	}
}
