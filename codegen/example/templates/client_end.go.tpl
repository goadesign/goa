
}

{{- if .WritesEndpointResult }}
// writeEndpointResult calls one normal endpoint and writes its result as JSON.
func writeEndpointResult(ctx context.Context, stdout io.Writer, endpoint goa.Endpoint, payload any) error {
	data, err := endpoint(ctx, payload)
	if err != nil {
		return err
	}
	return writeJSON(stdout, data)
}
{{- end }}

{{- if .WritesStreamResults }}
// writeStreamResults writes each server result until the server ends the stream.
func writeStreamResults[T any](ctx context.Context, stdout io.Writer, recv func(context.Context) (T, error)) error {
	for {
		data, err := recv(ctx)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("receive result: %w", err)
		}
		if err := writeJSON(stdout, data); err != nil {
			return err
		}
	}
}
{{- end }}

{{- if or .WritesEndpointResult .WritesStreamResults }}
// writeJSON writes one indented JSON value followed by a newline.
func writeJSON(stdout io.Writer, data any) error {
	if data == nil {
		return nil
	}
	encoded, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("encode result: %w", err)
	}
	if _, err := fmt.Fprintln(stdout, string(encoded)); err != nil {
		return fmt.Errorf("write result: %w", err)
	}
	return nil
}
{{- end }}
