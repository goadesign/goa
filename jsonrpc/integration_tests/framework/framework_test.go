package framework

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestParseMethod verifies method name parsing
func TestParseMethod(t *testing.T) {
	tests := []struct {
		method   string
		action   string
		dataType string
		modifier string
		wantErr  bool
	}{
		// Valid methods
		{"echo_string", "echo", "string", "", false},
		{"transform_array", "transform", "array", "", false},
		{"generate_object", "generate", "object", "", false},
		{"echo_string_notify", "echo", "string", "notify", false},
		{"transform_map_error", "transform", "map", "error", false},
		{"stream_string_sse", "stream", "string", "", false},
		{"stream_string_unknown_sse", "", "", "", true},
		{"stream_string", "", "", "", true},
		{"echo_string_ws", "", "", "", true},

		// Invalid methods
		{"invalid", "", "", "", true},
		{"echo", "", "", "", true},
		{"echo_", "", "", "", true},
		{"_string", "", "", "", true},
		{"", "", "", "", true},
		{"echo__string", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			info, err := ParseMethod(tt.method)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseMethod(%q) should have failed", tt.method)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseMethod(%q) failed: %v", tt.method, err)
			}

			if info.Action != tt.action {
				t.Errorf("Action: got %q, want %q", info.Action, tt.action)
			}
			if info.Type != tt.dataType {
				t.Errorf("Type: got %q, want %q", info.Type, tt.dataType)
			}
			if info.Modifier != tt.modifier {
				t.Errorf("Modifier: got %q, want %q", info.Modifier, tt.modifier)
			}
		})
	}
}

// TestScenariosUseSupportedTransports checks that every checked-in scenario
// uses HTTP or server-sent events and that each server-sent event scenario
// contains only receive steps and ends with its request ID when one is present.
func TestScenariosUseSupportedTransports(t *testing.T) {
	runner, err := NewRunner("../scenarios/scenarios.yaml")
	require.NoError(t, err)

	for _, scenario := range runner.config.Scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			require.Contains(t, []string{TransportHTTP, TransportSSE}, scenario.Transport)
			if scenario.Transport != TransportSSE {
				return
			}
			for _, step := range scenario.Sequence {
				require.Equal(t, "receive", step.Type)
			}
			if scenario.Request.ID == nil {
				return
			}
			require.NotEmpty(t, scenario.Sequence)
			response, ok := scenario.Sequence[len(scenario.Sequence)-1].Expect.(map[string]any)
			require.True(t, ok)
			require.EqualValues(t, scenario.Request.ID, response["id"])
		})
	}
}
