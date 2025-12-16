package eval

import "testing"

func TestNormalizeFileForPackageMatch(t *testing.T) {
	cases := map[string]struct {
		in   string
		want string
	}{
		"no version": {
			in:   "/home/me/src/goa/eval/error.go",
			want: "/home/me/src/goa/eval/error.go",
		},
		"module cache version": {
			in:   "/home/me/go/pkg/mod/goa.design/goa/v3@v3.23.2/dsl/result_type.go",
			want: "/home/me/go/pkg/mod/goa.design/goa/v3/dsl/result_type.go",
		},
		"multiple @ segments": {
			in:   "/home/me/go/pkg/mod/example.com/foo@v1.2.3/bar@v0.1.0/baz.go",
			want: "/home/me/go/pkg/mod/example.com/foo/bar/baz.go",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := normalizeFileForPackageMatch(tc.in)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
