
syntax = {{ printf "%q" .ProtoVersion }};

package {{ .Pkg }};

option go_package = "/{{ .Pkg }}pb";
{{- range .Imports }}
import "{{ . }}";
{{- end }}
