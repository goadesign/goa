// This fixture compiles generated HTTP client and server packages, then passes
// named scalar headers through their real codecs without an HTTP listener.
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

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestGeneratedNamedScalarHeaders(t *testing.T) {
	root := expr.RunDSL(t, namedScalarHeaderDSL)
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
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
	files = append(files, httpPlans[0].ServerFiles()...)
	files = append(files, httpPlans[0].ServerTypeFiles()...)
	files = append(files, httpPlans[0].PathFiles()...)

	directory := t.TempDir()
	repository, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	module := "module generated.local\n\ngo 1.26\n\n" +
		"require goa.design/goa/v3 v3.0.0\n\n" +
		"replace goa.design/goa/v3 => " + filepath.ToSlash(repository) + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(directory, "go.mod"), []byte(module), 0o600))
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	testPath := filepath.Join(directory, "gen", "http", "named_headers", "client", "named_scalar_header_test.go")
	require.NoError(t, os.WriteFile(testPath, []byte(generatedNamedScalarHeaderTest), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "-count=1", "./gen/http/named_headers/client")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off", "GOPROXY=off", "GOSUMDB=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "run generated named scalar header test:\n%s", output)
}

// namedScalarHeaderDSL reproduces RequestID -> UUID -> string in a shared Go
// package. The original service declaration must stay typed after HTTP shaping.
func namedScalarHeaderDSL() {
	uuid := dsl.Type("UUID", dsl.String, func() {
		dsl.Pattern("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")
		dsl.Meta("struct:pkg:path", "types")
	})
	requestID := dsl.Type("RequestID", uuid, func() {
		dsl.Meta("struct:pkg:path", "types")
	})
	restartID := dsl.Type("RestartID", requestID, func() {
		dsl.Meta("struct:pkg:path", "types")
	})
	number := dsl.Type("Number", dsl.Int64, func() {
		dsl.Meta("struct:pkg:path", "types")
	})
	count := dsl.Type("Count", number, func() {
		dsl.Meta("struct:pkg:path", "types")
	})
	word := dsl.Type("Word", dsl.String, func() {
		dsl.Pattern("^[a-z]+$")
		dsl.MinLength(4)
		dsl.Meta("struct:pkg:path", "types")
	})
	label := dsl.Type("Label", word, func() {
		dsl.Meta("struct:pkg:path", "types")
	})
	dsl.Service("named_headers", func() {
		dsl.Method("create", func() {
			dsl.Payload(func() {
				dsl.Attribute("requestId", requestID)
				dsl.Attribute("optionalId", restartID)
				dsl.Attribute("plain", dsl.String)
				dsl.Attribute("oneLevel", uuid)
				dsl.Attribute("count", count)
				dsl.Attribute("label", label)
				dsl.Required("requestId", "plain", "oneLevel", "count", "label")
			})
			dsl.HTTP(func() {
				dsl.POST("/uploads")
				dsl.SkipRequestBodyEncodeDecode()
				dsl.Header("requestId:X-Request-ID")
				dsl.Header("optionalId:X-Optional-ID")
				dsl.Header("plain:X-Plain")
				dsl.Header("oneLevel:X-One-Level")
				dsl.Header("count:X-Count")
				dsl.Header("label:X-Label")
			})
		})
	})
}

const generatedNamedScalarHeaderTest = `package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	gennamedheaders "generated.local/gen/named_headers"
	genserver "generated.local/gen/http/named_headers/server"
	gentypes "generated.local/gen/types"
	goahttp "goa.design/goa/v3/http"
)

func TestNamedHeaderRoundTrip(t *testing.T) {
	for _, id := range []string{
		"AbCdEfAB-1234-0000-FFFF-0123456789aB",
		"00000000-0000-0000-0000-000000000000",
	} {
		for _, present := range []bool{false, true} {
			t.Run(id+"/optional_"+map[bool]string{false:"absent", true:"present"}[present], func(t *testing.T) {
				payload := namedHeaderPayload(id)
				if present {
					optional := gentypes.RestartID(id)
					payload.OptionalID = &optional
				}
				// A caller deliberately repeats one creation identity. Encoding
				// may not replace or normalize the ID on either attempt.
				for range 2 {
					body := io.NopCloser(strings.NewReader("unchanged raw bytes"))
					request := httptest.NewRequest(http.MethodPost, "/uploads", body)
					data := &gennamedheaders.CreateRequestData{Payload: payload, Body: body}
					require.NoError(t, EncodeCreateRequest(nil)(request, data))
					require.Equal(t, id, request.Header.Get("X-Request-ID"))
					require.Equal(t, id, request.Header.Get("X-One-Level"))
					require.Equal(t, "plain-value", request.Header.Get("X-Plain"))
					require.Equal(t, "65", request.Header.Get("X-Count"), "numeric aliases use decimal formatting, not rune conversion")
					if present {
						require.Equal(t, id, request.Header.Get("X-Optional-ID"))
					} else {
						require.NotContains(t, request.Header, http.CanonicalHeaderKey("X-Optional-ID"))
					}
					actual, err := genserver.DecodeCreateRequest(goahttp.NewMuxer(), goahttp.RequestDecoder)(request)
					require.NoError(t, err)
					require.Equal(t, payload, actual)
					var typedID gentypes.RequestID = actual.RequestID
					require.Equal(t, id, string(typedID))
					remaining, err := io.ReadAll(request.Body)
					require.NoError(t, err)
					require.Equal(t, "unchanged raw bytes", string(remaining))
					require.NoError(t, request.Body.Close())
				}
			})
		}
	}
}

func TestNamedHeaderGeneratedRejection(t *testing.T) {
	for _, test := range []struct {
		name, header, value, field string
		missing bool
	}{
		{name:"missing request ID", header:"X-Request-ID", field:"requestId", missing:true},
		{name:"empty request ID", header:"X-Request-ID", value:"", field:"requestId"},
		{name:"malformed request ID", header:"X-Request-ID", value:"not-a-uuid", field:"requestId"},
		{name:"whitespace is not normalized", header:"X-Request-ID", value:" 00000000-0000-0000-0000-000000000000 ", field:"requestId"},
		{name:"malformed third-level ID", header:"X-Optional-ID", value:"not-a-uuid", field:"optionalId"},
		{name:"base pattern", header:"X-Label", value:"word7", field:"label"},
		{name:"inherited length", header:"X-Label", value:"abc", field:"label"},
		{name:"invalid numeric value", header:"X-Count", value:"sixty-five", field:"count"},
	} {
		t.Run(test.name, func(t *testing.T) {
			payload := namedHeaderPayload("00000000-0000-0000-0000-000000000000")
			body := io.NopCloser(strings.NewReader(""))
			request := httptest.NewRequest(http.MethodPost, "/uploads", body)
			data := &gennamedheaders.CreateRequestData{Payload: payload, Body: body}
			require.NoError(t, EncodeCreateRequest(nil)(request, data))
			if test.missing {
				request.Header.Del(test.header)
			} else {
				request.Header.Set(test.header, test.value)
			}
			_, err := genserver.DecodeCreateRequest(goahttp.NewMuxer(), goahttp.RequestDecoder)(request)
			require.Error(t, err)
			require.Contains(t, err.Error(), test.field)
			require.NoError(t, request.Body.Close())
		})
	}
}

func namedHeaderPayload(id string) *gennamedheaders.CreatePayload {
	return &gennamedheaders.CreatePayload{
		RequestID: gentypes.RequestID(id),
		Plain: "plain-value",
		OneLevel: gentypes.UUID(id),
		Count: gentypes.Count(65),
		Label: gentypes.Label("word"),
	}
}
`
