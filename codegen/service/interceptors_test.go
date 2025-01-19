package service

import (
	"bytes"
	"flag"
	"go/format"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata"
	"goa.design/goa/v3/expr"
)

var updateGolden = false

func init() {
	flag.BoolVar(&updateGolden, "w", false, "update golden files")
}

func TestInterceptors(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"no-interceptors", testdata.NoInterceptorsDSL},
		{"single-api-server-interceptor", testdata.SingleAPIServerInterceptorDSL},
		{"single-service-server-interceptor", testdata.SingleServiceServerInterceptorDSL},
		{"single-method-server-interceptor", testdata.SingleMethodServerInterceptorDSL},
		{"single-client-interceptor", testdata.SingleClientInterceptorDSL},
		{"multiple-interceptors", testdata.MultipleInterceptorsDSL},
		{"interceptor-with-read-payload", testdata.InterceptorWithReadPayloadDSL},
		{"interceptor-with-write-payload", testdata.InterceptorWithWritePayloadDSL},
		{"interceptor-with-read-write-payload", testdata.InterceptorWithReadWritePayloadDSL},
		{"interceptor-with-read-result", testdata.InterceptorWithReadResultDSL},
		{"interceptor-with-write-result", testdata.InterceptorWithWriteResultDSL},
		{"interceptor-with-read-write-result", testdata.InterceptorWithReadWriteResultDSL},
		{"streaming-interceptors", testdata.StreamingInterceptorsDSL},
		{"streaming-interceptors-with-read-payload", testdata.StreamingInterceptorsWithReadPayloadDSL},
		{"streaming-interceptors-with-read-result", testdata.StreamingInterceptorsWithReadResultDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := runDSL(t, c.DSL)
			require.Len(t, root.Services, 1)

			fs := InterceptorsFile("goa.design/goa/example", root.Services[0])

			if c.Name == "no-interceptors" {
				assert.Nil(t, fs)
				return
			}

			require.NotNil(t, fs)

			buf := new(bytes.Buffer)
			for _, s := range fs.SectionTemplates[1:] {
				require.NoError(t, s.Write(buf))
			}
			bs, err := format.Source(buf.Bytes())
			require.NoError(t, err, buf.String())
			code := strings.ReplaceAll(string(bs), "\r\n", "\n")

			golden := filepath.Join("testdata", "interceptors", c.Name+".golden")
			compareOrUpdateGolden(t, code, golden)
		})
	}
}

func TestInvalidInterceptors(t *testing.T) {
	cases := []struct {
		Name        string
		DSL         func()
		ErrContains string
	}{
		{
			Name:        "streaming-result-interceptor",
			DSL:         testdata.StreamingResultInterceptorDSL,
			ErrContains: "cannot be applied because the method result is streaming",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := runDSLWithError(t, c.DSL)
			require.Error(t, err)
			assert.Contains(t, err.Error(), c.ErrContains)
		})
	}
}

func TestCollectAttributes(t *testing.T) {
	cases := []struct {
		name      string
		attrNames *expr.AttributeExpr
		parent    *expr.AttributeExpr
		want      []*AttributeData
	}{
		{
			name:      "nil-attributes",
			attrNames: nil,
			parent:    &expr.AttributeExpr{Type: &expr.Object{}},
			want:      nil,
		},
		{
			name:      "non-object-attributes",
			attrNames: &expr.AttributeExpr{Type: expr.Primitive(expr.StringKind)},
			parent:    &expr.AttributeExpr{Type: &expr.Object{}},
			want:      nil,
		},
		{
			name: "simple-string-attribute",
			attrNames: &expr.AttributeExpr{
				Type: &expr.Object{
					{Name: "name", Attribute: &expr.AttributeExpr{Type: expr.Primitive(expr.StringKind)}},
				},
			},
			parent: &expr.AttributeExpr{
				Type: &expr.Object{
					{Name: "name", Attribute: &expr.AttributeExpr{Type: expr.Primitive(expr.StringKind)}},
				},
				Validation: &expr.ValidationExpr{Required: []string{"name"}},
			},
			want: []*AttributeData{
				{Name: "Name", TypeRef: "string", FieldPointer: false},
			},
		},
		{
			name: "pointer-primitive",
			attrNames: &expr.AttributeExpr{
				Type: &expr.Object{
					{Name: "age", Attribute: &expr.AttributeExpr{Type: expr.Primitive(expr.IntKind)}},
				},
			},
			parent: &expr.AttributeExpr{
				Type: &expr.Object{
					{Name: "age", Attribute: &expr.AttributeExpr{Type: expr.Primitive(expr.IntKind), Meta: map[string][]string{"struct:field:pointer": {"true"}}}},
				},
			},
			want: []*AttributeData{
				{Name: "Age", TypeRef: "int", FieldPointer: true},
			},
		},
		{
			name: "multiple-attributes",
			attrNames: &expr.AttributeExpr{
				Type: &expr.Object{
					{Name: "name", Attribute: &expr.AttributeExpr{Type: expr.Primitive(expr.StringKind)}},
					{Name: "age", Attribute: &expr.AttributeExpr{Type: expr.Primitive(expr.IntKind)}},
				},
			},
			parent: &expr.AttributeExpr{
				Type: &expr.Object{
					{Name: "name", Attribute: &expr.AttributeExpr{Type: expr.Primitive(expr.StringKind)}},
					{Name: "age", Attribute: &expr.AttributeExpr{Type: expr.Primitive(expr.IntKind), Meta: map[string][]string{"struct:field:pointer": {"true"}}}},
				},
				Validation: &expr.ValidationExpr{Required: []string{"name"}},
			},
			want: []*AttributeData{
				{Name: "Name", TypeRef: "string", FieldPointer: false},
				{Name: "Age", TypeRef: "int", FieldPointer: true},
			},
		},
		{
			name: "attribute-not-in-parent",
			attrNames: &expr.AttributeExpr{
				Type: &expr.Object{
					{Name: "missing", Attribute: &expr.AttributeExpr{Type: expr.Primitive(expr.StringKind)}},
				},
			},
			parent: &expr.AttributeExpr{
				Type: &expr.Object{
					{Name: "name", Attribute: &expr.AttributeExpr{Type: expr.Primitive(expr.StringKind)}},
				},
				Validation: &expr.ValidationExpr{Required: []string{"name"}},
			},
			want: []*AttributeData{nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scope := codegen.NewNameScope()
			got := collectAttributes(tc.attrNames, tc.parent, scope)
			assert.Equal(t, tc.want, got)
		})
	}
}

func compareOrUpdateGolden(t *testing.T, code, golden string) {
	t.Helper()
	if updateGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0750))
		require.NoError(t, os.WriteFile(golden, []byte(code), 0640))
		return
	}
	data, err := os.ReadFile(golden)
	require.NoError(t, err)
	if runtime.GOOS == "windows" {
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	}
	assert.Equal(t, string(data), code)
}
