package codegen

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

// renderFileToString renders all sections of a file to a string without writing to disk
func renderFileToString(file *codegen.File) (string, error) {
	var buf bytes.Buffer
	for _, section := range file.SectionTemplates {
		if err := section.Write(&buf); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

// TestWebSocketGoldenFiles tests WebSocket code generation against golden files
// to ensure comprehensive coverage of all WebSocket templates and scenarios.
func TestWebSocketGoldenFiles(t *testing.T) {
	cases := []struct {
		name     string
		dsl      func()
		fileType string // "server" or "client"
	}{
		// Server Streaming (Result only)
		{"websocket-server-streaming-primitive", serverStreamingPrimitiveDSL, "server"},
		{"websocket-server-streaming-array", serverStreamingArrayDSL, "server"},
		{"websocket-server-streaming-object", serverStreamingObjectDSL, "server"},
		{"websocket-server-streaming-user-type", serverStreamingUserTypeDSL, "server"},
		{"websocket-server-streaming-with-views", serverStreamingWithViewsDSL, "server"},
		{"websocket-server-streaming-with-views-client", serverStreamingWithViewsDSL, "client"},
		{"websocket-server-streaming-collection-with-views-client", testdata.StreamingResultCollectionWithViewsDSL, "client"},
		{"websocket-server-streaming-with-explicit-body-client", serverStreamingWithExplicitBodyDSL, "client"},

		// Client Streaming (Payload only)
		{"websocket-client-streaming-primitive", clientStreamingPrimitiveDSL, "client"},
		{"websocket-client-streaming-array", clientStreamingArrayDSL, "client"},
		{"websocket-client-streaming-object", clientStreamingObjectDSL, "client"},
		{"websocket-client-streaming-user-type", clientStreamingUserTypeDSL, "client"},
		{"websocket-client-streaming-with-validation", clientStreamingWithValidationDSL, "client"},

		// Bidirectional Streaming
		{"websocket-bidirectional-streaming-primitive", bidirectionalStreamingPrimitiveDSL, "server"},
		{"websocket-bidirectional-streaming-complex", bidirectionalStreamingComplexDSL, "server"},
		{"websocket-bidirectional-streaming-with-views", bidirectionalStreamingWithViewsDSL, "server"},

		// Client-side Bidirectional
		{"websocket-bidirectional-streaming-primitive-client", bidirectionalStreamingPrimitiveDSL, "client"},
		{"websocket-bidirectional-streaming-complex-client", bidirectionalStreamingComplexDSL, "client"},
		{"websocket-bidirectional-streaming-with-views-client", bidirectionalStreamingWithViewsDSL, "client"},

		// Edge Cases
		{"websocket-no-payload-streaming", noPayloadStreamingDSL, "server"},
		{"websocket-no-result-streaming", noResultStreamingDSL, "server"},
		{"websocket-mixed-endpoints", mixedEndpointsDSL, "server"},
		{"websocket-mixed-endpoints-client", mixedEndpointsDSL, "client"},

		// Template Coverage Tests
		{"websocket-conn-configurer", connConfigurerDSL, "server"},
		{"websocket-conn-configurer-client", connConfigurerDSL, "client"},
		{"websocket-struct-types", structTypesDSL, "server"},
		{"websocket-struct-types-client", structTypesDSL, "client"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := expr.RunDSL(t, c.dsl)
			plan := linkedHTTPPlanForRoot(t, root)

			var files []*codegen.File
			if c.fileType == "server" {
				files = plan.ServerFiles()
			} else {
				files = plan.ClientFiles()
			}

			// Find the websocket.go file
			var wsFile *codegen.File
			for _, f := range files {
				if filepath.Base(f.Path) == "websocket.go" {
					wsFile = f
					break
				}
			}

			if wsFile == nil {
				t.Skip("No WebSocket file generated for this test case")
				return
			}

			code, err := renderFileToString(wsFile)
			require.NoError(t, err)

			golden := filepath.Join("testdata", "golden", "websocket", c.name+".golden")

			// Create golden file if it doesn't exist (for initial creation)
			if _, err := os.Stat(golden); os.IsNotExist(err) {
				dir := filepath.Dir(golden)
				require.NoError(t, os.MkdirAll(dir, 0750))
				require.NoError(t, os.WriteFile(golden, []byte(code), 0644))
				t.Logf("Created golden file: %s", golden)
				return
			}

			// Use testutil for proper line ending normalization
			testutil.AssertString(t, golden, code)
		})
	}
}

// TestWebSocketTemplateExercise ensures all WebSocket templates are exercised
func TestWebSocketTemplateExercise(t *testing.T) {
	// Run a comprehensive test that should exercise all templates
	root := expr.RunDSL(t, comprehensiveWebSocketDSL)
	plan := linkedHTTPPlanForRoot(t, root)

	// Generate both server and client files
	serverFiles := plan.ServerFiles()
	clientFiles := plan.ClientFiles()

	// Verify WebSocket files were generated
	var serverWSFile, clientWSFile *codegen.File
	for _, f := range serverFiles {
		if filepath.Base(f.Path) == "websocket.go" {
			serverWSFile = f
			break
		}
	}
	for _, f := range clientFiles {
		if filepath.Base(f.Path) == "websocket.go" {
			clientWSFile = f
			break
		}
	}

	require.NotNil(t, serverWSFile, "Server WebSocket file should be generated")
	require.NotNil(t, clientWSFile, "Client WebSocket file should be generated")

	// Render the files to ensure all templates work
	_, err := renderFileToString(serverWSFile)
	require.NoError(t, err, "Server WebSocket file should render without error")

	_, err = renderFileToString(clientWSFile)
	require.NoError(t, err, "Client WebSocket file should render without error")
}

// DSL definitions for comprehensive WebSocket testing

func serverStreamingPrimitiveDSL() {
	Service("TestService", func() {
		Method("StreamPrimitive", func() {
			StreamingResult(String)
			HTTP(func() {
				GET("/stream/primitive")
			})
		})
	})
}

func serverStreamingArrayDSL() {
	Service("TestService", func() {
		Method("StreamArray", func() {
			StreamingResult(ArrayOf(String))
			HTTP(func() {
				GET("/stream/array")
			})
		})
	})
}

func serverStreamingObjectDSL() {
	Service("TestService", func() {
		Method("StreamObject", func() {
			StreamingResult(func() {
				Attribute("id", String)
				Attribute("name", String)
				Required("id")
			})
			HTTP(func() {
				GET("/stream/object")
			})
		})
	})
}

func serverStreamingUserTypeDSL() {
	var UserType = Type("User", func() {
		Attribute("id", String)
		Attribute("name", String)
		Attribute("email", String, func() {
			Format("email")
		})
		Required("id", "name")
	})

	Service("TestService", func() {
		Method("StreamUser", func() {
			StreamingResult(UserType)
			HTTP(func() {
				GET("/stream/user")
			})
		})
	})
}

func serverStreamingWithViewsDSL() {
	var UserType = ResultType("User", func() {
		Attributes(func() {
			Attribute("id", String)
			Attribute("name", String)
			Attribute("email", String)
		})
		Required("id", "name")
		View("default", func() {
			Attribute("id")
			Attribute("name")
			Attribute("email")
		})
		View("tiny", func() {
			Attribute("id")
		})
	})

	Service("TestService", func() {
		Method("StreamUserWithViews", func() {
			StreamingResult(UserType)
			HTTP(func() {
				GET("/stream/user/views")
			})
		})
	})
}

func serverStreamingWithExplicitBodyDSL() {
	var UserType = ResultType("User", func() {
		Attribute("id", String, func() {
			MinLength(3)
		})
		Attribute("name", String)
		Required("id", "name")
		View("default", func() {
			Attribute("id")
			Attribute("name")
		})
		View("tiny", func() {
			Attribute("id")
		})
	})
	var StreamBody = Type("StreamBody", func() {
		Attribute("id", String, func() {
			MinLength(3)
		})
		Required("id")
	})

	Service("TestService", func() {
		Method("StreamUserID", func() {
			StreamingResult(UserType)
			HTTP(func() {
				GET("/stream/user/id")
				Response(func() {
					Body(StreamBody)
				})
			})
		})
	})
}

func clientStreamingPrimitiveDSL() {
	Service("TestService", func() {
		Method("ClientStreamPrimitive", func() {
			StreamingPayload(String)
			Result(String)
			HTTP(func() {
				GET("/client/stream/primitive")
			})
		})
	})
}

func clientStreamingArrayDSL() {
	Service("TestService", func() {
		Method("ClientStreamArray", func() {
			StreamingPayload(ArrayOf(Int))
			Result(String)
			HTTP(func() {
				GET("/client/stream/array")
			})
		})
	})
}

func clientStreamingObjectDSL() {
	Service("TestService", func() {
		Method("ClientStreamObject", func() {
			StreamingPayload(func() {
				Attribute("data", String)
				Attribute("timestamp", Int64)
				Required("data")
			})
			Result(String)
			HTTP(func() {
				GET("/client/stream/object")
			})
		})
	})
}

func clientStreamingUserTypeDSL() {
	var Message = Type("Message", func() {
		Attribute("content", String)
		Attribute("sender", String)
		Attribute("timestamp", Int64)
		Required("content", "sender")
	})

	Service("TestService", func() {
		Method("ClientStreamMessage", func() {
			StreamingPayload(Message)
			Result(String)
			HTTP(func() {
				GET("/client/stream/message")
			})
		})
	})
}

func clientStreamingWithValidationDSL() {
	Service("TestService", func() {
		Method("ClientStreamValidated", func() {
			StreamingPayload(func() {
				Attribute("value", String, func() {
					MinLength(1)
					MaxLength(100)
				})
				Attribute("count", Int, func() {
					Minimum(0)
					Maximum(1000)
				})
				Required("value")
			})
			Result(String)
			HTTP(func() {
				GET("/client/stream/validated")
			})
		})
	})
}

func bidirectionalStreamingPrimitiveDSL() {
	Service("TestService", func() {
		Method("BidirectionalPrimitive", func() {
			StreamingPayload(String)
			StreamingResult(String)
			HTTP(func() {
				GET("/bidirectional/primitive")
			})
		})
	})
}

func bidirectionalStreamingComplexDSL() {
	var Request = Type("Request", func() {
		Attribute("id", String)
		Attribute("data", String)
		Required("id", "data")
	})

	var Response = Type("Response", func() {
		Attribute("id", String)
		Attribute("result", String)
		Attribute("status", String)
		Required("id", "result")
	})

	Service("TestService", func() {
		Method("BidirectionalComplex", func() {
			StreamingPayload(Request)
			StreamingResult(Response)
			HTTP(func() {
				GET("/bidirectional/complex")
			})
		})
	})
}

func bidirectionalStreamingWithViewsDSL() {
	var Request = Type("Request", func() {
		Attribute("id", String)
		Attribute("data", String)
		Required("id", "data")
	})

	var Response = ResultType("Response", func() {
		Attributes(func() {
			Attribute("id", String)
			Attribute("result", String)
			Attribute("metadata", func() {
				Attribute("timestamp", Int64)
				Attribute("source", String)
			})
		})
		Required("id", "result")
		View("default", func() {
			Attribute("id")
			Attribute("result")
			Attribute("metadata")
		})
		View("minimal", func() {
			Attribute("id")
			Attribute("result")
		})
	})

	Service("TestService", func() {
		Method("BidirectionalWithViews", func() {
			StreamingPayload(Request)
			StreamingResult(Response, func() {
				View("minimal")
			})
			HTTP(func() {
				GET("/bidirectional/views")
			})
		})
	})
}

func noPayloadStreamingDSL() {
	Service("TestService", func() {
		Method("NoPayloadStream", func() {
			StreamingResult(String)
			HTTP(func() {
				GET("/stream/no-payload")
			})
		})
	})
}

func noResultStreamingDSL() {
	Service("TestService", func() {
		Method("NoResultStream", func() {
			StreamingPayload(String)
			HTTP(func() {
				GET("/stream/no-result")
			})
		})
	})
}

func mixedEndpointsDSL() {
	Service("TestService", func() {
		Method("RegularEndpoint", func() {
			Payload(func() {
				Attribute("data", String)
			})
			Result(String)
			HTTP(func() {
				POST("/regular")
			})
		})

		Method("StreamingEndpoint", func() {
			StreamingResult(String)
			HTTP(func() {
				GET("/streaming")
			})
		})
	})
}

func connConfigurerDSL() {
	Service("TestService", func() {
		Method("ConfigurableStream", func() {
			StreamingResult(String)
			HTTP(func() {
				GET("/configurable")
			})
		})
	})
}

func structTypesDSL() {
	Service("TestService", func() {
		Method("StructStream", func() {
			StreamingPayload(func() {
				Attribute("field1", String)
				Attribute("field2", Int)
			})
			StreamingResult(func() {
				Attribute("result1", String)
				Attribute("result2", Boolean)
			})
			HTTP(func() {
				GET("/struct")
			})
		})
	})
}

func comprehensiveWebSocketDSL() {
	var UserType = ResultType("User", func() {
		Attributes(func() {
			Attribute("id", String)
			Attribute("name", String)
			Attribute("email", String, func() {
				Format("email")
			})
		})
		Required("id", "name")
		View("default", func() {
			Attribute("id")
			Attribute("name")
			Attribute("email")
		})
		View("tiny", func() {
			Attribute("id")
		})
	})

	Service("TestService", func() {
		Method("ServerStreaming", func() {
			StreamingResult(UserType)
			HTTP(func() {
				GET("/server-stream")
			})
		})

		Method("ClientStreaming", func() {
			StreamingPayload(func() {
				Attribute("data", String, func() {
					MinLength(1)
				})
				Required("data")
			})
			Result(String)
			HTTP(func() {
				GET("/client-stream")
			})
		})

		Method("BidirectionalStreaming", func() {
			StreamingPayload(UserType)
			StreamingResult(UserType, func() {
				View("tiny")
			})
			HTTP(func() {
				GET("/bidirectional")
			})
		})

		Method("RegularMethod", func() {
			Payload(String)
			Result(String)
			HTTP(func() {
				POST("/regular")
			})
		})
	})
}
