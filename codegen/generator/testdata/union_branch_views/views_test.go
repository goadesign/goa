// These tests use the ordinary Goa endpoints, HTTP codecs and generated client.
// A selected branch view must hide its private field without changing unfinished data.
package checks_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	genreceiptcli "generated.local/gen/http/receipts/client"
	genreceiptsrv "generated.local/gen/http/receipts/server"
	genrpccli "generated.local/gen/jsonrpc/rpcreceipts/client"
	genrpcsrv "generated.local/gen/jsonrpc/rpcreceipts/server"
	genreceipts "generated.local/gen/receipts"
	genrpcreceipts "generated.local/gen/rpcreceipts"
	goahttp "goa.design/goa/v3/http"
)

type (
	receiptService struct {
		branch string
		calls  int
	}

	// rpcReceiptService returns the same declared branch through native JSON-RPC.
	rpcReceiptService struct{}
)

func (s *receiptService) Book(_ context.Context, payload *genreceipts.BookPayload) (*genreceipts.OutcomeResult, error) {
	if payload.Destination != "Paris" {
		panic("native decoder changed the domain input")
	}
	s.calls++
	receipt := &genreceipts.Receipt{Reference: "booked", Internal: "must-not-leak"}
	if s.branch == "invalid_public" {
		receipt.Reference = ""
	}
	if s.branch == "private_empty" {
		receipt.Internal = ""
	}
	switch s.branch {
	case "input_required":
		return &genreceipts.OutcomeResult{Outcome: genreceipts.NewOutcomeInputRequired(&genreceipts.MoreInput{State: "original-state"})}, nil
	case "note", "note_present":
		if s.branch == "note_present" {
			note := "optional-note"
			receipt.Note = &note
		}
		return &genreceipts.OutcomeResult{Outcome: genreceipts.NewOutcomeNote(receipt)}, nil
	case "full":
		return &genreceipts.OutcomeResult{Outcome: genreceipts.NewOutcomeFull(receipt)}, nil
	case "group":
		return &genreceipts.OutcomeResult{Outcome: genreceipts.NewOutcomeGroup(&genreceipts.ReceiptGroup{
			Single:    receipt,
			Many:      []*genreceipts.Receipt{receipt},
			Named:     map[string]*genreceipts.Receipt{"first": receipt},
			Collected: genreceipts.ReceiptCollection{receipt},
		})}, nil
	default:
		return &genreceipts.OutcomeResult{Outcome: genreceipts.NewOutcomeComplete(receipt)}, nil
	}
}

func TestNativeUnionBranchView(t *testing.T) {
	for _, name := range []string{"complete", "full", "note", "note_present", "group", "input_required"} {
		t.Run(name, func(t *testing.T) {
			service := &receiptService{branch: name}
			mux := goahttp.NewMuxer()
			server := genreceiptsrv.New(genreceipts.NewEndpoints(service), mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, nil, nil)
			genreceiptsrv.Mount(mux, server)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/book", strings.NewReader(`{"destination":"Paris"}`)))
			if response.Code != http.StatusOK {
				t.Fatalf("native HTTP status %d: %s", response.Code, response.Body.String())
			}
			if name == "input_required" {
				if !bytes.Contains(response.Body.Bytes(), []byte(`"state":"original-state"`)) {
					t.Fatal("unfinished branch lost its state")
				}
			} else if name == "note" || name == "note_present" {
				if bytes.Contains(response.Body.Bytes(), []byte(`"reference"`)) || bytes.Contains(response.Body.Bytes(), []byte(`"internal"`)) {
					t.Fatal("note view returned excluded fields")
				}
				if name == "note_present" && !bytes.Contains(response.Body.Bytes(), []byte(`"note":"optional-note"`)) {
					t.Fatal("note view lost its optional value")
				}
			} else if name == "full" {
				if !bytes.Contains(response.Body.Bytes(), []byte(`"internal":"must-not-leak"`)) {
					t.Fatal("default branch view lost its required field")
				}
			} else {
				if bytes.Contains(response.Body.Bytes(), []byte(`"internal"`)) {
					t.Fatalf("selected public branch view leaked its private field: %s", response.Body.String())
				}
				if !bytes.Contains(response.Body.Bytes(), []byte(`"reference":"booked"`)) {
					t.Fatal("public receipt field was lost")
				}
			}
			peer := httptest.NewServer(mux)
			defer peer.Close()
			address, err := url.Parse(peer.URL)
			if err != nil {
				t.Fatal(err)
			}
			transport := genreceiptcli.NewClient(address.Scheme, address.Host, peer.Client(), goahttp.RequestEncoder, goahttp.ResponseDecoder, false)
			client := genreceipts.NewClient(transport.Book())
			result, err := client.Book(t.Context(), &genreceipts.BookPayload{Destination: "Paris"})
			if err != nil {
				t.Fatal(err)
			}
			if name == "input_required" {
				branch, ok := result.Outcome.AsInputRequired()
				if !ok || branch.State != "original-state" {
					t.Fatal("native client lost unfinished state")
				}
			} else if name == "note" || name == "note_present" {
				branch, ok := result.Outcome.AsNote()
				if !ok || branch.Reference != "" || branch.Internal != "" {
					t.Fatal("native client did not retain the optional note view")
				}
				if name == "note_present" && (branch.Note == nil || *branch.Note != "optional-note") {
					t.Fatal("native client lost the optional note value")
				}
			} else if name == "full" {
				branch, ok := result.Outcome.AsFull()
				if !ok || branch.Reference != "booked" || branch.Internal != "must-not-leak" {
					t.Fatal("native client lost the default branch view")
				}
			} else if name == "group" {
				branch, ok := result.Outcome.AsGroup()
				if !ok || len(branch.Many) != 1 || len(branch.Named) != 1 || len(branch.Collected) != 1 {
					t.Fatal("native client lost nested receipts")
				}
				for _, receipt := range []*genreceipts.Receipt{branch.Single, branch.Many[0], branch.Named["first"], branch.Collected[0]} {
					if receipt.Reference != "booked" || receipt.Internal != "" {
						t.Fatal("native client did not retain exactly the nested public view")
					}
				}
			} else {
				branch, ok := result.Outcome.AsComplete()
				if !ok || branch.Reference != "booked" || branch.Internal != "" {
					t.Fatal("native client did not retain exactly the public view")
				}
			}
			if service.calls != 2 {
				t.Fatalf("expected one endpoint invocation per call, got %d", service.calls)
			}
		})
	}
}

func (*rpcReceiptService) Book(_ context.Context, payload *genrpcreceipts.BookPayload) (*genrpcreceipts.OutcomeResult, error) {
	if payload.Destination != "Paris" {
		panic("native RPC decoder changed the domain input")
	}
	return &genrpcreceipts.OutcomeResult{Outcome: genrpcreceipts.NewOutcomeComplete(&genrpcreceipts.Receipt{Reference: "booked", Internal: "must-not-leak"})}, nil
}

func TestNativeJSONRPCUnionBranchView(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
	id, err := genrpccli.EncodeBookRequest(goahttp.RequestEncoder)(request, &genrpcreceipts.BookPayload{Destination: "Paris"})
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := request.Body.Close(); err != nil {
		t.Fatal(err)
	}
	server := genrpcsrv.New(genrpcreceipts.NewEndpoints(&rpcReceiptService{}), goahttp.NewMuxer(), goahttp.RequestDecoder, goahttp.ResponseEncoder, func(_ context.Context, _ http.ResponseWriter, err error) { t.Errorf("RPC failed: %v", err) })
	response := httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(body)))
	if response.Code != http.StatusOK || bytes.Contains(response.Body.Bytes(), []byte(`"internal"`)) || !bytes.Contains(response.Body.Bytes(), []byte(`"reference":"booked"`)) {
		t.Fatalf("RPC selected branch view: %d %s", response.Code, response.Body.String())
	}
	message := response.Result()
	defer func() {
		if err := message.Body.Close(); err != nil {
			t.Error(err)
		}
	}()
	message.Request = request
	result, err := genrpccli.DecodeBookResponse(goahttp.ResponseDecoder, false)(message, id)
	if err != nil {
		t.Fatal(err)
	}
	client := genrpcreceipts.NewClient(func(context.Context, any) (any, error) { return result, nil })
	decoded, err := client.Book(t.Context(), &genrpcreceipts.BookPayload{Destination: "Paris"})
	if err != nil {
		t.Fatal(err)
	}
	branch, ok := decoded.Outcome.AsComplete()
	if !ok || branch.Reference != "booked" || branch.Internal != "" {
		t.Fatal("RPC client did not retain exactly the public branch view")
	}
}

func TestSelectedUnionViewValidation(t *testing.T) {
	for _, test := range []struct {
		name    string
		invalid bool
	}{
		{"invalid_public", true},
		{"private_empty", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			endpoint := genreceipts.NewEndpoints(&receiptService{branch: test.name}).Book
			_, err := endpoint(t.Context(), &genreceipts.BookPayload{Destination: "Paris"})
			if (err != nil) != test.invalid {
				t.Errorf("selected branch validation: %v, want invalid=%t", err, test.invalid)
			}
		})
	}
}
