package client

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// BuildAddPayloadFromFlags constructs a "add" endpoint payload from command
// line flag values.
func BuildAddPayloadFromFlags(nameFlag, wineryFlag, vintageFlag, compositionFlag, descriptionFlag, ratingFlag string) (*AddRequestBody, error) {
	var winery WineryRequestBody
	{
		err := json.Unmarshal([]byte(wineryFlag), &winery)
		if err != nil {
			ex := WineryRequestBody{} // ...
			js, _ := json.Marshal(ex)
			return nil, fmt.Errorf("invalid JSON for winery, example of valid JSON:\n%s", js)
		}
	}

	var composition []*ComponentRequestBody
	if compositionFlag != "" {
		err := json.Unmarshal([]byte(compositionFlag), &composition)
		if err != nil {
			ex := []*ComponentRequestBody{} // ...
			js, _ := json.Marshal(ex)
			return nil, fmt.Errorf("invalid JSON for composition, example of valid JSON:\n%s", string(js))
		}
	}

	var vintage uint32
	{
		if v, err := strconv.ParseUint(ratingFlag, 10, 32); err == nil {
			vintage = uint32(v)
		}
	}

	var rating *uint32
	if ratingFlag != "" {
		if v, err := strconv.ParseUint(ratingFlag, 10, 32); err == nil {
			val := uint32(v)
			rating = &val
		}
	}

	var description *string
	if descriptionFlag != "" {
		description = &descriptionFlag
	}

	body := &AddRequestBody{
		Name:        nameFlag,
		Winery:      &winery,
		Vintage:     vintage,
		Composition: composition,
		Description: description,
		Rating:      rating,
	}

	return body, nil
}
