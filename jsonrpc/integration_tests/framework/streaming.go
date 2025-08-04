package framework

import (
	"fmt"
	"strings"
	"time"
)

// StreamingConfig holds configuration for streaming behaviors
type StreamingConfig struct {
	// NotificationCount is the number of notifications before final response
	NotificationCount int
	// NotificationInterval is the delay between notifications
	NotificationInterval time.Duration
	// BroadcastCount is the number of broadcasts to send
	BroadcastCount int
	// BroadcastInterval is the delay between broadcasts
	BroadcastInterval time.Duration
}

// DefaultStreamingConfig returns default streaming configuration
func DefaultStreamingConfig() *StreamingConfig {
	return &StreamingConfig{
		NotificationCount:    3,
		NotificationInterval: 100 * time.Millisecond,
		BroadcastCount:       2,
		BroadcastInterval:    200 * time.Millisecond,
	}
}

// StreamingBehavior defines how a streaming method behaves
type StreamingBehavior struct {
	// SendNotifications indicates if method sends notifications before response
	SendNotifications bool
	// SendFinalResponse indicates if method sends a final response after notifications
	SendFinalResponse bool
	// IsBroadcast indicates if this is a server-initiated broadcast
	IsBroadcast bool
	// IsCollector indicates if this collects multiple inputs
	IsCollector bool
}

// GetStreamingBehavior determines streaming behavior from method info
func GetStreamingBehavior(info MethodInfo) StreamingBehavior {
	behavior := StreamingBehavior{}
	
	switch info.Action {
	case "stream":
		behavior.SendNotifications = true
		behavior.SendFinalResponse = info.Modifier == "final"
		
	case "broadcast":
		behavior.IsBroadcast = true
		behavior.SendNotifications = true
		
	case "collect":
		behavior.IsCollector = true
	}
	
	return behavior
}

// StreamMessage represents a message in a stream
type StreamMessage struct {
	// Type is the message type (notification, result, error)
	Type string
	// Data is the message payload
	Data interface{}
	// HasID indicates if this message should have an ID
	HasID bool
	// ID is the message ID (if HasID is true)
	ID interface{}
}

// GenerateStreamMessages generates messages for a streaming method
func GenerateStreamMessages(method string, info MethodInfo, config *StreamingConfig) ([]StreamMessage, error) {
	behavior := GetStreamingBehavior(info)
	var messages []StreamMessage
	
	if behavior.SendNotifications {
		// Generate notification messages
		for i := 0; i < config.NotificationCount; i++ {
			msg := StreamMessage{
				Type:  "notification",
				HasID: false,
			}
			
			// Generate notification data based on type
			switch info.Type {
			case "string":
				msg.Data = fmt.Sprintf("notification-%d", i+1)
			case "array":
				msg.Data = []string{fmt.Sprintf("item-%d", i+1)}
			case "object":
				msg.Data = map[string]interface{}{
					"type":  "notification",
					"index": i + 1,
					"total": config.NotificationCount,
				}
			case "map":
				msg.Data = map[string]interface{}{
					"notification": i + 1,
					"timestamp":    time.Now().Unix(),
				}
			default:
				msg.Data = fmt.Sprintf("notification-%d", i+1)
			}
			
			messages = append(messages, msg)
		}
	}
	
	if behavior.SendFinalResponse {
		// Generate final response with ID
		msg := StreamMessage{
			Type:  "result",
			HasID: true,
			ID:    "stream-final",
		}
		
		// Generate final data based on type
		switch info.Type {
		case "string":
			msg.Data = "stream-complete"
		case "array":
			msg.Data = []string{"final", "result"}
		case "object":
			msg.Data = map[string]interface{}{
				"type":   "complete",
				"status": "success",
				"count":  config.NotificationCount,
			}
		case "map":
			msg.Data = map[string]interface{}{
				"complete":  true,
				"processed": config.NotificationCount,
			}
		default:
			msg.Data = "complete"
		}
		
		messages = append(messages, msg)
	}
	
	return messages, nil
}

// GenerateBroadcastMessages generates server-initiated broadcast messages
func GenerateBroadcastMessages(method string, info MethodInfo, config *StreamingConfig) ([]StreamMessage, error) {
	var messages []StreamMessage
	
	// First, acknowledge the subscription
	ack := StreamMessage{
		Type:  "result",
		HasID: true,
		ID:    "subscription",
		Data: map[string]interface{}{
			"subscribed": true,
			"channel":    method,
		},
	}
	messages = append(messages, ack)
	
	// Then generate broadcast messages
	for i := 0; i < config.BroadcastCount; i++ {
		msg := StreamMessage{
			Type:  "broadcast",
			HasID: false,
		}
		
		// Generate broadcast data based on type
		switch info.Type {
		case "string":
			msg.Data = fmt.Sprintf("broadcast-%d", i+1)
		case "array":
			msg.Data = []string{fmt.Sprintf("update-%d", i+1)}
		case "object":
			msg.Data = map[string]interface{}{
				"type":      "broadcast",
				"sequence":  i + 1,
				"timestamp": time.Now().Unix(),
				"data":      fmt.Sprintf("update-%d", i+1),
			}
		case "map":
			msg.Data = map[string]interface{}{
				"broadcast": i + 1,
				"message":   fmt.Sprintf("Server update %d", i+1),
			}
		default:
			msg.Data = fmt.Sprintf("broadcast-%d", i+1)
		}
		
		messages = append(messages, msg)
	}
	
	return messages, nil
}

// StreamingContext holds context for streaming operations
type StreamingContext struct {
	// Messages to be sent
	Messages []StreamMessage
	// CurrentIndex tracks current message position
	CurrentIndex int
	// ClientData stores data from client for collectors
	ClientData []interface{}
}

// NewStreamingContext creates a new streaming context
func NewStreamingContext(messages []StreamMessage) *StreamingContext {
	return &StreamingContext{
		Messages:     messages,
		CurrentIndex: 0,
		ClientData:   make([]interface{}, 0),
	}
}

// HasNext returns true if there are more messages to send
func (sc *StreamingContext) HasNext() bool {
	return sc.CurrentIndex < len(sc.Messages)
}

// Next returns the next message to send
func (sc *StreamingContext) Next() (StreamMessage, bool) {
	if !sc.HasNext() {
		return StreamMessage{}, false
	}
	
	msg := sc.Messages[sc.CurrentIndex]
	sc.CurrentIndex++
	return msg, true
}

// AddClientData adds data received from client (for collectors)
func (sc *StreamingContext) AddClientData(data interface{}) {
	sc.ClientData = append(sc.ClientData, data)
}

// GetCollectedResult generates a result from collected client data
func (sc *StreamingContext) GetCollectedResult(info MethodInfo) interface{} {
	switch info.Type {
	case "string":
		// Concatenate all strings
		var parts []string
		for _, data := range sc.ClientData {
			if s, ok := data.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ", ")
		
	case "array":
		// Flatten all arrays
		var result []string
		for _, data := range sc.ClientData {
			if arr, ok := data.([]string); ok {
				result = append(result, arr...)
			}
		}
		return result
		
	case "object":
		// Merge all objects
		result := map[string]interface{}{
			"collected": len(sc.ClientData),
			"items":     sc.ClientData,
		}
		return result
		
	case "map":
		// Merge all maps
		result := make(map[string]interface{})
		for i, data := range sc.ClientData {
			if m, ok := data.(map[string]interface{}); ok {
				for k, v := range m {
					result[fmt.Sprintf("%s_%d", k, i)] = v
				}
			}
		}
		result["total_collected"] = len(sc.ClientData)
		return result
		
	default:
		return map[string]interface{}{
			"collected": sc.ClientData,
			"count":     len(sc.ClientData),
		}
	}
}