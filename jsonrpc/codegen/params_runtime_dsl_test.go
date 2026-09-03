// This file declares the JSON-RPC request and stream shapes used by the
// generated runtime tests and runs their generated client and server packages.
package codegen_test

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
)

// paramsRuntimeDSL declares the request and stream shapes exercised by the
// generated runtime tests.
func paramsRuntimeDSL() {
	alias := dsl.Type("Alias", dsl.String, func() {
		dsl.Pattern("^[a-z]+$")
	})
	eventID := dsl.Type("EventID", dsl.String, func() {
		dsl.Pattern("^[a-z]*$")
	})
	eventType := dsl.Type("EventType", dsl.String, func() {
		dsl.Enum("", "update")
	})
	retry := dsl.Type("Retry", dsl.UInt32, func() {
		dsl.Maximum(10)
	})
	defaultEventID := dsl.Type("DefaultEventID", dsl.String, func() {
		dsl.Default("fallback")
	})
	defaultEventType := dsl.Type("DefaultEventType", dsl.String, func() {
		dsl.Default("update")
	})
	defaultRetry := dsl.Type("DefaultRetry", dsl.UInt32, func() {
		dsl.Default(0)
	})
	defaultStreamData := dsl.Type("DefaultStreamData", dsl.String, func() {
		dsl.Pattern("^[a-z]*$")
		dsl.Default("fallback")
	})
	resumeID := dsl.Type("ResumeID", dsl.String, func() {
		dsl.Default("cursor")
	})
	object := dsl.Type("Object", func() {
		dsl.Attribute("name", dsl.String)
		dsl.Required("name")
	})
	objects := dsl.Type("Objects", dsl.ArrayOf(object))
	defaultValues := dsl.Type("DefaultValues", dsl.ArrayOf(dsl.String))
	objectMap := dsl.Type("ObjectMap", dsl.MapOf(dsl.String, object))
	empty := dsl.Type("Empty", func() {})
	details := dsl.Type("Details", func() {
		dsl.Attribute("name", dsl.String)
		dsl.Required("name")
	})
	aliasEvent := dsl.Type("AliasEvent", func() {
		dsl.Attribute("data", alias)
		dsl.Attribute("event_id", eventID)
		dsl.Attribute("event_type", eventType)
		dsl.Attribute("retry", retry)
		dsl.Required("data")
	})
	objectEvent := dsl.Type("ObjectEvent", func() {
		dsl.Attribute("data", details)
		dsl.Attribute("event_id", dsl.String)
		dsl.Attribute("event_type", dsl.String)
		dsl.Attribute("retry", dsl.Int64)
		dsl.Required("data", "event_id", "event_type", "retry")
	})
	defaultEvent := dsl.Type("DefaultEvent", func() {
		dsl.Attribute("data", alias)
		dsl.Attribute("event_id", defaultEventID)
		dsl.Attribute("event_type", defaultEventType)
		dsl.Attribute("retry", defaultRetry)
		dsl.Required("data")
	})
	optionalTextEvent := dsl.Type("OptionalTextEvent", func() {
		dsl.Attribute("data", dsl.String, func() {
			dsl.Pattern("^[a-z]*$")
		})
	})
	defaultTextEvent := dsl.Type("DefaultTextEvent", func() {
		dsl.Attribute("data", defaultStreamData)
	})
	optionalObjectEvent := dsl.Type("OptionalObjectEvent", func() {
		dsl.Attribute("data", object)
	})
	optionalArrayEvent := dsl.Type("OptionalArrayEvent", func() {
		dsl.Attribute("data", objects)
	})
	optionalMapEvent := dsl.Type("OptionalMapEvent", func() {
		dsl.Attribute("data", objectMap)
	})
	optionalUnionEvent := dsl.Type("OptionalUnionEvent", func() {
		dsl.OneOf("data", func() {
			dsl.TypeName("OptionalStreamData")
			dsl.Field(1, "name", dsl.String)
			dsl.Field(2, "count", dsl.Int)
			dsl.Field(3, "inactive", empty)
		})
	})
	dsl.Service("Param Shapes", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/rpc")
		})
		for _, method := range []struct {
			name  string
			type_ any
		}{
			{name: "text", type_: dsl.String},
			{name: "alias", type_: alias},
			{name: "anything", type_: dsl.Any},
			{name: "bytes", type_: dsl.Bytes},
			{name: "object", type_: object},
			{name: "array", type_: dsl.ArrayOf(dsl.String)},
			{name: "map", type_: dsl.MapOf(dsl.String, dsl.Int)},
		} {
			dsl.Method(method.name, func() {
				dsl.Payload(method.type_)
				dsl.JSONRPC(func() {})
			})
		}
		dsl.Method("optional_text", func() {
			dsl.Payload(func() {
				dsl.Attribute("value", dsl.String, func() {
					dsl.Pattern("^[a-z]*$")
				})
			})
			dsl.JSONRPC(func() {
				dsl.Body("value")
			})
		})
		dsl.Method("default_text", func() {
			dsl.Payload(func() {
				dsl.Attribute("value", defaultStreamData)
			})
			dsl.JSONRPC(func() {
				dsl.Body("value")
			})
		})
		dsl.Method("default_array", func() {
			dsl.Payload(func() {
				dsl.Attribute("value", defaultValues, func() {
					dsl.Default([]string{"fallback"})
				})
			})
			dsl.JSONRPC(func() {
				dsl.Body("value")
			})
		})
		for _, method := range []struct {
			name  string
			type_ any
		}{
			{name: "optional_object", type_: object},
			{name: "optional_array", type_: objects},
			{name: "optional_map", type_: objectMap},
		} {
			dsl.Method(method.name, func() {
				dsl.Payload(func() {
					dsl.Attribute("value", method.type_)
				})
				dsl.JSONRPC(func() {
					dsl.Body("value")
				})
			})
		}
		dsl.Method("optional_union", func() {
			dsl.Payload(func() {
				dsl.OneOf("value", func() {
					dsl.TypeName("OptionalRequestValue")
					dsl.Field(1, "name", dsl.String)
					dsl.Field(2, "count", dsl.Int)
					dsl.Field(3, "inactive", empty)
				})
			})
			dsl.JSONRPC(func() {
				dsl.Body("value")
			})
		})
		dsl.Method("stream_text", func() {
			dsl.StreamingResult(dsl.String)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents()
			})
		})
		dsl.Method("stream_any", func() {
			dsl.StreamingResult(dsl.Any)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents()
			})
		})
		dsl.Method("stream_array", func() {
			dsl.StreamingResult(dsl.ArrayOf(dsl.String))
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents()
			})
		})
		dsl.Method("stream_alias_fields", func() {
			dsl.StreamingResult(aliasEvent)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {
					dsl.SSEEventData("data")
					dsl.SSEEventID("event_id")
					dsl.SSEEventType("event_type")
					dsl.SSEEventRetry("retry")
				})
			})
		})
		dsl.Method("stream_object_fields", func() {
			dsl.StreamingResult(objectEvent)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {
					dsl.SSEEventData("data")
					dsl.SSEEventID("event_id")
					dsl.SSEEventType("event_type")
					dsl.SSEEventRetry("retry")
				})
			})
		})
		dsl.Method("stream_default_fields", func() {
			dsl.StreamingResult(defaultEvent)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {
					dsl.SSEEventData("data")
					dsl.SSEEventID("event_id")
					dsl.SSEEventType("event_type")
					dsl.SSEEventRetry("retry")
				})
			})
		})
		dsl.Method("stream_optional_text", func() {
			dsl.StreamingResult(optionalTextEvent)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {
					dsl.SSEEventData("data")
				})
			})
		})
		dsl.Method("stream_default_text", func() {
			dsl.StreamingResult(defaultTextEvent)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {
					dsl.SSEEventData("data")
				})
			})
		})
		for _, method := range []struct {
			name  string
			type_ any
		}{
			{name: "stream_optional_object", type_: optionalObjectEvent},
			{name: "stream_optional_array", type_: optionalArrayEvent},
			{name: "stream_optional_map", type_: optionalMapEvent},
			{name: "stream_optional_union", type_: optionalUnionEvent},
		} {
			dsl.Method(method.name, func() {
				dsl.StreamingResult(method.type_)
				dsl.JSONRPC(func() {
					dsl.ServerSentEvents(func() {
						dsl.SSEEventData("data")
					})
				})
			})
		}
		dsl.Method("required_resume", func() {
			dsl.Payload(func() {
				dsl.Attribute("last_event_id", dsl.String)
				dsl.Required("last_event_id")
			})
			dsl.StreamingResult(dsl.String)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {
					dsl.SSERequestID("last_event_id")
				})
			})
		})
		dsl.Method("optional_resume", func() {
			dsl.Payload(func() {
				dsl.Attribute("last_event_id", dsl.String)
			})
			dsl.StreamingResult(dsl.String)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {
					dsl.SSERequestID("last_event_id")
				})
			})
		})
		dsl.Method("default_resume", func() {
			dsl.Payload(func() {
				dsl.Attribute("last_event_id", resumeID)
			})
			dsl.StreamingResult(dsl.String)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {
					dsl.SSERequestID("last_event_id")
				})
			})
		})
	})
}

// runParamsRuntimeTests runs the generated package tests with a fixed timeout.
func runParamsRuntimeTests(t *testing.T, dir string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "./jsonrpc/param_shapes/client", "./jsonrpc/param_shapes/server")
	command.Dir = dir
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
}
