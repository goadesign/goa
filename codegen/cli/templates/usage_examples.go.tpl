// UsageExamples produces an example of a valid invocation of the CLI tool.
func UsageExamples() string {
	return {{ range . }}os.Args[0] + ` {{ . }}` + "\n" +
	{{ end }}""
}
