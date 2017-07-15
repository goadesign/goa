package rest

import (
	"bytes"
	"testing"

	"goa.design/goa.v2/codegen/service"
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/design/rest"
)

func TestPaths(t *testing.T) {
	const (
		pathWithoutParams = `// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath() string {
	return "/account/test"
}

`

		pathWithOneParam = `// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(id int32) string {
	return fmt.Sprintf("/account/test/%v", id)
}

`
		pathWithMultipleParams = `// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(id int32, view string) string {
	return fmt.Sprintf("/account/test/%v/view/%v", id, view)
}

`

		pathWithAlternatives = `// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
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

		pathWithStringSliceParam = `// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(sliceString []string) string {
	encodedSliceString := make([]string, len(sliceString))
	for i, v := range sliceString {
		encodedSliceString[i] = url.QueryEscape(v)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedSliceString, ","))
}

`

		pathWithIntSliceParam = `// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(sliceInt []int) string {
	encodedSliceInt := make([]string, len(sliceInt))
	for i, v := range sliceInt {
		encodedSliceInt[i] = strconv.FormatInt(int64(v), 10)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedSliceInt, ","))
}

`

		pathWithInt32SliceParam = `// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(sliceInt32 []int32) string {
	encodedSliceInt32 := make([]string, len(sliceInt32))
	for i, v := range sliceInt32 {
		encodedSliceInt32[i] = strconv.FormatInt(int64(v), 10)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedSliceInt32, ","))
}

`

		pathWithInt64SliceParam = `// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(sliceInt64 []int64) string {
	encodedSliceInt64 := make([]string, len(sliceInt64))
	for i, v := range sliceInt64 {
		encodedSliceInt64[i] = strconv.FormatInt(v, 10)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedSliceInt64, ","))
}

`

		pathWithUintSliceParam = `// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(sliceUint []uint) string {
	encodedSliceUint := make([]string, len(sliceUint))
	for i, v := range sliceUint {
		encodedSliceUint[i] = strconv.FormatUint(uint64(v), 10)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedSliceUint, ","))
}

`

		pathWithUint32SliceParam = `// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(sliceUint32 []uint32) string {
	encodedSliceUint32 := make([]string, len(sliceUint32))
	for i, v := range sliceUint32 {
		encodedSliceUint32[i] = strconv.FormatUint(uint64(v), 10)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedSliceUint32, ","))
}

`

		pathWithUint64SliceParam = `// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(sliceUint64 []uint64) string {
	encodedSliceUint64 := make([]string, len(sliceUint64))
	for i, v := range sliceUint64 {
		encodedSliceUint64[i] = strconv.FormatUint(v, 10)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedSliceUint64, ","))
}

`

		pathWithFloat32SliceParam = `// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(sliceFloat32 []float32) string {
	encodedSliceFloat32 := make([]string, len(sliceFloat32))
	for i, v := range sliceFloat32 {
		encodedSliceFloat32[i] = strconv.FormatFloat(float64(v), 'f', -1, 32)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedSliceFloat32, ","))
}

`

		pathWithFloat64SliceParam = `// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(sliceFloat64 []float64) string {
	encodedSliceFloat64 := make([]string, len(sliceFloat64))
	for i, v := range sliceFloat64 {
		encodedSliceFloat64[i] = strconv.FormatFloat(v, 'f', -1, 64)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedSliceFloat64, ","))
}

`

		pathWithBoolSliceParam = `// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(sliceBool []bool) string {
	encodedSliceBool := make([]string, len(sliceBool))
	for i, v := range sliceBool {
		encodedSliceBool[i] = strconv.FormatBool(v)
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedSliceBool, ","))
}

`

		pathWithInterfaceSliceParam = `// ShowAccountPath returns the URL path to the Account service Show HTTP endpoint.
func ShowAccountPath(sliceInterface []interface{}) string {
	encodedSliceInterface := make([]string, len(sliceInterface))
	for i, v := range sliceInterface {
		encodedSliceInterface[i] = url.QueryEscape(fmt.Sprintf("%v", v))
	}

	return fmt.Sprintf("/account/test/%v", strings.Join(encodedSliceInterface, ","))
}

`
	)
	var (
		svc = design.ServiceExpr{
			Name: "Account",
		}

		method = design.MethodExpr{
			Name:    "Show",
			Service: &svc,
			Payload: &design.AttributeExpr{Type: design.Empty},
		}

		httpSvc = rest.HTTPServiceExpr{
			ServiceExpr: &svc,
			Path:        "/account",
		}

		setParams = func(a *design.AttributeExpr) {
			a.Type = &design.Object{
				{"id", &design.AttributeExpr{Type: design.Int32}},
				{"view", &design.AttributeExpr{Type: design.String}},
				{"slice_string", &design.AttributeExpr{Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.String}}}},
				{"slice_int", &design.AttributeExpr{Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Int}}}},
				{"slice_int32", &design.AttributeExpr{Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Int32}}}},
				{"slice_int64", &design.AttributeExpr{Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Int64}}}},
				{"slice_uint", &design.AttributeExpr{Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.UInt}}}},
				{"slice_uint32", &design.AttributeExpr{Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.UInt32}}}},
				{"slice_uint64", &design.AttributeExpr{Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.UInt64}}}},
				{"slice_float32", &design.AttributeExpr{Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Float32}}}},
				{"slice_float64", &design.AttributeExpr{Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Float64}}}},
				{"slice_bool", &design.AttributeExpr{Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Boolean}}}},
				{"slice_interface", &design.AttributeExpr{Type: &design.Array{ElemType: &design.AttributeExpr{Type: design.Any}}}},
			}
		}

		endpoint = func(paths ...string) *rest.HTTPEndpointExpr {
			routes := make([]*rest.RouteExpr, len(paths))
			for i, path := range paths {
				routes[i] = &rest.RouteExpr{Path: path}
			}

			a := &rest.HTTPEndpointExpr{
				MethodExpr: &method,
				Service:    &httpSvc,
				Routes:     routes,
			}

			for _, r := range a.Routes {
				r.Endpoint = a
			}
			setParams(a.Params())

			return a
		}
	)

	service.Services = make(service.ServicesData)
	design.Root.Services = []*design.ServiceExpr{httpSvc.ServiceExpr}

	cases := map[string]struct {
		Endpoint *rest.HTTPEndpointExpr
		Expected string
	}{
		"single-path-no-param":            {Endpoint: endpoint("/test"), Expected: pathWithoutParams},
		"single-path-one-param":           {Endpoint: endpoint("/test/{id}"), Expected: pathWithOneParam},
		"single-path-multiple-params":     {Endpoint: endpoint("/test/{id}/view/{view}"), Expected: pathWithMultipleParams},
		"alternative-paths":               {Endpoint: endpoint("/test", "/test/{id}", "/test/{id}/view/{view}"), Expected: pathWithAlternatives},
		"path-with-string-slice-param":    {Endpoint: endpoint("/test/{slice_string}"), Expected: pathWithStringSliceParam},
		"path-with-int-slice-param":       {Endpoint: endpoint("/test/{slice_int}"), Expected: pathWithIntSliceParam},
		"path-with-int32-slice-param":     {Endpoint: endpoint("/test/{slice_int32}"), Expected: pathWithInt32SliceParam},
		"path-with-int64-slice-param":     {Endpoint: endpoint("/test/{slice_int64}"), Expected: pathWithInt64SliceParam},
		"path-with-uint-slice-param":      {Endpoint: endpoint("/test/{slice_uint}"), Expected: pathWithUintSliceParam},
		"path-with-uint32-slice-param":    {Endpoint: endpoint("/test/{slice_uint32}"), Expected: pathWithUint32SliceParam},
		"path-with-uint64-slice-param":    {Endpoint: endpoint("/test/{slice_uint64}"), Expected: pathWithUint64SliceParam},
		"path-with-float33-slice-param":   {Endpoint: endpoint("/test/{slice_float32}"), Expected: pathWithFloat32SliceParam},
		"path-with-float64-slice-param":   {Endpoint: endpoint("/test/{slice_float64}"), Expected: pathWithFloat64SliceParam},
		"path-with-bool-slice-param":      {Endpoint: endpoint("/test/{slice_bool}"), Expected: pathWithBoolSliceParam},
		"path-with-interface-slice-param": {Endpoint: endpoint("/test/{slice_interface}"), Expected: pathWithInterfaceSliceParam},
	}

	for k, tc := range cases {
		buf := new(bytes.Buffer)
		s := PathSection(tc.Endpoint)
		e := s.Write(buf)
		actual := buf.String()

		if e != nil {
			t.Errorf("%s: failed to execute template, error %s", k, e)
		} else if actual != tc.Expected {
			t.Errorf("%s: got %v, expected %v", k, actual, tc.Expected)
		}
	}
}
