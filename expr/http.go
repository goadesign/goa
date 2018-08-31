package expr

import "regexp"

// HTTPWildcardRegex is the regular expression used to capture path
// parameters.
var HTTPWildcardRegex = regexp.MustCompile(`/{\*?([a-zA-Z0-9_]+)}`)

// ExtractHTTPWildcards returns the names of the wildcards that appear in
// a HTTP path.
func ExtractHTTPWildcards(path string) []string {
	matches := HTTPWildcardRegex.FindAllStringSubmatch(path, -1)
	wcs := make([]string, len(matches))
	for i, m := range matches {
		wcs[i] = m[1]
	}
	return wcs
}
