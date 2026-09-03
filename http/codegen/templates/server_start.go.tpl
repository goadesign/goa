{{ comment "handleHTTPServer starts configures and starts a HTTP server on the given URL. It shuts down the server if any error is received in the error channel." }}
func handleHTTPServer(ctx context.Context, u *url.URL{{ range .HandlerArgs }}, {{ .Name }} {{ if .Pointer }}*{{ end }}{{ .PkgName }}.{{ .TypeName }}{{ end }}, wg *sync.WaitGroup, errc chan error, dbg bool) {{ printf "{" }}{{ if not .JSONRPCServices }}
{{ end -}}
