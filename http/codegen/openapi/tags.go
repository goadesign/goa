package openapi

import (
	"fmt"
	"sort"
	"strings"

	"goa.design/goa/v3/expr"
)

// Tag allows adding meta data to a single tag that is used by the Operation Object. It is
// not mandatory to have a Tag Object per tag used there.
type Tag struct {
	// Name of the tag.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Description is a short description of the tag.
	// GFM syntax can be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// ExternalDocs is additional external documentation for this tag.
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	// Extensions defines the OpenAPI extensions.
	Extensions map[string]interface{} `json:"-" yaml:"-"`
}

// TagsFromExpr extracts the OpenAPI related metadata from the given expression.
func TagsFromExpr(mdata expr.MetaExpr) (tags []*Tag) {
	var keys []string
	for k := range mdata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		chunks := strings.Split(key, ":")
		if len(chunks) != 3 {
			continue
		}
		if (chunks[0] != "swagger" && chunks[0] != "openapi") || chunks[1] != "tag" {
			continue
		}

		tag := &Tag{Name: chunks[2]}

		mdata[key] = mdata[fmt.Sprintf("%s:desc", key)]
		if len(mdata[key]) != 0 {
			tag.Description = mdata[key][0]
		}

		hasDocs := false
		docs := &ExternalDocs{}

		mdata[key] = mdata[fmt.Sprintf("%s:url", key)]
		if len(mdata[key]) != 0 {
			docs.URL = mdata[key][0]
			hasDocs = true
		}

		mdata[key] = mdata[fmt.Sprintf("%s:url:desc", key)]
		if len(mdata[key]) != 0 {
			docs.Description = mdata[key][0]
			hasDocs = true
		}

		if hasDocs {
			tag.ExternalDocs = docs
		}

		extensionsPrefix := fmt.Sprintf("%s:extension:", key)
		tag.Extensions = extensionsFromExprWithPrefix(mdata, extensionsPrefix)

		tags = append(tags, tag)
	}

	return
}

// TagNamesFromExpr computes the names of the OpenAPI tags specified in the
// given metadata expressions.
func TagNamesFromExpr(mdatas ...expr.MetaExpr) (tagNames []string) {
	for _, mdata := range mdatas {
		tags := TagsFromExpr(mdata)
		for _, tag := range tags {
			tagNames = append(tagNames, tag.Name)
		}
	}
	return
}

type _tag Tag

// MarshalJSON returns the JSON encoding of t.
func (t Tag) MarshalJSON() ([]byte, error) {
	return MarshalJSON(_tag(t), t.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (t Tag) MarshalYAML() (interface{}, error) {
	return MarshalYAML(_tag(t), t.Extensions)
}
