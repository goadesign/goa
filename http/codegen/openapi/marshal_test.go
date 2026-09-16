// This file verifies that adding OpenAPI extensions does not change values in
// the object being encoded.
package openapi

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type (
	marshalExample struct {
		Value int64 `json:"value"`
	}
)

func TestMarshalJSONPreservesLargeIntegers(t *testing.T) {
	const value int64 = 9007199254740993

	encoded, err := MarshalJSON(marshalExample{Value: value}, map[string]any{"x-test": true})
	require.NoError(t, err)

	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var decoded map[string]any
	require.NoError(t, decoder.Decode(&decoded))
	require.Equal(t, json.Number("9007199254740993"), decoded["value"])
}
