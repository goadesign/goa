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
		Name              string
		DSL               func()
		expectedFileCount int
	}{
		{"no-interceptors", testdata.NoInterceptorsDSL, 0},
		{"single-api-server-interceptor", testdata.SingleAPIServerInterceptorDSL, 2},
		{"single-service-server-interceptor", testdata.SingleServiceServerInterceptorDSL, 2},
		{"single-method-server-interceptor", testdata.SingleMethodServerInterceptorDSL, 2},
		{"single-client-interceptor", testdata.SingleClientInterceptorDSL, 2},
		{"multiple-interceptors", testdata.MultipleInterceptorsExampleDSL, 3},
		{"interceptor-with-read-payload", testdata.InterceptorWithReadPayloadDSL, 3},
		{"interceptor-with-write-payload", testdata.InterceptorWithWritePayloadDSL, 3},
		{"interceptor-with-read-write-payload", testdata.InterceptorWithReadWritePayloadDSL, 3},
		{"interceptor-with-read-result", testdata.InterceptorWithReadResultDSL, 3},
		{"interceptor-with-write-result", testdata.InterceptorWithWriteResultDSL, 3},
		{"interceptor-with-read-write-result", testdata.InterceptorWithReadWriteResultDSL, 3},
		{"streaming-interceptors", testdata.StreamingInterceptorsDSL, 2},
		{"streaming-interceptors-with-read-payload", testdata.StreamingInterceptorsWithReadPayloadDSL, 2},
		{"streaming-interceptors-with-read-result", testdata.StreamingInterceptorsWithReadResultDSL, 2},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := runDSL(t, c.DSL)
			require.Len(t, root.Services, 1)

			fs := InterceptorsFiles("goa.design/goa/example", root.Services[0])

			require.Len(t, fs, c.expectedFileCount)
			for _, f := range fs {
				buf := new(bytes.Buffer)
				for _, s := range f.SectionTemplates[1:] {
					require.NoError(t, s.Write(buf))
				}
				bs, err := format.Source(buf.Bytes())
				require.NoError(t, err, buf.String())
				code := strings.ReplaceAll(string(bs), "\r\n", "\n")

				golden := filepath.Join("testdata", "interceptors", c.Name+"_"+filepath.Base(f.Path)+".golden")
				compareOrUpdateGolden(t, code, golden)
			}
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
				{Name: "Name", TypeRef: "string", Pointer: false},
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
				{Name: "Age", TypeRef: "int", Pointer: true},
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
				{Name: "Name", TypeRef: "string", Pointer: false},
				{Name: "Age", TypeRef: "int", Pointer: true},
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
