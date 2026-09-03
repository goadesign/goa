{{ printf "%s adds a fixed directory to file paths before opening them." .AppendFSDeclaration.Name | comment }}
type {{ .AppendFSDeclaration.Name }} struct {
	prefix string
	fs     http.FileSystem
}

// Open opens the named file, appending the prefix to the file path before
// passing it to the underlying file system.
func (s {{ .AppendFSDeclaration.Name }}) Open(name string) (http.File, error) {
	switch name {
	{{- range $requested, $embedded := .Mappings }}
	case {{ printf "%q" $requested }}:
		name = {{ printf "%q" $embedded }}
	{{- end }}
	}
	return s.fs.Open(path.Join(s.prefix, name))
}

{{ printf "%s returns a file system that adds prefix before opening each path." .AppendPrefixDeclaration.Name | comment }}
func {{ .AppendPrefixDeclaration.Name }}(fsys http.FileSystem, prefix string) http.FileSystem {
	return {{ .AppendFSDeclaration.Name }}{prefix: prefix, fs: fsys}
}
