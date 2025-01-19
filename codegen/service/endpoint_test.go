package service

import (
	"bytes"
	"go/format"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata"
	"goa.design/goa/v3/expr"
)

func TestEndpoint(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"endpoint-single", testdata.SingleEndpointDSL, testdata.SingleEndpoint},
		{"endpoint-use", testdata.UseEndpointDSL, testdata.UseEndpoint},
		{"endpoint-multiple", testdata.MultipleEndpointsDSL, testdata.MultipleEndpoints},
		{"endpoint-no-payload", testdata.NoPayloadEndpointDSL, testdata.NoPayloadEndpoint},
		{"endpoint-with-result", testdata.WithResultEndpointDSL, testdata.WithResultEndpoint},
		{"endpoint-with-result-multiple-views", testdata.WithResultMultipleViewsEndpointDSL, testdata.WithResultMultipleViewsEndpoint},
		{"endpoint-streaming-result", testdata.StreamingResultEndpointDSL, testdata.StreamingResultMethodEndpoint},
		{"endpoint-streaming-result-no-payload", testdata.StreamingResultNoPayloadEndpointDSL, testdata.StreamingResultNoPayloadMethodEndpoint},
		{"endpoint-streaming-result-with-views", testdata.StreamingResultWithViewsMethodDSL, testdata.StreamingResultWithViewsMethodEndpoint},
		{"endpoint-streaming-payload", testdata.StreamingPayloadEndpointDSL, testdata.StreamingPayloadMethodEndpoint},
		{"endpoint-streaming-payload-no-payload", testdata.StreamingPayloadNoPayloadMethodDSL, testdata.StreamingPayloadNoPayloadMethodEndpoint},
		{"endpoint-streaming-payload-no-result", testdata.StreamingPayloadNoResultMethodDSL, testdata.StreamingPayloadNoResultMethodEndpoint},
		{"endpoint-bidirectional-streaming", testdata.BidirectionalStreamingEndpointDSL, testdata.BidirectionalStreamingMethodEndpoint},
		{"endpoint-bidirectional-streaming-no-payload", testdata.BidirectionalStreamingNoPayloadMethodDSL, testdata.BidirectionalStreamingNoPayloadMethodEndpoint},
		{"endpoint-with-server-interceptor", testdata.EndpointWithServerInterceptorDSL, testdata.EndpointWithServerInterceptor},
		{"endpoint-with-multiple-interceptors", testdata.EndpointWithMultipleInterceptorsDSL, testdata.EndpointWithMultipleInterceptors},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			codegen.RunDSL(t, c.DSL)
			require.Len(t, expr.Root.Services, 1)
			fs := EndpointFile("goa.design/goa/example", expr.Root.Services[0])
			require.NotNil(t, fs)
			buf := new(bytes.Buffer)
			for _, s := range fs.SectionTemplates[1:] {
				require.NoError(t, s.Write(buf))
			}
			bs, err := format.Source(buf.Bytes())
			require.NoError(t, err, buf.String())
			code := string(bs)
			code = strings.ReplaceAll(code, "\r\n", "\n")
			assert.Equal(t, c.Code, code)
		})
	}
}
