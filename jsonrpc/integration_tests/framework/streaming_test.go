package framework

import (
	"testing"
)

func TestGenerateStreamMessages(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		info     MethodInfo
		wantMsgs int
	}{
		{
			name:   "stream_string_final",
			method: "test",
			info: MethodInfo{
				Action:   "stream",
				Type:     "string",
				Modifier: "final",
			},
			wantMsgs: 4, // 3 notifications + 1 final
		},
		{
			name:   "stream_object",
			method: "test",
			info: MethodInfo{
				Action: "stream",
				Type:   "object",
			},
			wantMsgs: 3, // 3 notifications only
		},
	}

	config := DefaultStreamingConfig()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := GenerateStreamMessages(tt.method, tt.info, config)
			if err != nil {
				t.Fatalf("GenerateStreamMessages() error = %v", err)
			}
			
			if len(messages) != tt.wantMsgs {
				t.Errorf("GenerateStreamMessages() got %d messages, want %d", len(messages), tt.wantMsgs)
			}
			
			// Check final message has ID if modifier is "final"
			if tt.info.Modifier == "final" {
				lastMsg := messages[len(messages)-1]
				if !lastMsg.HasID {
					t.Error("Final message should have ID")
				}
				if lastMsg.Type != "result" {
					t.Errorf("Final message type = %v, want result", lastMsg.Type)
				}
			}
		})
	}
}

func TestGenerateBroadcastMessages(t *testing.T) {
	info := MethodInfo{
		Action: "broadcast",
		Type:   "string",
	}
	
	config := DefaultStreamingConfig()
	messages, err := GenerateBroadcastMessages("test", info, config)
	if err != nil {
		t.Fatalf("GenerateBroadcastMessages() error = %v", err)
	}
	
	// Should have 1 ack + 2 broadcasts
	if len(messages) != 3 {
		t.Errorf("GenerateBroadcastMessages() got %d messages, want 3", len(messages))
	}
	
	// First message should be acknowledgment
	if messages[0].Type != "result" || !messages[0].HasID {
		t.Error("First message should be result with ID")
	}
	
	// Rest should be broadcasts without ID
	for i := 1; i < len(messages); i++ {
		if messages[i].HasID {
			t.Errorf("Broadcast message %d should not have ID", i)
		}
	}
}

func TestStreamingContext(t *testing.T) {
	messages := []StreamMessage{
		{Type: "notification", Data: "msg1"},
		{Type: "notification", Data: "msg2"},
		{Type: "result", Data: "final", HasID: true, ID: "test"},
	}
	
	ctx := NewStreamingContext(messages)
	
	// Test iteration
	count := 0
	for ctx.HasNext() {
		msg, ok := ctx.Next()
		if !ok {
			t.Error("Next() returned false while HasNext() is true")
		}
		if msg.Data != messages[count].Data {
			t.Errorf("Message %d data mismatch", count)
		}
		count++
	}
	
	if count != len(messages) {
		t.Errorf("Iterated %d messages, expected %d", count, len(messages))
	}
	
	// Test client data collection
	ctx.AddClientData("client1")
	ctx.AddClientData("client2")
	
	result := ctx.GetCollectedResult(MethodInfo{Type: "string"})
	if result != "client1, client2" {
		t.Errorf("GetCollectedResult() = %v, want 'client1, client2'", result)
	}
}

func TestGetStreamingBehavior(t *testing.T) {
	tests := []struct {
		name     string
		info     MethodInfo
		expected StreamingBehavior
	}{
		{
			name: "stream_final",
			info: MethodInfo{Action: "stream", Modifier: "final"},
			expected: StreamingBehavior{
				SendNotifications: true,
				SendFinalResponse: true,
			},
		},
		{
			name: "broadcast",
			info: MethodInfo{Action: "broadcast"},
			expected: StreamingBehavior{
				IsBroadcast:       true,
				SendNotifications: true,
			},
		},
		{
			name: "collect",
			info: MethodInfo{Action: "collect"},
			expected: StreamingBehavior{
				IsCollector: true,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStreamingBehavior(tt.info)
			if got != tt.expected {
				t.Errorf("GetStreamingBehavior() = %+v, want %+v", got, tt.expected)
			}
		})
	}
}