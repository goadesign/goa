package genapp

import (
	"fmt"
	"sort"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
)

// BuildEncoders builds the template data needed to render the given encoding definitions.
// This extra map is needed to handle the case where a single encoding definition maps to multiple
// encoding packages. The data is indexed by mime type.
func BuildEncoders(info []*design.EncodingDefinition, encoder bool) ([]*EncoderTemplateData, error) {
	if len(info) == 0 {
		return nil, nil
	}
	// knownStdPackages lists the stdlib packages known by BuildEncoders
	var knownStdPackages = map[string]string{
		"encoding/json": "json",
		"encoding/xml":  "xml",
		"encoding/gob":  "gob",
	}
	encs := normalizeEncodingDefinitions(info)
	data := make([]*EncoderTemplateData, len(encs))
	defaultMediaType := info[0].MIMETypes[0]
	for i, enc := range encs {
		var pkgName string
		if name, ok := knownStdPackages[enc.PackagePath]; ok {
			pkgName = name
		} else {
			srcPath, err := codegen.PackageSourcePath(enc.PackagePath)
			if err != nil {
				return nil, fmt.Errorf("failed to locate package source of %s (%s)",
					enc.PackagePath, err)
			}
			pkgName, err = codegen.PackageName(srcPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load package %s (%s)",
					enc.PackagePath, err)
			}
		}
		isDefault := false
		for _, m := range enc.MIMETypes {
			if m == defaultMediaType {
				isDefault = true
			}
		}
		d := &EncoderTemplateData{
			PackagePath: enc.PackagePath,
			PackageName: pkgName,
			Function:    enc.Function,
			MIMETypes:   enc.MIMETypes,
			Default:     isDefault,
		}
		data[i] = d
	}
	return data, nil
}

// normalizeEncodingDefinitions figures out the package path and function of all encoding
// definitions and groups them by package and function name.
// We're going for simple rather than efficient (this is codegen after all)
// Also we assume that the encoding definitions have been validated: they have at least
// one mime type and definitions with no package path use known encoders.
func normalizeEncodingDefinitions(defs []*design.EncodingDefinition) []*design.EncodingDefinition {
	// First splat all definitions so each only have one mime type
	var encs []*design.EncodingDefinition
	for _, enc := range defs {
		if len(enc.MIMETypes) == 1 {
			encs = append(encs, enc)
			continue
		}
		for _, m := range enc.MIMETypes {
			encs = append(encs, &design.EncodingDefinition{
				MIMETypes:   []string{m},
				PackagePath: enc.PackagePath,
				Function:    enc.Function,
				Encoder:     enc.Encoder,
			})
		}
	}

	// Next make sure all definitions have a package path
	for _, enc := range encs {
		if enc.PackagePath == "" {
			mt := enc.MIMETypes[0]
			enc.PackagePath = design.KnownEncoders[mt]
			idx := 0
			if !enc.Encoder {
				idx = 1
			}
			enc.Function = design.KnownEncoderFunctions[mt][idx]
		} else if enc.Function == "" {
			if enc.Encoder {
				enc.Function = "NewEncoder"
			} else {
				enc.Function = "NewDecoder"
			}
		}
	}

	// Regroup by package and function name
	byfn := make(map[string][]*design.EncodingDefinition)
	var first string
	for _, enc := range encs {
		key := enc.PackagePath + "#" + enc.Function
		if first == "" {
			first = key
		}
		if _, ok := byfn[key]; ok {
			byfn[key] = append(byfn[key], enc)
		} else {
			byfn[key] = []*design.EncodingDefinition{enc}
		}
	}

	// Reserialize into array keeping the first element identical since it's the default
	// encoder.
	return serialize(byfn, first)
}

func serialize(byfn map[string][]*design.EncodingDefinition, first string) []*design.EncodingDefinition {
	res := make([]*design.EncodingDefinition, len(byfn))
	i := 0
	keys := make([]string, len(byfn))
	for k := range byfn {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	var idx int
	for j, k := range keys {
		if k == first {
			idx = j
			break
		}
	}
	keys[0], keys[idx] = keys[idx], keys[0]
	i = 0
	for _, key := range keys {
		encs := byfn[key]
		res[i] = &design.EncodingDefinition{
			MIMETypes:   encs[0].MIMETypes,
			PackagePath: encs[0].PackagePath,
			Function:    encs[0].Function,
		}
		if len(encs) > 0 {
			encs = encs[1:]
			for _, enc := range encs {
				for _, m := range enc.MIMETypes {
					found := false
					for _, rm := range res[i].MIMETypes {
						if m == rm {
							found = true
							break
						}
					}
					if !found {
						res[i].MIMETypes = append(res[i].MIMETypes, m)
					}
				}
			}
		}
		i++
	}
	return res
}
