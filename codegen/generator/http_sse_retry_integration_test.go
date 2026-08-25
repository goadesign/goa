// This file checks retry values in complete generated HTTP SSE transports.
// The server writes the designed integer field and the client rebuilds the
// service result and ignores retry fields that are not ASCII digits.
package generator

import (
	"testing"

	"goa.design/goa/v3/dsl"
)

// httpSSERetryContractTest runs against the temporary generated module.
const httpSSERetryContractTest = `package integration

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/sse_retry"
	client "generated.local/gen/http/sse_retry/client"
	server "generated.local/gen/http/sse_retry/server"
	goahttp "goa.design/goa/v3/http"
)

// retryService sends one event with both fields selected.
type retryService struct{}

func (*retryService) Watch(_ context.Context, stream service.WatchServerStream) error {
	data := "null"
	retry := 2500
	return stream.Send(&service.Event{Data: &data, Retry: &retry})
}

func TestRetryRoundTrip(t *testing.T) {
	handler := server.NewWatchHandler(
		service.NewWatchEndpoint(&retryService{}),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(_ context.Context, _ http.ResponseWriter, err error) { t.Errorf("serve SSE: %v", err) },
		nil,
	)
	httpServer := httptest.NewServer(handler)
	defer httpServer.Close()

	stream := openWatch(t, httpServer.URL, http.DefaultClient)
	event, err := stream.Recv()
	require.NoError(t, err)
	require.NotNil(t, event.Data)
	require.Equal(t, "null", *event.Data)
	require.NotNil(t, event.Retry)
	require.Equal(t, 2500, *event.Retry)
}

func TestMalformedRetryIsIgnored(t *testing.T) {
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, err := io.WriteString(w, "retry: later\ndata: ready\n\n")
		require.NoError(t, err)
	}))
	defer httpServer.Close()

	stream := openWatch(t, httpServer.URL, http.DefaultClient)
	event, err := stream.Recv()
	require.NoError(t, err)
	require.NotNil(t, event.Data)
	require.Equal(t, "ready", *event.Data)
	require.Nil(t, event.Retry)
}

// openWatch starts the generated client stream against url.
func openWatch(t *testing.T, url string, doer goahttp.Doer) client.WatchClientStream {
	t.Helper()
	transport := client.NewClient(
		"http",
		strings.TrimPrefix(url, "http://"),
		doer,
		goahttp.RequestEncoder,
		goahttp.ResponseDecoder,
		false,
	)
	raw, err := transport.Watch()(context.Background(), nil)
	require.NoError(t, err)
	return raw.(client.WatchClientStream)
}
`

// TestGeneratedHTTPSSERetryParsing checks valid and invalid retry values using
// a generated server and client rather than template fragments.
func TestGeneratedHTTPSSERetryParsing(t *testing.T) {
	dir := generateViewedTransportModule(t, httpSSERetryDSL)
	writeGeneratedContractTest(t, dir, ".", httpSSERetryContractTest)
	runGeneratedPackageTests(t, dir, ".")
}

// httpSSERetryDSL defines one event with optional data and retry fields. The
// generated service uses pointers so nil and selected values remain distinct.
func httpSSERetryDSL() {
	event := dsl.Type("Event", func() {
		dsl.Attribute("data", dsl.String)
		dsl.Attribute("retry", dsl.Int)
	})
	dsl.Service("SSE Retry", func() {
		dsl.Method("Watch", func() {
			dsl.StreamingResult(event)
			dsl.HTTP(func() {
				dsl.GET("/watch")
				dsl.ServerSentEvents("data", func() {
					dsl.SSEEventRetry("retry")
				})
			})
		})
	})
}
