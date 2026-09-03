package openapi

import (
	"bytes"
	"encoding/json"
	"maps"

	"gopkg.in/yaml.v3"
)

// MarshalJSON produces the JSON resulting from encoding an object composed of
// the fields in v (which must me a struct) and the keys in extensions.
func MarshalJSON(v any, extensions map[string]any) ([]byte, error) {
	marshaled, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(extensions) == 0 {
		return marshaled, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(marshaled))
	decoder.UseNumber()
	var unmarshaled map[string]any
	if err := decoder.Decode(&unmarshaled); err != nil {
		return nil, err
	}
	maps.Copy(unmarshaled, extensions)
	merged, err := json.Marshal(unmarshaled)
	if err != nil {
		return nil, err
	}
	return merged, nil
}

// MarshalYAML produces the JSON resulting from encoding an object composed of
// the fields in v (which must me a struct) and the keys in extensions.
func MarshalYAML(v any, extensions map[string]any) (any, error) {
	if len(extensions) == 0 {
		return v, nil
	}
	marshaled, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	var unmarshaled map[string]any
	if err := yaml.Unmarshal(marshaled, &unmarshaled); err != nil {
		return nil, err
	}
	maps.Copy(unmarshaled, extensions)
	return unmarshaled, nil
}
