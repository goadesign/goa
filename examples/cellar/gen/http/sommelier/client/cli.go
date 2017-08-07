package client

import (
	"encoding/json"
	"fmt"
)

// BuildPickPayloadFromFlags constructs a "add" endpoint payload from command
// line flag values.
func BuildPickPayloadFromFlags(nameFlag, wineryFlag, varietalFlag string) (*PickRequestBody, error) {
	var name *string
	if nameFlag != "" {
		name = &nameFlag
	}

	var winery *string
	if wineryFlag != "" {
		winery = &wineryFlag
	}

	var varietal []string
	if varietalFlag != "" {
		err := json.Unmarshal([]byte(varietalFlag), &varietal)
		if err != nil {
			ex := []string{"pinot noir"}
			js, _ := json.Marshal(ex)
			return nil, fmt.Errorf("invalid JSON for varietal, example of valid JSON:\n%s", js)
		}
	}

	body := &PickRequestBody{
		Name:     name,
		Winery:   winery,
		Varietal: varietal,
	}

	return body, nil
}
