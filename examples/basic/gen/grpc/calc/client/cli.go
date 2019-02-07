// Code generated by goa v2.0.0-wip, DO NOT EDIT.
//
// calc gRPC client CLI support package
//
// Command:
// $ goa gen goa.design/goa/examples/basic/design -o
// $(GOPATH)/src/goa.design/goa/examples/basic

package client

import (
	"encoding/json"
	"fmt"

	calcsvc "goa.design/goa/examples/basic/gen/calc"
	calcpb "goa.design/goa/examples/basic/gen/grpc/calc/pb"
)

// BuildAddPayload builds the payload for the calc add endpoint from CLI flags.
func BuildAddPayload(calcAddMessage string) (*calcsvc.AddPayload, error) {
	var err error
	var message calcpb.AddRequest
	{
		if calcAddMessage != "" {
			err = json.Unmarshal([]byte(calcAddMessage), &message)
			if err != nil {
				return nil, fmt.Errorf("invalid JSON for message, example of valid JSON:\n%s", "'{\n      \"a\": 8399553735696626949,\n      \"b\": 360622074634248926\n   }'")
			}
		}
	}
	if err != nil {
		return nil, err
	}
	payload := &calcsvc.AddPayload{
		A: int(message.A),
		B: int(message.B),
	}
	return payload, nil
}
