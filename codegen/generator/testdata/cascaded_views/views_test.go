// These tests exercise generated Goa endpoints, HTTP codecs, JSON-RPC codecs,
// and clients to verify that cascaded result views preserve nested fields in
// extended views, omit them in default views, and run view-specific checks.
package checks_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	genorderscli "generated.local/gen/http/orders/client"
	genorderssrv "generated.local/gen/http/orders/server"
	genrpccli "generated.local/gen/jsonrpc/rpcorders/client"
	genrpcsrv "generated.local/gen/jsonrpc/rpcorders/server"
	genorders "generated.local/gen/orders"
	genordersviews "generated.local/gen/orders/views"
	genrpcorders "generated.local/gen/rpcorders"
	goahttp "goa.design/goa/v3/http"
)

type (
	// orderService returns a three-level Order with populated Label.Name values.
	orderService struct {
		// labelName controls the nested Label.Name value returned by the handler.
		labelName string
	}

	// rpcOrderService returns the same three-level Order over JSON-RPC.
	rpcOrderService struct {
		// labelName controls the nested Label.Name value returned by the handler.
		labelName string
	}
)

// Get returns an Order projected with the requested view.
func (s *orderService) Get(_ context.Context, payload *genorders.GetPayload) (*genorders.Order, string, error) {
	return s.sampleOrder(), payload.View, nil
}

// GetExtended returns an Order for an endpoint pinned to the extended view.
func (s *orderService) GetExtended(_ context.Context) (*genorders.Order, error) {
	return s.sampleOrder(), nil
}

// sampleOrder builds a populated three-level service Order.
func (s *orderService) sampleOrder() *genorders.Order {
	name := s.labelName
	item := &genorders.Item{
		ID: "item-1",
		Labels: []*genorders.Label{{
			ID:   "lbl-1",
			Name: &name,
		}},
	}
	return &genorders.Order{
		ID:    "ord-1",
		Items: []*genorders.Item{item},
		ByKey: map[string]*genorders.Item{"primary": item},
	}
}

// Get returns an Order projected with the requested view over JSON-RPC.
func (s *rpcOrderService) Get(_ context.Context, payload *genrpcorders.GetPayload) (*genrpcorders.Order, string, error) {
	name := s.labelName
	item := &genrpcorders.Item{
		ID: "item-1",
		Labels: []*genrpcorders.Label{{
			ID:   "lbl-1",
			Name: &name,
		}},
	}
	return &genrpcorders.Order{
		ID:    "ord-1",
		Items: []*genrpcorders.Item{item},
		ByKey: map[string]*genrpcorders.Item{"primary": item},
	}, payload.View, nil
}

// TestHTTPCascadedResultViews verifies wire encoding and client decoding of
// nested result views over HTTP for both dynamic and fixed view endpoints.
func TestHTTPCascadedResultViews(t *testing.T) {
	service := &orderService{labelName: "fragile"}
	mux := goahttp.NewMuxer()
	server := genorderssrv.New(
		genorders.NewEndpoints(service),
		mux,
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		nil,
		nil,
	)
	genorderssrv.Mount(mux, server)

	peer := httptest.NewServer(mux)
	defer peer.Close()
	address, err := url.Parse(peer.URL)
	if err != nil {
		t.Fatal(err)
	}
	transport := genorderscli.NewClient(
		address.Scheme,
		address.Host,
		peer.Client(),
		goahttp.RequestEncoder,
		goahttp.ResponseDecoder,
		false,
	)
	client := genorders.NewClient(transport.Get(), transport.GetExtended())

	for _, tc := range []struct {
		name     string
		view     string
		path     string
		wantName bool
	}{
		{
			name:     "default view omits nested label name",
			view:     "default",
			path:     "/orders?view=default",
			wantName: false,
		},
		{
			name:     "extended view preserves nested label name across array and map",
			view:     "extended",
			path:     "/orders?view=extended",
			wantName: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if rec.Code != http.StatusOK {
				t.Fatalf("HTTP status %d: %s", rec.Code, rec.Body.String())
			}
			hasWireName := bytes.Contains(rec.Body.Bytes(), []byte(`"name":"fragile"`))
			if hasWireName != tc.wantName {
				t.Errorf("wire JSON contains nested label name = %t, want %t: %s", hasWireName, tc.wantName, rec.Body.String())
			}

			got, err := client.Get(t.Context(), &genorders.GetPayload{View: tc.view})
			if err != nil {
				t.Fatalf("client.Get(%q): %v", tc.view, err)
			}
			assertDecodedOrder(t, got, tc.wantName)
		})
	}

	t.Run("fixed extended endpoint preserves nested label name", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/orders/extended", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("HTTP status %d: %s", rec.Code, rec.Body.String())
		}
		if !bytes.Contains(rec.Body.Bytes(), []byte(`"name":"fragile"`)) {
			t.Errorf("fixed extended wire JSON lost nested label name: %s", rec.Body.String())
		}
		got, err := client.GetExtended(t.Context())
		if err != nil {
			t.Fatalf("client.GetExtended(): %v", err)
		}
		assertDecodedOrder(t, got, true)
	})
}

// TestJSONRPCCascadedResultViews verifies wire encoding and client decoding of
// nested result views over JSON-RPC.
func TestJSONRPCCascadedResultViews(t *testing.T) {
	service := &rpcOrderService{labelName: "fragile"}
	server := genrpcsrv.New(
		genrpcorders.NewEndpoints(service),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(_ context.Context, _ http.ResponseWriter, err error) {
			t.Errorf("JSON-RPC server error: %v", err)
		},
	)

	for _, tc := range []struct {
		name     string
		view     string
		wantName bool
	}{
		{
			name:     "default view omits nested label name",
			view:     "default",
			wantName: false,
		},
		{
			name:     "extended view preserves nested label name",
			view:     "extended",
			wantName: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/rpc", nil)
			id, err := genrpccli.EncodeGetRequest(goahttp.RequestEncoder)(req, &genrpcorders.GetPayload{View: tc.view})
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}
			if err := req.Body.Close(); err != nil {
				t.Fatal(err)
			}

			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(body)))
			if rec.Code != http.StatusOK {
				t.Fatalf("JSON-RPC status %d: %s", rec.Code, rec.Body.String())
			}
			hasWireName := bytes.Contains(rec.Body.Bytes(), []byte(`"name":"fragile"`))
			if hasWireName != tc.wantName {
				t.Errorf("JSON-RPC wire contains nested label name = %t, want %t: %s", hasWireName, tc.wantName, rec.Body.String())
			}

			msg := rec.Result()
			defer func() {
				if err := msg.Body.Close(); err != nil {
					t.Error(err)
				}
			}()
			msg.Request = req
			raw, err := genrpccli.DecodeGetResponse(goahttp.ResponseDecoder, false)(msg, id)
			if err != nil {
				t.Fatalf("DecodeGetResponse(%q): %v", tc.view, err)
			}
			client := genrpcorders.NewClient(func(context.Context, any) (any, error) {
				return raw, nil
			})
			got, err := client.Get(t.Context(), &genrpcorders.GetPayload{View: tc.view})
			if err != nil {
				t.Fatalf("JSON-RPC client.Get(%q): %v", tc.view, err)
			}
			if len(got.Items) != 1 || len(got.Items[0].Labels) != 1 {
				t.Fatalf("unexpected decoded JSON-RPC items: %+v", got.Items)
			}
			gotLabel := got.Items[0].Labels[0]
			if tc.wantName {
				if gotLabel.Name == nil || *gotLabel.Name != "fragile" {
					t.Errorf("JSON-RPC extended label Name = %v, want %q", gotLabel.Name, "fragile")
				}
			} else if gotLabel.Name != nil {
				t.Errorf("JSON-RPC default label Name = %q, want nil", *gotLabel.Name)
			}
		})
	}
}

// TestCascadedViewedResultValidation verifies that view validation checks
// nested extended fields only when the extended view is selected.
func TestCascadedViewedResultValidation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		view      string
		labelName string
		wantErr   bool
	}{
		{
			name:      "default view ignores invalid extended-only nested label name",
			view:      "default",
			labelName: "x",
			wantErr:   false,
		},
		{
			name:      "extended view rejects nested label name shorter than MinLength",
			view:      "extended",
			labelName: "x",
			wantErr:   true,
		},
		{
			name:      "extended view accepts valid nested label name",
			view:      "extended",
			labelName: "ok",
			wantErr:   false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &orderService{labelName: tc.labelName}
			viewed := genorders.NewViewedOrder(svc.sampleOrder(), tc.view)
			err := genordersviews.ValidateOrder(viewed)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateOrder(view=%q, labelName=%q) error = %v, wantErr=%t", tc.view, tc.labelName, err, tc.wantErr)
			}
		})
	}
}

// assertDecodedOrder checks both ArrayOf and MapOf nested Label.Name values on
// a client-decoded service Order.
func assertDecodedOrder(t *testing.T, got *genorders.Order, wantName bool) {
	t.Helper()
	if len(got.Items) != 1 || len(got.Items[0].Labels) != 1 {
		t.Fatalf("unexpected decoded items: %+v", got.Items)
	}
	mapItem := got.ByKey["primary"]
	if mapItem == nil || len(mapItem.Labels) != 1 {
		t.Fatalf("unexpected decoded map item: %+v", got.ByKey)
	}
	for _, source := range []struct {
		label string
		got   *genorders.Label
	}{
		{"items[0].labels[0]", got.Items[0].Labels[0]},
		{"by_key[primary].labels[0]", mapItem.Labels[0]},
	} {
		if source.got.ID != "lbl-1" {
			t.Errorf("%s ID = %q, want %q", source.label, source.got.ID, "lbl-1")
		}
		if wantName {
			if source.got.Name == nil || *source.got.Name != "fragile" {
				t.Errorf("%s Name = %v, want %q", source.label, source.got.Name, "fragile")
			}
		} else if source.got.Name != nil {
			t.Errorf("%s Name = %q, want nil", source.label, *source.got.Name)
		}
	}
}
