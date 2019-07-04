package testdata

import (
	. "goa.design/goa/v3/dsl"
)

var ConflictWithAPINameAndServiceNameDSL = func() {
	var _ = API("aloha", func() {
		Title("conflict with API name and service names")
	})
	var _ = Service("aloha", func() {})     // same as API name
	var _ = Service("alohaapi", func() {})  // API name + 'api' suffix
	var _ = Service("alohaapi1", func() {}) // API name + 'api' suffix + sequential no.
}

var ConflictWithGoifiedAPINameAndServiceNamesDSL = func() {
	var _ = API("good-by", func() {
		Title("conflict with goified API name and goified service names")
	})
	var _ = Service("good-by-", func() {})      // Goify name is same as API name
	var _ = Service("good-by-api", func() {})   // API name + 'api' suffix
	var _ = Service("good-by-api-1", func() {}) // API name + 'api' suffix + sequential no.
}
