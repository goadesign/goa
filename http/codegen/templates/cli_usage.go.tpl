
func {{ .VarPrefix }}UsageCommands() []string {
  return {{ .CLIPkg }}.UsageCommands()
}

func {{ .VarPrefix }}UsageExamples() string {
  return {{ .CLIPkg }}.UsageExamples()
}
