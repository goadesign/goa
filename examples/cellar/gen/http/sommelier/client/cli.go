// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// sommelier HTTP client CLI support package
//
// Command:
// $ goa gen goa.design/goa.v2/examples/cellar/design

package client

import (
	"encoding/json"
	"fmt"

	"goa.design/goa.v2/examples/cellar/gen/sommelier"
)

// BuildCriteria builds the payload for the sommelier pick endpoint from CLI
// flags.
func BuildCriteria(sommelierPickBody string) (*sommelier.Criteria, error) {
	var body PickRequestBody
	{
		err := json.Unmarshal([]byte(sommelierPickBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, example of valid JSON:\n%s", "{\"name\":\"Corrupti qui enim repellendus laboriosam accusamus.\",\"varietal\":[\"Aut reiciendis ea architecto magni tempora fugiat.\",\"Delectus vel earum doloribus.\",\"Consequuntur recusandae.\",\"Nemo omnis quae suscipit laudantium quis sapiente.\"],\"winery\":\"Sunt dolorem.\"}")
		}
	}
	v := &sommelier.Criteria{
		Name:   body.Name,
		Winery: body.Winery,
	}
	if body.Varietal != nil {
		v.Varietal = make([]string, len(body.Varietal))
		for i, val := range body.Varietal {
			v.Varietal[i] = val
		}
	}

	return v, nil
}
