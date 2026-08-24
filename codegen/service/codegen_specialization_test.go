// This file verifies that service generation omits runtime work whose answer
// is fixed by the evaluated design.
package service

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
)

// TestInterceptorAccessorsDoNotRediscoverPlannedMethods catches generated
// accessors that switch on the method name or the planned payload wrapper.
func TestInterceptorAccessorsDoNotRediscoverPlannedMethods(t *testing.T) {
	root := codegen.RunDSL(t, interceptorSpecializationDSL)
	plan := retainedServicePlanForPackage(t, root)
	files := interceptorsFiles(plan, plan.facts.services[0])

	var rendered strings.Builder
	for _, file := range files {
		var source bytes.Buffer
		for _, section := range file.SectionTemplates {
			require.NoError(t, section.Write(&source))
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), file.Path, source.Bytes(), 0)
		require.NoError(t, err, source.String())
		ast.Inspect(parsed, func(node ast.Node) bool {
			switch node.(type) {
			case *ast.SwitchStmt, *ast.TypeSwitchStmt:
				t.Errorf("%s contains a runtime switch for a planned interceptor fact", file.Path)
			}
			return true
		})
		rendered.Write(source.Bytes())
	}

	code := rendered.String()
	require.Contains(t, code, "InspectInfo interface")
	require.NotContains(t, code, "method     string")
	require.NotContains(t, code, "callType   goa.InterceptorCallType")

	generated, err := Files(plan)
	require.NoError(t, err)
	compileGeneratedServiceFilesWith(t, generated, map[string]string{
		"gen/interceptor_specialization/info_specialization_test.go": interceptorInfoRuntimeTest,
	})
}

// TestSharedInterceptorSpecializesDifferentClientAndServerMethods catches a
// client method implementation lost when the same interceptor is also used by
// a different server method.
func TestSharedInterceptorSpecializesDifferentClientAndServerMethods(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.Interceptor("inspect", func() {
			dsl.ReadPayload(func() {
				dsl.Attribute("value")
			})
		})
		dsl.Service("SplitInterceptorMethods", func() {
			dsl.Method("ServerOnly", func() {
				dsl.ServerInterceptor("inspect")
				dsl.Payload(func() {
					dsl.Attribute("value", dsl.String)
				})
			})
			dsl.Method("ClientOnly", func() {
				dsl.ClientInterceptor("inspect")
				dsl.Payload(func() {
					dsl.Attribute("value", dsl.String)
				})
			})
		})
	})
	plan := retainedServicePlanForPackage(t, root)
	files, err := Files(plan)
	require.NoError(t, err)
	compileGeneratedServiceFiles(t, files)
}

// TestEmptyProjectedValidatorsAreOmitted catches public validation functions
// and parent calls that cannot report an error for any value.
func TestEmptyProjectedValidatorsAreOmitted(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		empty := dsl.ResultType("application/vnd.empty", func() {
			dsl.TypeName("Empty")
			dsl.Attribute("name", dsl.String)
			dsl.View("default", func() {
				dsl.Attribute("name")
			})
		})
		dsl.Service("EmptyViews", func() {
			dsl.Method("Read", func() {
				dsl.Result(empty)
			})
		})
	})
	plan := retainedServicePlanForPackage(t, root)
	data := plan.Services().Get("EmptyViews")

	require.Len(t, data.projectedTypes, 1)
	require.Empty(t, data.projectedTypes[0].Validations)
	require.Len(t, data.viewedResultTypes, 1)
	require.Len(t, data.viewedResultTypes[0].Validate.Calls, 1)
	require.Nil(t, data.viewedResultTypes[0].Validate.Calls[0].Declaration)

	files, err := Files(plan)
	require.NoError(t, err)
	compileGeneratedServiceFiles(t, files)
}

// TestRequiredParentOmitsEmptyChildCall catches removal of the parent's
// missing-field check when its selected child view has no other rules.
func TestRequiredParentOmitsEmptyChildCall(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		child := dsl.ResultType("application/vnd.empty-child", func() {
			dsl.TypeName("EmptyChild")
			dsl.Attribute("name", dsl.String)
			dsl.View("default", func() {
				dsl.Attribute("name")
			})
		})
		parent := dsl.ResultType("application/vnd.required-parent", func() {
			dsl.TypeName("RequiredParent")
			dsl.Attribute("child", child)
			dsl.Required("child")
			dsl.View("default", func() {
				dsl.Attribute("child")
			})
		})
		dsl.Service("RequiredParentViews", func() {
			dsl.Method("Read", func() {
				dsl.Result(parent)
			})
		})
	})
	plan := retainedServicePlanForPackage(t, root)
	data := plan.Services().Get("RequiredParentViews")

	var parentValidation *ValidateData
	for _, projected := range data.projectedTypes {
		switch projected.Name {
		case "EmptyChildView":
			require.Empty(t, projected.Validations)
		case "RequiredParentView":
			require.Len(t, projected.Validations, 1)
			parentValidation = projected.Validations[0]
		}
	}
	require.NotNil(t, parentValidation)
	require.Empty(t, parentValidation.Calls)
	require.Contains(t, parentValidation.Validate, `MissingFieldError("child", "result")`)

	files, err := Files(plan)
	require.NoError(t, err)
	compileGeneratedServiceFiles(t, files)
}

// TestEmptyRecursiveProjectedValidatorsAreOmitted catches cycles that retain
// validators even though no node in the cycle can report an error.
func TestEmptyRecursiveProjectedValidatorsAreOmitted(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		tree := dsl.ResultType("application/vnd.empty-tree", func() {
			dsl.TypeName("EmptyTree")
			dsl.Attribute("next", "EmptyTree")
			dsl.View("default", func() {
				dsl.Attribute("next")
			})
		})
		dsl.Service("EmptyRecursiveViews", func() {
			dsl.Method("Read", func() {
				dsl.Result(tree)
			})
		})
	})
	plan := retainedServicePlanForPackage(t, root)
	data := plan.Services().Get("EmptyRecursiveViews")

	require.Len(t, data.projectedTypes, 1)
	require.Empty(t, data.projectedTypes[0].Validations)

	files, err := Files(plan)
	require.NoError(t, err)
	compileGeneratedServiceFiles(t, files)
}

// interceptorSpecializationDSL applies one interceptor to multiple methods so
// generated accessors must select exact method and streaming types in advance.
func interceptorSpecializationDSL() {
	dsl.Interceptor("inspect", func() {
		dsl.ReadPayload(func() {
			dsl.Attribute("initial")
		})
		dsl.ReadStreamingPayload(func() {
			dsl.Attribute("input")
		})
		dsl.ReadStreamingResult(func() {
			dsl.Attribute("output")
		})
	})
	dsl.Service("InterceptorSpecialization", func() {
		dsl.ServerInterceptor("inspect")
		dsl.ClientInterceptor("inspect")
		for _, name := range []string{"First", "Second"} {
			dsl.Method(name, func() {
				dsl.Payload(func() {
					dsl.Field(1, "initial", dsl.String)
				})
				dsl.StreamingPayload(func() {
					dsl.Field(1, "input", dsl.String)
				})
				dsl.StreamingResult(func() {
					dsl.Field(1, "output", dsl.String)
				})
				dsl.GRPC(func() {})
			})
		}
	})
}

const interceptorInfoRuntimeTest = `package interceptorspecialization

import (
	"testing"

	goa "goa.design/goa/v3/pkg"
)

func TestSpecializedInterceptorInfo(t *testing.T) {
	initial := "start"
	input := "in"
	output := "out"
	payload := &FirstPayload{Initial: &initial}
	streamingPayload := &FirstStreamingPayload{Input: &input}
	streamingResult := &FirstResult{Output: &output}

	server := &inspectFirstServerUnaryInfo{inspectFirstInfo: &inspectFirstInfo{
		rawPayload: &FirstEndpointInput{Payload: payload},
	}}
	if server.Service() != "InterceptorSpecialization" || server.Method() != "First" || server.CallType() != goa.InterceptorUnary {
		t.Errorf("unexpected server metadata: %s %s %v", server.Service(), server.Method(), server.CallType())
	}
	if actual := server.Payload().Initial(); actual != initial {
		t.Errorf("server payload = %q, want %q", actual, initial)
	}

	client := &inspectFirstClientUnaryInfo{inspectFirstInfo: &inspectFirstInfo{rawPayload: payload}}
	if client.CallType() != goa.InterceptorUnary || client.Payload().Initial() != initial {
		t.Errorf("unexpected client endpoint metadata")
	}

	send := &inspectFirstStreamingSendInfo{inspectFirstInfo: &inspectFirstInfo{rawPayload: streamingResult}}
	if send.CallType() != goa.InterceptorStreamingSend || send.ServerStreamingResult().Output() != output {
		t.Errorf("unexpected server send metadata")
	}

	recv := &inspectFirstStreamingRecvInfo{inspectFirstInfo: &inspectFirstInfo{}}
	if recv.CallType() != goa.InterceptorStreamingRecv || recv.ServerStreamingPayload(streamingPayload).Input() != input {
		t.Errorf("unexpected server receive metadata")
	}

	clientSend := &inspectFirstStreamingSendInfo{inspectFirstInfo: &inspectFirstInfo{rawPayload: streamingPayload}}
	if clientSend.ClientStreamingPayload().Input() != input {
		t.Errorf("unexpected client send metadata")
	}
	clientRecv := &inspectFirstStreamingRecvInfo{inspectFirstInfo: &inspectFirstInfo{}}
	if clientRecv.ClientStreamingResult(streamingResult).Output() != output {
		t.Errorf("unexpected client receive metadata")
	}
}
`
