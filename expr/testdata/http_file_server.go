package testdata

import (
	. "goa.design/goa/dsl"
)

var FilesValidDSL = func() {
	Service("files-dsl", func() {
		Files("path", "filename")
	})
}

var FilesIncompatibleDSL = func() {
	API("files-incompatile", func() {
		Files("path", "filename")
	})
}
