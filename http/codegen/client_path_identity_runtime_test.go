// This file generates HTTP clients and servers and checks that string resource
// names survive normal and catch-all routes without changing their contents.
package codegen

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	goacodegen "goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestGeneratedPathIdentityRequestBuilder exercises generated requests over
// HTTP and WebSocket connections while preserving the public path helpers.
func TestGeneratedPathIdentityRequestBuilder(t *testing.T) {
	root := expr.RunDSL(t, pathIdentityDSL)
	generation, err := goacodegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, httpPlans[0].Link())

	serviceFiles, err := service.Files(servicePlan)
	require.NoError(t, err)
	files := slices.Clone(serviceFiles)
	files = append(files, httpPlans[0].ClientFiles()...)
	files = append(files, httpPlans[0].ClientTypeFiles()...)
	files = append(files, httpPlans[0].PathFiles()...)
	files = append(files, httpPlans[0].ServerFiles()...)
	files = append(files, httpPlans[0].ServerTypeFiles()...)
	runPathIdentityRuntimeTest(t, files)
}

// pathIdentityDSL gives ordinary and catch-all string routes the same service
// behavior so the runtime test can compare transport handling directly.
func pathIdentityDSL() {
	dsl.Service("path_identity", func() {
		dsl.Method("array", func() {
			dsl.Payload(dsl.ArrayOf(dsl.String))
			dsl.Result(dsl.ArrayOf(dsl.String))
			dsl.HTTP(func() {
				dsl.GET("/array/{names}")
				dsl.Param("names")
			})
		})
		for _, method := range []string{"single", "catch_all", "socket"} {
			dsl.Method(method, func() {
				dsl.Payload(dsl.String)
				if method == "socket" {
					dsl.StreamingResult(dsl.String)
				} else {
					dsl.Result(dsl.String)
				}
				dsl.HTTP(func() {
					if method == "catch_all" {
						dsl.GET("/catch/{*name}")
					} else {
						dsl.GET("/" + method + "/{name}")
					}
					dsl.Param("name")
				})
			})
		}
	})
}

// runPathIdentityRuntimeTest writes the generated client and runs its test in
// an isolated module that uses this Goa checkout.
func runPathIdentityRuntimeTest(t *testing.T, files []*goacodegen.File) {
	t.Helper()
	directory := t.TempDir()
	writeHTTPGeneratedModule(t, directory)
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	testPath := filepath.Join(directory, "gen", "http", "path_identity", "client", "path_identity_test.go")
	require.NoError(t, os.WriteFile(testPath, []byte(generatedPathIdentityTest), 0o600))

	parallelHTTPTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=readonly", "./gen/http/path_identity/client")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "run generated path identity test:\n%s", output)
}

const generatedPathIdentityTest = `package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	genserver "generated.local/gen/http/path_identity/server"
	genservice "generated.local/gen/path_identity"
	"github.com/gorilla/websocket"
	goahttp "goa.design/goa/v3/http"
)

func TestPathIdentity(t *testing.T) {
	mux := goahttp.NewMuxer()
	echo := func(ctx context.Context, value any) (any, error) {
		return value, nil
	}
	endpoints := &genservice.Endpoints{Single: echo, CatchAll: echo, Array: echo}
	server := genserver.New(endpoints, mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, nil, nil, nil, nil)
	server.Mount(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()
	address, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(address.Scheme, address.Host, ts.Client(), goahttp.RequestEncoder, goahttp.ResponseDecoder, false, nil, nil)
	for _, name := range []string{"ordinary", "équipe", "operations/shift", "literal%2Fpercent", "literal%percent", "what?now", "fragment#name", ".", ".."} {
		t.Run(name, func(t *testing.T) {
			for _, call := range []func(context.Context, any) (any, error){client.Single(), client.CatchAll()} {
				got, err := call(context.Background(), name)
				if err != nil || got != name {
					t.Errorf("got=%q err=%v want=%q", got, err, name)
				}
			}
			values := []string{name, "other"}
			got, err := client.Array()(context.Background(), values)
			if err != nil || !reflect.DeepEqual(got, values) {
				t.Errorf("array got=%q err=%v want=%q", got, err, values)
			}
			// This is the previous generated client's array wire representation.
			legacyURL := &url.URL{Scheme: address.Scheme, Host: address.Host, Path: "/array/" + url.QueryEscape(name) + ",other"}
			request, err := client.BuildArrayRequest(context.Background(), values)
			if err != nil {
				t.Fatal(err)
			}
			if request.URL.String() != legacyURL.String() {
				t.Errorf("array wire changed: %q want %q", request.URL.String(), legacyURL.String())
			}
			if got := ArrayPathIdentityPath(values); got != "/array/"+strings.Join([]string{url.QueryEscape(name), "other"}, ",") {
				t.Errorf("array helper changed: %q", got)
			}
			if got := SinglePathIdentityPath(name); got != "/single/"+name {
				t.Errorf("public helper changed: %q", got)
			}
		})
	}
}

func TestSocketPathIdentity(t *testing.T) {
	mux := goahttp.NewMuxer()
	upgrader := websocket.Upgrader{}
	mux.Handle("GET", "/socket/{name}", func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if err := connection.WriteMessage(websocket.TextMessage, []byte(mux.Vars(r)["name"])); err != nil {
			t.Error(err)
		}
		if err := connection.Close(); err != nil {
			t.Error(err)
		}
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()
	address, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(address.Scheme, address.Host, ts.Client(), nil, nil, false, nil, nil)
	for _, name := range []string{"équipe", "operations/shift", "literal%2Fpercent"} {
		t.Run(name, func(t *testing.T) {
			request, err := client.BuildSocketRequest(context.Background(), name)
			if err != nil {
				t.Fatal(err)
			}
			connection, response, err := websocket.DefaultDialer.Dial(request.URL.String(), nil)
			if response != nil && response.Body != nil {
				if err := response.Body.Close(); err != nil {
					t.Error(err)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			_, got, err := connection.ReadMessage()
			if err != nil || string(got) != name {
				t.Errorf("got=%q err=%v want=%q", got, err, name)
			}
			if err := connection.Close(); err != nil {
				t.Error(err)
			}
		})
	}
}
`
