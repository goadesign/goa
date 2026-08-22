
func {{ .VarPrefix }}UsageCommands() []string {
  return {{ .CLIPkg }}.{{ .Parser.UsageCommands.Name }}()
}

func {{ .VarPrefix }}UsageExamples() string {
  return {{ .CLIPkg }}.{{ .Parser.UsageExamples.Name }}()
}
