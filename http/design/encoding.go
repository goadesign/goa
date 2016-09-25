package design

type (
	// EncodingExpr defines an encoder supported by the API.
	EncodingExpr struct {
		// MIMETypes is the set of possible MIME types for the content being encoded or decoded.
		MIMETypes []string
		// PackagePath is the path to the Go package that implements the encoder/decoder.
		// The package must expose a `EncoderFactory` or `DecoderFactory` function
		// that the generated code calls. The methods must return objects that implement
		// the goa.EncoderFactory or goa.DecoderFactory interface respectively.
		PackagePath string
		// Function is the name of the Go function used to instantiate the encoder/decoder.
		// Defaults to NewEncoder and NewDecoder respecitively.
		Function string
		// Encoder is true if the definition is for a encoder, false if it's for a decoder.
		Encoder bool
	}
)
