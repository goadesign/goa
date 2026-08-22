// UsageExamples produces an example of a valid invocation of the CLI tool.
func {{ .Declaration.Name }}() string {
	return {{ range .Examples }}os.Args[0] + " " + {{ printf "%q" . }} + "\n" +
	{{ end }}""
}
