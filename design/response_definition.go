package design

import "regexp"

var (
	Ok = ResponseDefinition{Status: 200
// An action response definition
// Specifies a status, a media type and header definitions
// Header definitions consist of exact values or validating regular expressions
type ResponseDefinition struct {
	Status         int            // Response status code
	MediaType      *MediaType     // Response media type if any
	HeaderPatterns HeaderPatterns // Response headers
}

// Response header patterns
type HeaderPatterns map[string]*regexp.Regexp
