package openapi

import (
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
	Extensions map[string]any `json:"-" yaml:"-"`
}

// TagsFromExpr extracts the OpenAPI related metadata from the given expression.
func TagsFromExpr(mdata expr.MetaExpr) (tags []*Tag) {
	var keys []string
	for k := range mdata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		chunks := strings.SplitN(key, ":", 4)
		if len(chunks) < 3 {
			continue
		}
		if (chunks[0] != "swagger" && chunks[0] != "openapi") || chunks[1] != "tag" {
			continue
		}

		name := chunks[2]
		var tag *Tag
		for _, t := range tags {
			if t.Name == name {
				tag = t
				break
			}
		}
		if tag == nil {
			tag = &Tag{Name: chunks[2]}
			tags = append(tags, tag)
		}
		if len(chunks) == 4 {
			switch chunks[3] {
			case "desc":
				tag.Description = mdata[key][0]
			case "url":
				if tag.ExternalDocs == nil {
					tag.ExternalDocs = &ExternalDocs{}
				}
				tag.ExternalDocs.URL = mdata[key][0]
			case "url:desc":
				if tag.ExternalDocs == nil {
					tag.ExternalDocs = &ExternalDocs{}
				}
				tag.ExternalDocs.Description = mdata[key][0]
			default:
				idx := strings.Index(key, "extension:")
				if idx == -1 {
					continue
				}
				tag.Extensions = extensionsFromExprWithPrefix(mdata, key[:idx+10])
			}
		}
	}

	return
}

// TagNamesFromExpr computes the names of the OpenAPI tags specified in the
// given metadata expressions.
func TagNamesFromExpr(mdata expr.MetaExpr) (tagNames []string) {
	tags := TagsFromExpr(mdata)
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}
	return
}

type _tag Tag

// MarshalJSON returns the JSON encoding of t.
func (t Tag) MarshalJSON() ([]byte, error) {
	return MarshalJSON(_tag(t), t.Extensions)
}

// MarshalYAML returns value which marshaled in place of the original value
func (t Tag) MarshalYAML() (any, error) {
	return MarshalYAML(_tag(t), t.Extensions)
}
