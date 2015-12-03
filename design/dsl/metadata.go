package dsl

import (
	"fmt"

	"github.com/raphael/goa/design"
)

// Metadata is a key/value pair that can be assigned
// to an object.  The value must be a JSON string.
// Metadata is not currently used in standard generation but may be
// used by user-defined generators.
//	 Metadata("creator", `{"name":"goagen"}`)
func Metadata(name string, value string) {
	fmt.Println("ut metadata call")
	var uparent *design.AttributeDefinition
	if at, ok := attributeDefinition(false); ok {
		fmt.Println("IS ATTRIBUTE")
		uparent = at
	}
	if uparent != nil {
		fmt.Println("uparent not nil")
		if uparent.Type == nil {
			uparent.Type = design.Object{}
		}
		if _, ok := uparent.Type.(design.Object); !ok {
			ReportError("can't define metadata on attribute of type %s", uparent.Type.Name())
			return
		}

		var baseMeta *design.MetadataDefinition

		baseMeta = &design.MetadataDefinition{
			Name:  name,
			Value: value,
		}
		uparent.Metadata = baseMeta
		fmt.Println("assigned metadata:", uparent.Metadata)
		return
	}
	fmt.Println("mt metadata call")
	var parent *design.MediaTypeDefinition
	if mt, ok := mediaTypeDefinition(true); ok {
		parent = mt
		fmt.Println("IS MT")
	}
	fmt.Println(" - parent", parent)
	if parent != nil {
		if parent.Type == nil {
			parent.Type = design.Object{}
		}
		if _, ok := parent.Type.(design.Object); !ok {
			ReportError("can't define metadata on attribute of type %s", parent.Type.Name())
			return
		}

		var baseMeta *design.MetadataDefinition

		baseMeta = &design.MetadataDefinition{
			Name:  name,
			Value: value,
		}

		parent.Metadata = baseMeta
		return
	}
	return
}
