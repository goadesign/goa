{{ printf "%s service example implementation.\nThe example methods log the requests and return zero values." .Name | comment }}
type {{ .VarName }}srvc struct {
	logger *log.Logger
}
