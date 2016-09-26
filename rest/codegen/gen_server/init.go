package genserver

import "github.com/goadesign/goa/goagen/codegen"

type (
	// InitWriter generate the service setup code.
	InitWriter struct {
		*codegen.SourceFile
	}

	// EncoderTemplateData contains the data needed to render the registration code for a single
	// encoder or decoder package.
	EncoderTemplateData struct {
		// PackagePath is the Go package path to the package implmenting the encoder/decoder.
		PackagePath string
		// PackageName is the name of the Go package implementing the encoder/decoder.
		PackageName string
		// Function is the name of the package function implementing the decoder/encoder factory.
		Function string
		// MIMETypes is the list of supported MIME types.
		MIMETypes []string
		// Default is true if this encoder/decoder should be set as the default.
		Default bool
	}
)

// NewInitWriter returns a service setup code writer.
func NewInitWriter(filename string) (*InitWriter, error) {
	file, err := codegen.SourceFileFor(filename)
	if err != nil {
		return nil, err
	}
	return &InitWriter{SourceFile: file}, nil
}

// WriteInitService writes the initService function
func (w *InitWriter) Write(encoders, decoders []*EncoderTemplateData) error {
	ctx := map[string]interface{}{
		"Encoders": encoders,
		"Decoders": decoders,
	}
	return w.ExecuteTemplate("initServer", initT, nil, ctx)
}

const (
	// initT generates the server initialization code.
	// template input: *ControllerTemplateData
	initT = `
// initServer sets up the server encoders, decoders and mux.
func initServer(server *res.HTTPServer) {
	// Setup encoders and decoders
{{ range .Encoders }}{{/*
*/}}	server.Encoder.Register({{ .PackageName }}.{{ .Function }}, "{{ join .MIMETypes "\", \"" }}")
{{ end }}{{ range .Decoders }}{{/*
*/}}	server.Decoder.Register({{ .PackageName }}.{{ .Function }}, "{{ join .MIMETypes "\", \"" }}")
{{ end }}

	// Setup default encoder and decoder
{{ range .Encoders }}{{ if .Default }}{{/*
*/}}	server.Encoder.Register({{ .PackageName }}.{{ .Function }}, "*/*")
{{ end }}{{ end }}{{ range .Decoders }}{{ if .Default }}{{/*
*/}}	server.Decoder.Register({{ .PackageName }}.{{ .Function }}, "*/*")
{{ end }}{{ end }}}
`
)
