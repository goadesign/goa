// This file verifies that the prepared-design snapshot reports persistent
// semantic mutations made after the lifecycle's explicit preparation phase.
package generator

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestPreparedDesignSnapshotRejectsUnsupportedState proves live behavior not
// owned by the evaluated design cannot enter the persistent mutation audit.
func TestPreparedDesignSnapshotRejectsUnsupportedState(t *testing.T) {
	value := 1
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"channel", make(chan int), "unsupported non-nil channel"},
		{"unsafe pointer", unsafe.Pointer(&value), "unsupported non-nil unsafe pointer"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := &expr.RootExpr{Types: []expr.UserType{&expr.UserTypeExpr{
				TypeName: "Value",
				AttributeExpr: &expr.AttributeExpr{
					Type:         expr.String,
					DefaultValue: test.value,
				},
			}}}
			_, err := snapshotPreparedDesign([]eval.Root{root})
			require.ErrorContains(t, err, "roots[0].Types[0].AttributeExpr.DefaultValue")
			require.ErrorContains(t, err, test.want)
		})
	}
}

// TestPreparedDesignSnapshotTracksFunctions proves unchanged functions remain
// valid while replacement and nilness changes are reported as mutations.
func TestPreparedDesignSnapshotTracksFunctions(t *testing.T) {
	first := func(string) string { return "first" }
	second := func(string) string { return "second" }
	root := &expr.RootExpr{Types: []expr.UserType{&expr.UserTypeExpr{
		TypeName: "Value",
		AttributeExpr: &expr.AttributeExpr{
			Type:         expr.String,
			DefaultValue: first,
		},
	}}}

	snapshot, err := snapshotPreparedDesign([]eval.Root{root})
	require.NoError(t, err)
	changed, err := snapshot.changedPath([]eval.Root{root})
	require.NoError(t, err)
	require.Empty(t, changed)

	root.Types[0].Attribute().DefaultValue = second
	changed, err = snapshot.changedPath([]eval.Root{root})
	require.NoError(t, err)
	require.Equal(t, "roots[0].Types[0].AttributeExpr.DefaultValue", changed)

	root.Types[0].Attribute().DefaultValue = (func(string) string)(nil)
	changed, err = snapshot.changedPath([]eval.Root{root})
	require.NoError(t, err)
	require.Equal(t, "roots[0].Types[0].AttributeExpr.DefaultValue", changed)
}

// TestPreparedDesignSnapshotDetectsNilFunctionReplacement proves a function
// added where the prepared design stored nil is reported as a mutation.
func TestPreparedDesignSnapshotDetectsNilFunctionReplacement(t *testing.T) {
	root := &expr.RootExpr{Types: []expr.UserType{&expr.UserTypeExpr{
		TypeName: "Value",
		AttributeExpr: &expr.AttributeExpr{
			Type:         expr.String,
			DefaultValue: (func(string) string)(nil),
		},
	}}}
	snapshot, err := snapshotPreparedDesign([]eval.Root{root})
	require.NoError(t, err)

	root.Types[0].Attribute().DefaultValue = func(string) string { return "added" }
	changed, err := snapshot.changedPath([]eval.Root{root})
	require.NoError(t, err)
	require.Equal(t, "roots[0].Types[0].AttributeExpr.DefaultValue", changed)
}

// TestPreparedDesignSnapshotTreatsConversionExternalAsTypeToken proves that
// conversion exemplars contribute their exact Go type but not instance state.
func TestPreparedDesignSnapshotTreatsConversionExternalAsTypeToken(t *testing.T) {
	type firstExternal struct {
		channel chan int
		values  []string
	}
	type secondExternal struct{}
	external := &firstExternal{channel: make(chan int), values: []string{"before"}}
	typeMap := &expr.TypeMap{External: external}
	root := &expr.RootExpr{Conversions: []*expr.TypeMap{typeMap}}
	snapshot, err := snapshotPreparedDesign([]eval.Root{root})
	require.NoError(t, err)

	external.values[0] = "after"
	changed, err := snapshot.changedPath([]eval.Root{root})
	require.NoError(t, err)
	require.Empty(t, changed)

	typeMap.External = &secondExternal{}
	changed, err = snapshot.changedPath([]eval.Root{root})
	require.NoError(t, err)
	require.Equal(t, "roots[0].Conversions[0].External", changed)
}

// TestPreparedDesignSnapshotMapOrderIsStable proves randomized Go map
// iteration does not produce false mutation reports.
func TestPreparedDesignSnapshotMapOrderIsStable(t *testing.T) {
	root := &expr.RootExpr{Types: []expr.UserType{&expr.UserTypeExpr{
		TypeName: "Value",
		AttributeExpr: &expr.AttributeExpr{
			Type: expr.String,
			DefaultValue: map[any]any{
				"zeta":  []string{"last"},
				"alpha": []string{"first"},
				42:      "number",
			},
		},
	}}}

	snapshot, err := snapshotPreparedDesign([]eval.Root{root})
	require.NoError(t, err)
	for range 100 {
		changed, err := snapshot.changedPath([]eval.Root{root})
		require.NoError(t, err)
		require.Empty(t, changed)
	}
}

// TestPreparedDesignSnapshotDetectsAliasReplacement proves replacing one of
// two aliases with an equal-value allocation changes the recorded topology.
func TestPreparedDesignSnapshotDetectsAliasReplacement(t *testing.T) {
	service := &expr.ServiceExpr{Name: "service"}
	root := &expr.RootExpr{Services: []*expr.ServiceExpr{service, service}}
	snapshot, err := snapshotPreparedDesign([]eval.Root{root})
	require.NoError(t, err)

	replacement := *service
	root.Services[1] = &replacement
	changed, err := snapshot.changedPath([]eval.Root{root})
	require.NoError(t, err)
	require.Equal(t, "roots[0].Services[1]", changed)
}

// TestPreparedDesignSnapshotDetectsInPlaceContainerMutation proves changes to
// existing map and slice storage are visible without replacing a container.
func TestPreparedDesignSnapshotDetectsInPlaceContainerMutation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string][]string)
	}{
		{"map", func(values map[string][]string) { values["second"] = []string{"new"} }},
		{"slice", func(values map[string][]string) { values["first"][0] = "changed" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			values := map[string][]string{"first": {"original"}}
			root := &expr.RootExpr{Types: []expr.UserType{&expr.UserTypeExpr{
				TypeName: "Value",
				AttributeExpr: &expr.AttributeExpr{
					Type:         expr.String,
					DefaultValue: values,
				},
			}}}
			snapshot, err := snapshotPreparedDesign([]eval.Root{root})
			require.NoError(t, err)

			test.mutate(values)
			changed, err := snapshot.changedPath([]eval.Root{root})
			require.NoError(t, err)
			require.NotEmpty(t, changed)
		})
	}
}
