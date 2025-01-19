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

func TestClient(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{"client-single", testdata.SingleEndpointDSL, testdata.SingleMethodClient},
		{"client-use", testdata.UseEndpointDSL, testdata.UseMethodClient},
		{"client-multiple", testdata.MultipleEndpointsDSL, testdata.MultipleMethodsClient},
		{"client-no-payload", testdata.NoPayloadEndpointDSL, testdata.NoPayloadMethodsClient},
		{"client-with-result", testdata.WithResultEndpointDSL, testdata.WithResultMethodClient},
		{"client-streaming-result", testdata.StreamingResultMethodDSL, testdata.StreamingResultMethodClient},
		{"client-streaming-result-no-payload", testdata.StreamingResultNoPayloadMethodDSL, testdata.StreamingResultNoPayloadMethodClient},
		{"client-streaming-payload", testdata.StreamingPayloadMethodDSL, testdata.StreamingPayloadMethodClient},
		{"client-streaming-payload-no-payload", testdata.StreamingPayloadNoPayloadMethodDSL, testdata.StreamingPayloadNoPayloadMethodClient},
		{"client-streaming-payload-no-result", testdata.StreamingPayloadNoResultMethodDSL, testdata.StreamingPayloadNoResultMethodClient},
		{"client-bidirectional-streaming", testdata.BidirectionalStreamingMethodDSL, testdata.BidirectionalStreamingMethodClient},
		{"client-bidirectional-streaming-no-payload", testdata.BidirectionalStreamingNoPayloadMethodDSL, testdata.BidirectionalStreamingNoPayloadMethodClient},
		{"client-interceptor", testdata.EndpointWithClientInterceptorDSL, testdata.InterceptorClient},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			codegen.RunDSL(t, c.DSL)
			require.Len(t, expr.Root.Services, 1)
			fs := ClientFile("test/gen", expr.Root.Services[0])
			require.NotNil(t, fs)
			buf := new(bytes.Buffer)
			for _, s := range fs.SectionTemplates[1:] {
				require.NoError(t, s.Write(buf))
			}
			bs, err := format.Source(buf.Bytes())
			require.NoError(t, err, buf.String())
			code := strings.ReplaceAll(string(bs), "\r\n", "\n")
			assert.Equal(t, c.Code, code)
		})
	}
}
