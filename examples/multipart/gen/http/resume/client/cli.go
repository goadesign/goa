// Code generated by goa v2.0.0-wip, DO NOT EDIT.
//
// resume HTTP client CLI support package
//
// Command:
// $ goa gen goa.design/goa/examples/multipart/design -o
// $(GOPATH)/src/goa.design/goa/examples/multipart

package client

import (
	"encoding/json"
	"fmt"

	resume "goa.design/goa/examples/multipart/gen/resume"
)

// BuildAddPayload builds the payload for the resume add endpoint from CLI
// flags.
func BuildAddPayload(resumeAddBody string) ([]*resume.Resume, error) {
	var err error
	var body []*ResumeRequestBody
	{
		err = json.Unmarshal([]byte(resumeAddBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, example of valid JSON:\n%s", "'[\n      {\n         \"education\": [\n            {\n               \"institution\": \"Error accusantium autem asperiores reprehenderit dolorem.\",\n               \"major\": \"Id ipsam consectetur omnis earum eos.\"\n            },\n            {\n               \"institution\": \"Error accusantium autem asperiores reprehenderit dolorem.\",\n               \"major\": \"Id ipsam consectetur omnis earum eos.\"\n            }\n         ],\n         \"experience\": [\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            },\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            },\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            },\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            }\n         ],\n         \"name\": \"Similique ipsum enim voluptas.\"\n      },\n      {\n         \"education\": [\n            {\n               \"institution\": \"Error accusantium autem asperiores reprehenderit dolorem.\",\n               \"major\": \"Id ipsam consectetur omnis earum eos.\"\n            },\n            {\n               \"institution\": \"Error accusantium autem asperiores reprehenderit dolorem.\",\n               \"major\": \"Id ipsam consectetur omnis earum eos.\"\n            }\n         ],\n         \"experience\": [\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            },\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            },\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            },\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            }\n         ],\n         \"name\": \"Similique ipsum enim voluptas.\"\n      },\n      {\n         \"education\": [\n            {\n               \"institution\": \"Error accusantium autem asperiores reprehenderit dolorem.\",\n               \"major\": \"Id ipsam consectetur omnis earum eos.\"\n            },\n            {\n               \"institution\": \"Error accusantium autem asperiores reprehenderit dolorem.\",\n               \"major\": \"Id ipsam consectetur omnis earum eos.\"\n            }\n         ],\n         \"experience\": [\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            },\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            },\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            },\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            }\n         ],\n         \"name\": \"Similique ipsum enim voluptas.\"\n      },\n      {\n         \"education\": [\n            {\n               \"institution\": \"Error accusantium autem asperiores reprehenderit dolorem.\",\n               \"major\": \"Id ipsam consectetur omnis earum eos.\"\n            },\n            {\n               \"institution\": \"Error accusantium autem asperiores reprehenderit dolorem.\",\n               \"major\": \"Id ipsam consectetur omnis earum eos.\"\n            }\n         ],\n         \"experience\": [\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            },\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            },\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            },\n            {\n               \"company\": \"Qui repellat eveniet dignissimos et reiciendis.\",\n               \"duration\": 9027220452237938752,\n               \"role\": \"Quia hic.\"\n            }\n         ],\n         \"name\": \"Similique ipsum enim voluptas.\"\n      }\n   ]'")
		}
	}
	if err != nil {
		return nil, err
	}
	v := make([]*resume.Resume, len(body))
	for i, val := range body {
		v[i] = &resume.Resume{
			Name: val.Name,
		}
		if val.Experience != nil {
			v[i].Experience = make([]*resume.Experience, len(val.Experience))
			for j, val := range val.Experience {
				v[i].Experience[j] = &resume.Experience{
					Company:  val.Company,
					Role:     val.Role,
					Duration: val.Duration,
				}
			}
		}
		if val.Education != nil {
			v[i].Education = make([]*resume.Education, len(val.Education))
			for j, val := range val.Education {
				v[i].Education[j] = &resume.Education{
					Institution: val.Institution,
					Major:       val.Major,
				}
			}
		}
	}
	return v, nil
}
