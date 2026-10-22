// This file checks that a linked gRPC plan uses the Go names produced by the
// supported protobuf tools in every generated client and server file.
package codegen

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

type (
	protobufModuleTemplate struct {
		goMod []byte
		goSum []byte
	}

	protobufCompileCase struct {
		name string
		dir  string
	}
)

const protobufModuleSeed = `package seed

import (
	_ "github.com/stretchr/testify/require"
	_ "goa.design/goa/v3/grpc"
	_ "goa.design/goa/v3/grpc/pb"
	_ "goa.design/goa/v3/http"
	_ "goa.design/goa/v3/pkg"
	_ "google.golang.org/grpc"
	_ "google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/status"
	_ "google.golang.org/protobuf/proto"
	_ "google.golang.org/protobuf/reflect/protoreflect"
	_ "google.golang.org/protobuf/runtime/protoimpl"
)
`

var (
	parallelGRPCTests      sync.Map
	cachedProtobufModFiles = sync.OnceValues(func() (protobufModuleTemplate, error) {
		goaRoot, err := filepath.Abs(filepath.Join("..", ".."))
		if err != nil {
			return protobufModuleTemplate{}, err
		}
		tmpDir, err := os.MkdirTemp("", "goa-grpc-mod-template-*")
		if err != nil {
			return protobufModuleTemplate{}, err
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		module := "module generated.local\n\ngo 1.25\n\nrequire goa.design/goa/v3 v3.0.0\n\n" +
			"replace goa.design/goa/v3 => " + filepath.ToSlash(goaRoot) + "\n"
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(module), 0o600); err != nil {
			return protobufModuleTemplate{}, err
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "seed.go"), []byte(protobufModuleSeed), 0o600); err != nil {
			return protobufModuleTemplate{}, err
		}
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = tmpDir
		cmd.Env = append(os.Environ(), "GOWORK=off")
		if out, err := cmd.CombinedOutput(); err != nil {
			return protobufModuleTemplate{}, &os.PathError{Op: "go mod tidy", Path: string(out), Err: err}
		}
		goMod, err := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
		if err != nil {
			return protobufModuleTemplate{}, err
		}
		goSum, err := os.ReadFile(filepath.Join(tmpDir, "go.sum"))
		if err != nil {
			return protobufModuleTemplate{}, err
		}
		return protobufModuleTemplate{goMod: goMod, goSum: goSum}, nil
	})
)

func parallelGRPCTest(t *testing.T) {
	t.Helper()
	if _, loaded := parallelGRPCTests.LoadOrStore(t, struct{}{}); !loaded {
		t.Parallel()
	}
}

func compileProtobufDescriptorCases(t *testing.T, cases []protobufCompileCase) {
	t.Helper()
	parallelGRPCTest(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compileProtobufDescriptorModule(t, tc.dir)
		})
	}
}

// TestPlanUsesNamesFromSupportedProtobufTools checks both method orders because
// changing source order must not change the Go declarations chosen for a file.
func TestPlanUsesNamesFromSupportedProtobufTools(t *testing.T) {
	compiled := make([]protobufCompileCase, 0, 2)
	for _, reverse := range []bool{false, true} {
		name := "unary-first"
		if reverse {
			name = "stream-first"
		}
		moduleDir, protoPath, generatedGo := renderProtobufDescriptorPlan(t, reverse)
		descriptor := describeGeneratedProto(t, protoPath)
		rules, err := newProtocNameRules(protocNameVersionGo1_36GRPC1_6)
		require.NoError(t, err)
		names, err := rules.file(descriptor)
		require.NoError(t, err)

		serviceDescriptor := descriptor.GetService()[0]
		messageDescriptor := messageWithOneof(t, descriptor, "result2_kind")
		oneofDescriptor := messageDescriptor.GetOneofDecl()[0]
		resetDescriptor := fieldNamed(t, messageDescriptor, "reset")
		apiURLDescriptor := fieldNamed(t, messageDescriptor, "api_url")
		dns2Descriptor := fieldNamed(t, messageDescriptor, "dns2_server")
		stringDescriptor := fieldNamed(t, messageDescriptor, "string_")
		oneofStringDescriptor := fieldNamed(t, messageDescriptor, "string_2")
		require.NotEqual(t, stringDescriptor.GetName(), oneofStringDescriptor.GetName())
		require.NotNil(t, oneofStringDescriptor.OneofIndex)
		packageName := descriptor.GetPackage()
		serviceName := packageName + "." + serviceDescriptor.GetName()
		messageName := packageName + "." + messageDescriptor.GetName()
		outerStringName, ok := names.lookup(messageName+"."+stringDescriptor.GetName(), protocFieldName)
		require.True(t, ok)
		require.Equal(t, "String_", outerStringName)
		oneofStringName, ok := names.lookup(messageName+"."+oneofStringDescriptor.GetName(), protocFieldName)
		require.True(t, ok)
		require.Equal(t, "String_2", oneofStringName)
		oneofStringWrapper, ok := names.lookup(messageName+"."+oneofStringDescriptor.GetName(), protocOneofWrapperName)
		require.True(t, ok)
		require.Equal(t, messageDescriptor.GetName()+"_String_2", oneofStringWrapper)
		methodNames := make(map[string]string, len(serviceDescriptor.GetMethod()))
		for _, method := range serviceDescriptor.GetMethod() {
			methodNames[method.GetName()] = serviceName + "." + method.GetName()
		}

		checks := []struct {
			descriptor string
			role       protocNameRole
		}{
			{messageName, protocMessageName},
			{messageName + "." + apiURLDescriptor.GetName(), protocFieldName},
			{messageName + "." + dns2Descriptor.GetName(), protocFieldName},
			{messageName + "." + stringDescriptor.GetName(), protocFieldName},
			{messageName + "." + oneofStringDescriptor.GetName(), protocFieldName},
			{messageName + "." + oneofStringDescriptor.GetName(), protocOneofWrapperName},
			{messageName + "." + oneofDescriptor.GetName(), protocOneofFieldName},
			{messageName + "." + resetDescriptor.GetName(), protocOneofWrapperName},
			{serviceName, protocServiceClientName},
			{serviceName, protocServiceServerName},
			{methodNames["GetUrl2"], protocMethodName},
			{methodNames["SyncX509"], protocMethodName},
			{methodNames["SyncX509"], protocMethodClientStreamName},
			{methodNames["SyncX509"], protocMethodServerStreamName},
		}
		declarations := declaredGoNames(t,
			strings.TrimSuffix(protoPath, ".proto")+".pb.go",
			strings.TrimSuffix(protoPath, ".proto")+"_grpc.pb.go",
		)
		for _, check := range checks {
			declaredName, ok := names.lookup(check.descriptor, check.role)
			require.True(t, ok, "%s was not recorded for %s", check.role, check.descriptor)
			require.Contains(t, declarations, declaredName, "the protobuf tools did not declare %s", declaredName)
			require.True(t, strings.Contains(generatedGo, declaredName), "Goa did not use %s", declaredName)
		}

		protoSource, err := os.ReadFile(protoPath)
		require.NoError(t, err)
		require.Contains(t, string(protoSource), "message lower_snake_message {")
		require.Contains(t, string(protoSource), "lower_snake_message lower = 2;")
		require.Contains(t, string(protoSource), "message "+messageDescriptor.GetName()+" {")
		require.Contains(t, string(protoSource), "oneof result2_kind {")
		require.Contains(t, string(protoSource), "service "+serviceDescriptor.GetName()+" {")
		require.Contains(t, string(protoSource), "rpc GetUrl2 (")
		require.Contains(t, string(protoSource), "rpc CafRead (")
		require.Contains(t, string(protoSource), "rpc SyncX509 (stream ")
		compiled = append(compiled, protobufCompileCase{name: name, dir: moduleDir})
	}
	compileProtobufDescriptorCases(t, compiled)
}

// TestPlanWritesLegalFieldAndOneofNames checks names that would make protoc
// reject the complete generated file if Goa wrote them unchanged.
func TestPlanWritesLegalFieldAndOneofNames(t *testing.T) {
	_, protoPath, _ := renderProtobufDescriptorPlan(t, false)
	protoSource, err := os.ReadFile(protoPath)
	require.NoError(t, err)

	require.Contains(t, string(protoSource), "message LeadingDigit {")
	require.Contains(t, string(protoSource), "optional string _123_field = 1;")
	require.Contains(t, string(protoSource), "message UnicodeName {")
	require.Contains(t, string(protoSource), "optional string caf_field = 1;")
	require.Contains(t, string(protoSource), "oneof foo_bar_oneof {")
	require.Contains(t, string(protoSource), "optional string foo_bar = 3;")
	require.Contains(t, string(protoSource), "message CollisionReverse {")
	require.Equal(t, 2, strings.Count(string(protoSource), "oneof foo_bar_oneof {"))
}

// TestPlanRejectsIllegalExactProtobufName checks that an exact metadata name
// fails before Goa writes a protobuf file that protoc cannot parse.
func TestPlanRejectsIllegalExactProtobufName(t *testing.T) {
	root := RunGRPCDSL(t, func() {
		message := dsl.Type("Message", func() {
			dsl.Meta("struct:name:proto", "123_message")
			dsl.Field(1, "value", dsl.String)
		})
		dsl.Service("invalid", func() {
			dsl.Method("read", func() {
				dsl.Payload(message)
				dsl.Result(message)
				dsl.GRPC(func() {})
			})
		})
	})
	generation, servicePlans := grpcServicePlans(t, []*expr.RootExpr{root})

	_, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlans[0]})
	require.EqualError(t, err, `service "invalid" protobuf message name "123_message" from struct:name:proto is not a valid protobuf identifier`)
}

// renderProtobufDescriptorPlan writes one linked plan and returns the temporary
// module, its protobuf source file, and the Goa client and server source.
func renderProtobufDescriptorPlan(t *testing.T, reverse bool) (string, string, string) {
	t.Helper()
	root := RunGRPCDSL(t, protobufDescriptorPlanDSL(reverse))
	generation, servicePlans := grpcServicePlans(t, []*expr.RootExpr{root})
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlans[0]})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlans[0].Link())
	require.NoError(t, plans[0].Link())

	files, err := service.Files(servicePlans...)
	require.NoError(t, err)
	files = append(files, plans[0].ServerFiles()...)
	files = append(files, plans[0].ClientFiles()...)
	files = append(files, plans[0].ServerTypeFiles()...)
	files = append(files, plans[0].ClientTypeFiles()...)
	files = append(files, plans[0].ProtoFiles()...)

	moduleDir := t.TempDir()
	writeProtobufDescriptorModule(t, moduleDir)
	var protoPath string
	var generated strings.Builder
	for _, file := range files {
		path, err := file.Render(moduleDir)
		require.NoError(t, err)
		if filepath.Ext(path) == ".proto" {
			protoPath = path
			continue
		}
		if strings.Contains(filepath.ToSlash(path), "/grpc/") {
			source, err := os.ReadFile(path)
			require.NoError(t, err)
			generated.Write(source)
		}
	}
	require.NotEmpty(t, protoPath)
	return moduleDir, protoPath, generated.String()
}

// protobufDescriptorPlanDSL returns a design that uses one object in unary and
// streaming messages and places its preferred protobuf name beside a service
// declaration with the same Go name.
func protobufDescriptorPlanDSL(reverse bool) func() {
	return func() {
		leadingDigit := dsl.Type("LeadingDigit", func() {
			dsl.Field(1, "123_field", dsl.String, func() {
				dsl.Meta("struct:field:name", "LeadingField")
			})
		})
		unicodeName := dsl.Type("UnicodeName", func() {
			dsl.Field(1, "caféField", dsl.String, func() {
				dsl.Meta("struct:field:name", "CafeField")
			})
		})
		collision := dsl.Type("Collision", func() {
			dsl.OneOf("fooBar", func() {
				dsl.TypeName("CollisionFooBar")
				dsl.Field(2, "text", dsl.String)
			})
			dsl.Field(3, "foo_bar", dsl.String, func() {
				dsl.Meta("struct:field:name", "OtherFooBar")
			})
		})
		collisionReverse := dsl.Type("CollisionReverse", func() {
			dsl.Field(1, "foo_bar", dsl.String, func() {
				dsl.Meta("struct:field:name", "OtherFooBar")
			})
			dsl.OneOf("fooBar", func() {
				dsl.TypeName("CollisionReverseFooBar")
				dsl.Field(2, "text", dsl.String)
			})
		})
		shared := dsl.Type("Api2HttpServiceClient", func() {
			dsl.Field(1, "api_url", dsl.String)
			dsl.Field(2, "dns_2_server", dsl.String)
			dsl.Field(5, "string", dsl.String)
			dsl.OneOf("result_2_kind", func() {
				dsl.Field(3, "http_2xx", dsl.String)
				dsl.Field(4, "reset", dsl.String)
				dsl.Field(6, "string", dsl.String)
			})
		})
		lowerSnake := dsl.Type("LowerSnake", func() {
			dsl.Meta("struct:name:proto", "lower_snake_message")
			dsl.Field(1, "value", dsl.String)
		})
		envelope := dsl.Type("API2Envelope", func() {
			dsl.Field(1, "client", shared)
			dsl.Field(2, "lower", lowerSnake)
			dsl.Field(3, "leading", leadingDigit)
			dsl.Field(4, "collision", collision)
			dsl.Field(5, "unicode", unicodeName)
			dsl.Field(6, "collisionReverse", collisionReverse)
		})
		unary := func() {
			dsl.Method("get_url2", func() {
				dsl.Payload(envelope)
				dsl.Result(envelope)
				dsl.GRPC(func() {})
			})
		}
		stream := func() {
			dsl.Method("sync_x509", func() {
				dsl.StreamingPayload(envelope)
				dsl.StreamingResult(envelope)
				dsl.GRPC(func() {})
			})
		}
		unicodeMethod := func() {
			dsl.Method("café_read", func() {
				dsl.Payload(envelope)
				dsl.Result(envelope)
				dsl.GRPC(func() {})
			})
		}
		dsl.Service("api2_http_service", func() {
			if reverse {
				stream()
				unary()
				unicodeMethod()
				return
			}
			unary()
			stream()
			unicodeMethod()
		})
	}
}

// describeGeneratedProto asks protoc for the names and fields written in one
// generated protobuf source file.
func describeGeneratedProto(t *testing.T, protoPath string) *descriptorpb.FileDescriptorProto {
	t.Helper()
	descriptorPath := filepath.Join(t.TempDir(), "descriptor.pb")
	args := defaultProtocCmd[1:len(defaultProtocCmd):len(defaultProtocCmd)]
	args = append(args,
		"--proto_path", filepath.Dir(protoPath),
		"--descriptor_set_out", descriptorPath,
		protoPath,
	)
	output, err := exec.Command(defaultProtocCmd[0], args...).CombinedOutput()
	require.NoError(t, err, string(output))
	encoded, err := os.ReadFile(descriptorPath)
	require.NoError(t, err)
	set := &descriptorpb.FileDescriptorSet{}
	require.NoError(t, proto.Unmarshal(encoded, set))
	require.Len(t, set.File, 1)
	return set.File[0]
}

// messageWithOneof returns the message that declares the named choice field.
func messageWithOneof(t *testing.T, file *descriptorpb.FileDescriptorProto, name string) *descriptorpb.DescriptorProto {
	t.Helper()
	for _, message := range file.GetMessageType() {
		for _, oneof := range message.GetOneofDecl() {
			if oneof.GetName() == name {
				return message
			}
		}
	}
	t.Fatalf("protobuf source did not declare oneof %q", name)
	return nil
}

// fieldNamed returns the field with the requested protobuf source name.
func fieldNamed(t *testing.T, message *descriptorpb.DescriptorProto, name string) *descriptorpb.FieldDescriptorProto {
	t.Helper()
	for _, field := range message.GetField() {
		if field.GetName() == name {
			return field
		}
	}
	t.Fatalf("protobuf message %q did not declare field %q", message.GetName(), name)
	return nil
}

// writeProtobufDescriptorModule writes a module that imports this Goa checkout.
func writeProtobufDescriptorModule(t *testing.T, directory string) {
	t.Helper()
	tmpl, err := cachedProtobufModFiles()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(directory, "go.mod"), tmpl.goMod, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "go.sum"), tmpl.goSum, 0o600))
}

func hasProtobufTestFiles(dir string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(entry.Name(), "_test.go") {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// compileProtobufDescriptorModule compiles every package written by the plan.
func compileProtobufDescriptorModule(t *testing.T, directory string) {
	t.Helper()
	parallelGRPCTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	args := []string{"build", "-mod=readonly", "./..."}
	if hasProtobufTestFiles(directory) {
		args = []string{"test", "-mod=readonly", "./..."}
	}
	command := exec.CommandContext(ctx, "go", args...)
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
}
