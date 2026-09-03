
func {{ .VarPrefix }}UsageExamples() string {
  return {{ .CLIPkg }}.{{ .Parser.UsageExamples.Name }}()
}
