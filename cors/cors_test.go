package cors_test

import (
	"regexp"
	"testing"

	"github.com/goadesign/goa/cors"
)

func TestMatchOrigin(t *testing.T) {
	data := []struct {
		Origin string
		Spec   string
		Result bool
	}{
		{"http://example.com", "*", true},
		{"http://example.com", "http://example.com", true},
		{"http://example.com", "https://example.com", false},
		{"http://test.example.com", "*.example.com", true},
		{"http://test.example.com:80", "*.example.com", false},
		{"http://test.example.com:80", "http://test.example.com*", true},
	}

	for _, test := range data {
		result := cors.MatchOrigin(test.Origin, test.Spec)
		if result != test.Result {
			t.Errorf("cors.MatchOrigin(%s, %s) should return %t", test.Origin, test.Spec, test.Result)
		}
	}
}

func TestMatchOriginRegexp(t *testing.T) {
	data := []struct {
		Origin string
		Spec   string
		Result bool
	}{
		{"http://test.example.com:80", "(.*).example.com(.*)", true},
		{"http://test.example.com:80", ".*.example.com.*", true},
		{"http://test.example.com:80", ".*.other.com.*", false},
		{"http://test.example.com", "[test|swag].example.com", true},
		{"http://swag.example.com", "[test|swag].example.com", true},
		{"http://other.example.com", "[test|swag].example.com", false},
		{"http://other.example.com", "[test|swag].other.com", false},
	}

	for _, test := range data {
		result := cors.MatchOriginRegexp(test.Origin, regexp.MustCompile(test.Spec))
		if result != test.Result {
			t.Errorf("cors.MatchOrigin(%s, %s) should return %t", test.Origin, test.Spec, test.Result)
		}
	}
}
