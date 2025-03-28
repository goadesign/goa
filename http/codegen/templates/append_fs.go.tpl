// appendFS is a custom implementation of fs.FS that appends a specified prefix
// to the file paths before delegating the Open call to the underlying fs.FS.
type appendFS struct {
	prefix string
	fs     http.FileSystem
}

// Open opens the named file, appending the prefix to the file path before
// passing it to the underlying fs.FS.
func (s appendFS) Open(name string) (http.File, error) {
	switch name {
	{{- range $requested, $embedded := . }}
	case {{ printf "%q" $requested }}:
		name = {{ printf "%q" $embedded }}
	{{- end }}
	}
	return s.fs.Open(path.Join(s.prefix, name))
}

// appendPrefix returns a new fs.FS that appends the specified prefix to file paths
// before delegating to the provided embed.FS.
func appendPrefix(fsys http.FileSystem, prefix string) http.FileSystem {
	return appendFS{prefix: prefix, fs: fsys}
}
