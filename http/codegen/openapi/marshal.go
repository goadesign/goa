package openapi

import (
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
	var unmarshaled any
	if err := json.Unmarshal(marshaled, &unmarshaled); err != nil {
		return nil, err
	}
	asserted := unmarshaled.(map[string]any)
	maps.Copy(asserted, extensions)
	merged, err := json.Marshal(asserted)
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
