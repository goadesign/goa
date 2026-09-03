// This package defines the Go names shared by design validation and code
// generation. Keeping the conversion here ensures authored Go values are
// checked against the same field names that generators use.
package codegenname

import (
	"go/doc"
	"go/token"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	commonInitialisms = map[string]bool{
		"API":   true,
		"ASCII": true,
		"CPU":   true,
		"CSS":   true,
		"DNS":   true,
		"EOF":   true,
		"GUID":  true,
		"HTML":  true,
		"HTTP":  true,
		"HTTPS": true,
		"ID":    true,
		"IP":    true,
		"JMES":  true,
		"JSON":  true,
		"JWT":   true,
		"LHS":   true,
		"OK":    true,
		"QPS":   true,
		"RAM":   true,
		"RHS":   true,
		"RPC":   true,
		"SDK":   true,
		"SLA":   true,
		"SMTP":  true,
		"SQL":   true,
		"SSH":   true,
		"TCP":   true,
		"TLS":   true,
		"TTL":   true,
		"UDP":   true,
		"UI":    true,
		"UID":   true,
		"UUID":  true,
		"URI":   true,
		"URL":   true,
		"UTF8":  true,
		"VM":    true,
		"XML":   true,
		"XSRF":  true,
		"XSS":   true,
	}
	reservedPackages = map[string]bool{
		"errors": true,
		"fmt":    true,
		"http":   true,
		"json":   true,
		"os":     true,
		"url":    true,
		"time":   true,
	}
)

// AttributeName returns the name selected by struct:field:name metadata, or
// the design name when the metadata does not provide an override.
func AttributeName(name string, fieldNames []string) string {
	if len(fieldNames) > 0 {
		return fieldNames[0]
	}
	return name
}

// CamelCase removes characters that cannot appear in identifiers and joins
// the remaining words while preserving common abbreviations such as ID and
// HTTP.
func CamelCase(name string, firstUpper, acronym bool) string {
	if name == "" {
		return ""
	}
	runes := removeTrailingInvalid([]rune(name))
	if len(runes) == 0 {
		return ""
	}

	wordStart, index := 0, 0
	for index+1 <= len(runes) {
		wordEnd := false
		runes = removeInvalidAtIndex(index, runes)
		switch {
		case index+1 == len(runes):
			wordEnd = true
		case !validIdentifier(runes[index]):
			runes = append(runes[:index], runes[index+1:]...)
		case runes[index+1] == '_':
			wordEnd = true
			count := 1
			for index+count+1 < len(runes) && runes[index+count+1] == '_' {
				count++
			}
			copy(runes[index+1:], runes[index+count+1:])
			runes = runes[:len(runes)-count]
		case isLower(runes[index]) && !isLower(runes[index+1]):
			wordEnd = true
		}
		index++
		if !wordEnd {
			continue
		}

		word := string(runes[wordStart:index])
		if upper := strings.ToUpper(word); commonInitialisms[upper] {
			switch {
			case firstUpper && acronym:
			case firstUpper && !acronym:
				upper = title(strings.ToLower(upper))
			case wordStart > 0 && !acronym:
				upper = title(strings.ToLower(upper))
			case wordStart == 0:
				upper = strings.ToLower(upper)
			}
			copy(runes[wordStart:], []rune(upper))
		} else if wordStart > 0 && strings.ToLower(word) == word {
			runes[wordStart] = unicode.ToUpper(runes[wordStart])
		} else if wordStart == 0 && strings.ToLower(word) == word && firstUpper {
			runes[wordStart] = unicode.ToUpper(runes[wordStart])
		}
		if wordStart == 0 && !firstUpper {
			runes[wordStart] = unicode.ToLower(runes[wordStart])
		}
		wordStart = index
	}
	return string(runes)
}

// Goify returns a valid Go identifier for name. A suffix after a colon names
// a transport field and does not take part in the Go identifier.
func Goify(name string, firstUpper bool) string {
	if name == "" {
		return ""
	}
	if index := strings.Index(name, ":"); index > 0 {
		name = name[:index]
	}
	name = CamelCase(name, firstUpper, true)
	if name == "" {
		if firstUpper {
			return "Val"
		}
		return "val"
	}
	return FixReserved(name)
}

// FixReserved appends an underscore when name is reserved by Go or commonly
// used by generated imports.
func FixReserved(name string) string {
	if doc.IsPredeclared(name) || token.IsKeyword(name) || reservedPackages[name] {
		return name + "_"
	}
	return name
}

// isLower reports whether a rune continues a lower-case word.
func isLower(value rune) bool {
	return unicode.IsDigit(value) || unicode.IsLower(value)
}

// validIdentifier reports whether a rune may appear after the first character
// of a Go identifier.
func validIdentifier(value rune) bool {
	return unicode.IsLetter(value) || unicode.IsDigit(value)
}

// removeTrailingInvalid removes characters that cannot end an identifier.
func removeTrailingInvalid(runes []rune) []rune {
	valid := len(runes) - 1
	for ; valid >= 0 && !validIdentifier(runes[valid]); valid-- {
	}
	return runes[:valid+1]
}

// removeInvalidAtIndex removes consecutive invalid characters at index.
func removeInvalidAtIndex(index int, runes []rune) []rune {
	valid := index
	for ; valid < len(runes) && !validIdentifier(runes[valid]); valid++ {
	}
	return append(runes[:index], runes[valid:]...)
}

// title capitalizes the first character of each word without changing the
// remaining characters.
func title(value string) string {
	return cases.Title(language.Und, cases.NoLower).String(value)
}
