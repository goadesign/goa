module testservice

go 1.25.0

require (
	goa.design/goa/v3 v3.19.2
)

replace goa.design/goa/v3 => {{ .GoaPath }}

// The first tidy runs before starter code imports Clue. Keep the compatible
// version pinned even when tidy removes its unused require directive.
replace goa.design/clue => goa.design/clue v1.2.6
