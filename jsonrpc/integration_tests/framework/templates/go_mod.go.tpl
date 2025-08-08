module testservice

go 1.21

require (
	goa.design/goa/v3 v3.19.2
	goa.design/clue v1.2.0
)

replace goa.design/goa/v3 => {{ .GoaPath }}