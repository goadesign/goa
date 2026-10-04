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
	"goa.design/goa/v3/codegen/testutil"
)

func TestEndpoint(t *testing.T) {
	cases := []struct {
		Name   string
		DSL    func()
		Code   string
		Golden bool
	}{
		{"endpoint-single", testdata.SingleEndpointDSL, testdata.SingleEndpoint, false},
		{"endpoint-use", testdata.UseEndpointDSL, testdata.UseEndpoint, false},
		{"endpoint-multiple", testdata.MultipleEndpointsDSL, testdata.MultipleEndpoints, false},
		{"endpoint-no-payload", testdata.NoPayloadEndpointDSL, testdata.NoPayloadEndpoint, false},
		{"endpoint-with-result", testdata.WithResultEndpointDSL, "", true},
		{"endpoint-with-result-multiple-views", testdata.WithResultMultipleViewsEndpointDSL, "", true},
		{"endpoint-streaming-result", testdata.StreamingResultEndpointDSL, testdata.StreamingResultMethodEndpoint, false},
		{"endpoint-mixed-results", testdata.MixedResultsEndpointDSL, testdata.MixedResultsMethodEndpoint, false},
		{"endpoint-mixed-results-with-views", testdata.MixedResultsWithViewsEndpointDSL, "", true},
		{"endpoint-streaming-result-no-payload", testdata.StreamingResultNoPayloadEndpointDSL, testdata.StreamingResultNoPayloadMethodEndpoint, false},
		{"endpoint-streaming-result-with-views", testdata.StreamingResultWithViewsMethodDSL, testdata.StreamingResultWithViewsMethodEndpoint, false},
		{"endpoint-streaming-payload", testdata.StreamingPayloadEndpointDSL, testdata.StreamingPayloadMethodEndpoint, false},
		{"endpoint-streaming-payload-no-payload", testdata.StreamingPayloadNoPayloadMethodDSL, testdata.StreamingPayloadNoPayloadMethodEndpoint, false},
		{"endpoint-streaming-payload-no-result", testdata.StreamingPayloadNoResultMethodDSL, testdata.StreamingPayloadNoResultMethodEndpoint, false},
		{"endpoint-bidirectional-streaming", testdata.BidirectionalStreamingEndpointDSL, testdata.BidirectionalStreamingMethodEndpoint, false},
		{"endpoint-bidirectional-streaming-no-payload", testdata.BidirectionalStreamingNoPayloadMethodDSL, testdata.BidirectionalStreamingNoPayloadMethodEndpoint, false},
		{"endpoint-with-server-interceptor", testdata.EndpointWithServerInterceptorDSL, testdata.EndpointWithServerInterceptor, false},
		{"endpoint-with-multiple-interceptors", testdata.EndpointWithMultipleInterceptorsDSL, testdata.EndpointWithMultipleInterceptors, false},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)
			plan := mustServicePlan(t, root)
			require.Len(t, root.Services, 1)
			fs := endpointFile(plan, plan.facts.services[0])
			require.NotNil(t, fs)
			buf := new(bytes.Buffer)
			for _, s := range fs.SectionTemplates[1:] {
				require.NoError(t, s.Write(buf))
			}
			bs, err := format.Source(buf.Bytes())
			require.NoError(t, err, buf.String())
			code := string(bs)
			code = strings.ReplaceAll(code, "\r\n", "\n")
			if c.Golden {
				testutil.AssertGo(t, "testdata/golden/"+c.Name+".go.golden", code)
			} else {
				assert.Equal(t, c.Code, code)
			}
		})
	}
}
