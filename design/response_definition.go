package design

import "regexp"

// An action response definition
// Specifies a status, a media type and header patterns
type ResponseDefinition struct {
	Status         int                  // Response status code
	MediaType      *MediaTypeDefinition // Response media type if any
	HeaderPatterns HeaderPatterns       // Response headers
}

// Response header patterns
type HeaderPatterns map[string]*regexp.Regexp
