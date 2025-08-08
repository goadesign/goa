package framework

import (
	"testing"
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
		{"stream_string_final", "stream", "string", "final", false},
		
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