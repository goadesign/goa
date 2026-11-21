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
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestHTTPUnionRejectsChangedNamedDescendant(t *testing.T) {
	leaf := wireCatalogType("Leaf", "leaf", "label", false)
	branch := wireCatalogType("Branch", "branch", "leaf", false)
	branch.Attribute().Find("leaf").Type = leaf
	original := &expr.AttributeExpr{Type: &expr.Union{
		TypeName: "Choice",
		Values: []*expr.NamedAttributeExpr{
			{Name: "branch", Attribute: &expr.AttributeExpr{Type: branch}},
		},
	}}
	copied := expr.DupAtt(original)
	copiedBranch := expr.AsUnion(copied.Type).Values[0].Attribute.Type.(expr.UserType)
	copiedLeaf := copiedBranch.Attribute().Find("leaf").Type.(expr.UserType)
	require.Same(t, leaf.Origin(), copiedLeaf.Origin())
	expr.AsObject(copiedLeaf).Set("count", &expr.AttributeExpr{Type: expr.Int})

	catalog, _ := testWireTypeCatalog(t)
	catalog.collect(original, wireResponseBody, wireTypePolicy{})
	catalog.collect(copied, wireResponseBody, wireTypePolicy{})
	require.ErrorContains(t, catalog.Declare(), `OneOf "Choice" produces different ResponseBody definitions`)
}

func TestHTTPUnionSharedGraphKeepsRoles(t *testing.T) {
	for _, requestFirst := range []bool{false, true} {
		t.Run(map[bool]string{false: "response first", true: "request first"}[requestFirst], func(t *testing.T) {
			branch := wireCatalogType("Branch", "branch", "label", false)
			choice := &expr.AttributeExpr{Type: &expr.Union{
				TypeName: "Choice",
				Values: []*expr.NamedAttributeExpr{
					{Name: "branch", Attribute: &expr.AttributeExpr{Type: branch}},
				},
			}}
			secondRequest := expr.DupAtt(choice)
			secondResponse := expr.DupAtt(choice)
			secondDetailed := expr.DupAtt(choice)
			catalog, generation := testWireTypeCatalog(t)
			request := wireTypePolicy{request: true, pointer: true}
			response := wireTypePolicy{}
			detailed := wireTypePolicy{view: "detailed"}
			collectRequest := func() {
				catalog.collect(choice, wireRequestBody, request)
				catalog.collect(secondRequest, wireRequestBody, request)
			}
			collectResponse := func() {
				catalog.collect(choice, wireResponseBody, response)
				catalog.collect(secondResponse, wireResponseBody, response)
			}
			if requestFirst {
				collectRequest()
				collectResponse()
			} else {
				collectResponse()
				collectRequest()
			}
			catalog.collect(choice, wireResponseBody, detailed)
			catalog.collect(secondDetailed, wireResponseBody, detailed)
			linkTestWireTypeCatalog(t, generation, catalog)
			branchAttribute := expr.AsUnion(choice.Type).Values[0].Attribute
			requestBranch := catalog.find(newWireTypeIdentity(branchAttribute, wireAttribute, request, wireTypePreferredName(branch, request)))
			responseBranch := catalog.find(newWireTypeIdentity(branchAttribute, wireAttribute, response, wireTypePreferredName(branch, response)))
			require.NotSame(t, requestBranch, responseBranch)
			for _, union := range catalog.unionTypes() {
				switch union.Name {
				case "ChoiceRequestBody":
					require.Equal(t, requestBranch.ref, union.Fields[0].FieldType)
				case "ChoiceResponseBody", "ChoiceDetailedResponseBody":
					require.Equal(t, responseBranch.ref, union.Fields[0].FieldType)
				default:
					t.Errorf("unexpected union %q", union.Name)
				}
			}
			for _, test := range []struct {
				use    wireUnionUse
				policy wireTypePolicy
				second *expr.AttributeExpr
				name   string
			}{
				{wireUnionUse{role: wireRequestBody}, request, secondRequest, "ChoiceRequestBody"},
				{wireUnionUse{role: wireResponseBody}, response, secondResponse, "ChoiceResponseBody"},
				{wireUnionUse{role: wireResponseBody, view: "detailed"}, detailed, secondDetailed, "ChoiceDetailedResponseBody"},
				{wireUnionUse{role: wireRequestBody}, request, secondRequest, "ChoiceRequestBody"},
			} {
				scope := catalog.resolverForUse(catalog.scope, test.policy, test.use)
				for _, source := range []*expr.AttributeExpr{choice, test.second} {
					copied := expr.DupAtt(source)
					require.NoError(t, catalog.bindCopiedUnionOccurrences(copied, source, test.use, "body", make(map[wireAttributePair]struct{})))
					require.NotNil(t, catalog.unionBindings[wireUnionBinding{union: expr.AsUnion(copied.Type), use: test.use}])
					require.Equal(t, test.name, scope.Name(copied, "", false, false))
					require.Equal(t, test.name, scope.Name(source, "", false, false))
				}
			}
		})
	}
}

// TestHTTPUnionBodyRoleCallers verifies that HTTP error responses, starter
// multipart callbacks, WebSocket streaming payload validators, and mixed SSE
// streaming results resolve union declarations under their body role. It builds
// one combined generated module because go/parser alone cannot verify cross-file
// pointer, value, and package-qualifier compatibility between generated server,
// client, service, and starter example packages.
func TestHTTPUnionBodyRoleCallers(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.API("union_roles", func() {
			dsl.Server("union_roles", func() {
				dsl.Services("values", "uploads", "events", "mixed_events")
			})
		})

		problem := dsl.Type("Problem", func() {
			dsl.OneOf("detail", func() {
				dsl.Attribute("reason", dsl.String)
				dsl.Attribute("code", dsl.Int)
			})
			dsl.Required("detail")
		})
		optionalProblem := dsl.Type("OptionalProblem", func() {
			dsl.OneOf("optional_detail", func() {
				dsl.Attribute("reason", dsl.String)
				dsl.Attribute("code", dsl.Int)
			})
		})
		dsl.Service("values", func() {
			dsl.Error("bad_request", problem)
			dsl.Error("unprocessable", optionalProblem)
			dsl.Method("read", func() {
				dsl.Result(dsl.String)
				dsl.HTTP(func() {
					dsl.GET("/read")
					dsl.Response(dsl.StatusOK)
					dsl.Response("bad_request", dsl.StatusBadRequest, func() {
						dsl.Body("detail")
					})
					dsl.Response("unprocessable", dsl.StatusUnprocessableEntity, func() {
						dsl.Body("optional_detail")
					})
				})
			})
		})

		partBranch := dsl.Type("PartBranch", func() {
			dsl.Attribute("label", dsl.String)
			dsl.Required("label")
		})
		partItem := dsl.Type("PartItem", func() {
			dsl.OneOf("choice", func() {
				dsl.Attribute("branch", partBranch)
			})
			dsl.Required("choice")
		})
		dsl.Service("uploads", func() {
			dsl.Method("send_object", func() {
				dsl.Payload(partItem)
				dsl.HTTP(func() {
					dsl.POST("/send-object")
					dsl.MultipartRequest()
					dsl.Response(dsl.StatusNoContent)
				})
			})
			dsl.Method("send_array", func() {
				dsl.Payload(dsl.ArrayOf(partItem))
				dsl.HTTP(func() {
					dsl.POST("/send-array")
					dsl.MultipartRequest()
					dsl.Response(dsl.StatusNoContent)
				})
			})
			dsl.Method("send_map", func() {
				dsl.Payload(dsl.MapOf(dsl.Int, partItem, func() {
					dsl.Key(func() {
						dsl.Meta("struct:field:type", "time.Month", "time")
					})
				}))
				dsl.HTTP(func() {
					dsl.POST("/send-map")
					dsl.MultipartRequest()
					dsl.Response(dsl.StatusNoContent)
				})
			})
		})

		eventBranch := dsl.Type("EventBranch", func() {
			dsl.Attribute("meta", func() {
				dsl.Attribute("key", dsl.String, func() {
					dsl.MinLength(1)
				})
				dsl.Required("key")
			})
			dsl.Required("meta")
		})
		streamMessage := dsl.Type("StreamMessage", func() {
			dsl.Attribute("branch", eventBranch)
			dsl.OneOf("event", func() {
				dsl.Attribute("branch", eventBranch)
			})
			dsl.Required("branch", "event")
		})
		dsl.Service("events", func() {
			dsl.Method("stream", func() {
				dsl.StreamingPayload(streamMessage)
				dsl.HTTP(func() {
					dsl.GET("/stream")
					dsl.Response(dsl.StatusOK)
				})
			})
		})

		streamBranch := dsl.Type("StreamBranch", func() {
			dsl.Attribute("label", dsl.String)
			dsl.Required("label")
		})
		dsl.Service("mixed_events", func() {
			dsl.Method("watch", func() {
				dsl.Result(dsl.String)
				dsl.StreamingResult(func() {
					dsl.OneOf("item", func() {
						dsl.Attribute("branch", streamBranch)
					})
					dsl.Required("item")
				})
				dsl.HTTP(func() {
					dsl.GET("/watch")
					dsl.ServerSentEvents()
				})
			})
		})
	})

	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	examplePlan, err := example.NewPlan(generation, servicePlan)
	require.NoError(t, err)
	examples, err := NewExamplePlan(plans[0], examplePlan)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	readEndpoint := plans[0].services.Get("values").Endpoint("read")
	require.Len(t, readEndpoint.Errors, 2)
	for index, wantTypeRef := range []string{"*OptionalDetailResponseBody", "*DetailResponseBody"} {
		errInit := readEndpoint.Errors[index].Errors[0].Response.ResultInit
		require.Equal(t, "&body", errInit.ClientArgs[0].Ref)
		require.Equal(t, wantTypeRef, errInit.ClientArgs[0].TypeRef)
		require.True(t, errInit.ReturnIsUnion)
	}

	var multipartFile *codegen.File
	for _, file := range examples.ServerFiles() {
		if file.Path == "multipart.go" {
			multipartFile = file
		}
	}
	require.NotNil(t, multipartFile)
	multipartSource := renderedHTTPFile(t, multipartFile)
	require.Contains(t, multipartSource, "*uploadssvr.SendObjectRequestBody")
	require.Contains(t, multipartSource, "*[]*uploadssvr.PartItemRequestBody")
	require.Contains(t, multipartSource, "*map[time.Month]*uploadssvr.PartItemRequestBody")

	serviceFiles, err := service.Files(servicePlan)
	require.NoError(t, err)

	directory := t.TempDir()
	writeHTTPGeneratedModule(t, directory)
	allFiles := slices.Concat(
		serviceFiles,
		plans[0].ServerFiles(),
		plans[0].ClientFiles(),
		plans[0].ServerTypeFiles(),
		plans[0].ClientTypeFiles(),
		plans[0].PathFiles(),
		plans[0].ClientCLIFiles(),
		[]*codegen.File{multipartFile},
	)
	for _, file := range allFiles {
		_, err := file.Render(directory)
		require.NoError(t, err, file.Path)
	}
	multipartCheck := `package unionroles

import uploadssvr "generated.local/gen/http/uploads/server"

var (
	_ uploadssvr.UploadsSendObjectDecoderFunc = UploadsSendObjectDecoderFunc
	_ uploadssvr.UploadsSendArrayDecoderFunc  = UploadsSendArrayDecoderFunc
	_ uploadssvr.UploadsSendMapDecoderFunc    = UploadsSendMapDecoderFunc
)
`
	require.NoError(t, os.WriteFile(filepath.Join(directory, "multipart_check.go"), []byte(multipartCheck), 0o600))

	parallelHTTPTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "build", "-mod=readonly", "./...")
	command.Dir = directory
	output, err := command.CombinedOutput()
	require.NoError(t, err, "%s", output)
}

func TestExampleMultipartBodyTypeQualifiesNonUserTypeUnion(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("uploads", func() {
			dsl.Method("send", func() {
				dsl.Payload(dsl.ArrayOf(&expr.Object{}, func() {
					dsl.Attribute("at", dsl.String, func() {
						dsl.Meta("struct:field:type", "time.Time", "time")
					})
					dsl.OneOf("choice", func() {
						dsl.Attribute("text", dsl.String)
						dsl.Attribute("count", dsl.Int)
					})
					dsl.Required("at", "choice")
				}))
				dsl.HTTP(func() {
					dsl.POST("/send")
					dsl.MultipartRequest()
					dsl.Response(dsl.StatusNoContent)
				})
			})
		})
	})
	examples := linkedHTTPExamplePlanForRoot(t, root)
	var multipartSource, serverSource string
	for _, file := range examples.ServerFiles() {
		if file.Path == "multipart.go" {
			multipartSource = renderedHTTPFile(t, file)
		}
	}
	for _, file := range examples.transport.ServerFiles() {
		if filepath.Base(file.Path) == "server.go" {
			serverSource = renderedHTTPFile(t, file)
		}
	}
	require.Contains(t, multipartSource, "body *[]struct {\n\tAt     *time.Time")
	require.Contains(t, multipartSource, "\tChoice *uploadssvr.ChoiceRequestBody")
	require.Contains(t, serverSource, "func(*multipart.Reader, *[]struct {\n\tAt     *time.Time")
	require.Contains(t, serverSource, "\tChoice *ChoiceRequestBody")
}
