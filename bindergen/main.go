package main

func main() {
}

binderTmpl = `
func(r *http.Request, {{.Params}}) error {
	// validate request parameters
	// assign values to given parameters
	// order is fixed: payload (if defined), path params (in order), query params (in order)
}
`

// Will have to be recursive (members that are objects)
// Have to handle arrays
payloadStructTmpl = `
type {{.Name}} struct { {{range .Members}}
	{{.Name}} {{.Type}} // {{.Description}}
{{end}}}
`

