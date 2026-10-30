// This file checks that source planning needs no installed protobuf programs
// while explicit disk rendering retains real compilation semantics.
package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/expr"
)

// TestNewSourcePlansNeedsNoProtobufPrograms exercises real naming and source
// rendering with an empty executable search path, then checks explicit Render.
func TestNewSourcePlansNeedsNoProtobufPrograms(t *testing.T) {
	root := grpcPlanRoots(t, "Source")[0]
	generation, services := grpcServicePlans(t, []*expr.RootExpr{root})
	t.Setenv("PATH", t.TempDir())
	plans, err := NewSourcePlans(generation, PlanInput{Root: root, Service: services[0]})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, services[0].Link())
	require.NoError(t, plans[0].Link())
	files := plans[0].ProtoFiles()
	require.Len(t, files, 1)
	source, err := files[0].Source()
	require.NoError(t, err)
	require.Contains(t, string(source), "service Source")
	_, err = files[0].Render(t.TempDir())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "resolve protobuf plugin"), "%v", err)
}
