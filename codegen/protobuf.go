// This file converts authored names into identifiers that Goa can safely write
// to protobuf files. Transport generators and external plugins use the same
// functions so identical design names produce identical protobuf names.
package codegen

import (
	"regexp"
	"strings"
)

var (
	protobufDigits = regexp.MustCompile("[0-9]+")

	protobufKeywords = map[string]struct{}{
		"bool": {}, "bytes": {}, "double": {}, "fixed32": {}, "fixed64": {},
		"float": {}, "int32": {}, "int64": {}, "sfixed32": {}, "sfixed64": {},
		"sint32": {}, "sint64": {}, "string": {}, "uint32": {}, "uint64": {},
		"enum": {}, "import": {}, "map": {}, "message": {}, "oneof": {},
		"option": {}, "package": {}, "public": {}, "repeated": {}, "reserved": {},
		"returns": {}, "rpc": {}, "service": {}, "syntax": {},
	}
)

// ProtobufName returns the identifier written for a protobuf message, service,
// or method. It keeps common acronyms uppercase and makes the first character
// legal for protobuf source.
func ProtobufName(name string) string {
	return protobufIdentifier(name, true, true)
}

// ProtobufFieldName returns the snake-case identifier written for a protobuf
// field or oneof. It makes the first character legal for protobuf source.
func ProtobufFieldName(name string) string {
	name = SnakeCase(protobufIdentifier(name, false, false))
	if _, reserved := protobufKeywords[name]; reserved {
		name += "_"
	}
	return name
}

// protobufIdentifier removes characters protobuf identifiers cannot contain
// and separates digits so the generated Go name matches protoc-gen-go.
func protobufIdentifier(name string, firstUpper, acronym bool) string {
	if index := strings.Index(name, ":"); index > 0 {
		name = name[:index]
	}
	name = strings.Map(func(character rune) rune {
		if character >= 'a' && character <= 'z' ||
			character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' ||
			character == '_' {
			return character
		}
		return '_'
	}, name)
	name = string(protobufDigits.ReplaceAllFunc([]byte(name), func(match []byte) []byte {
		result := make([]byte, len(match)+1)
		copy(result, match)
		result[len(result)-1] = '_'
		return result
	}))
	name = CamelCase(name, firstUpper, acronym)
	if name == "" {
		if firstUpper {
			return "Val"
		}
		return "val"
	}
	if name[0] >= '0' && name[0] <= '9' {
		name = "_" + name
	}
	return name
}
