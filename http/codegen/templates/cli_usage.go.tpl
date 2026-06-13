
func {{ .VarPrefix }}UsageCommands() []string {
  return cli.UsageCommands()
}

func {{ .VarPrefix }}UsageExamples() string {
  return cli.UsageExamples()
}
