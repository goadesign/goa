// UsageExamples produces an example of a valid invocation of the CLI tool.
func {{ usageName }}() string {
	return {{ range . }}os.Args[0] + " " + {{ printf "%q" . }} + "\n" +
	{{ end }}""
}
