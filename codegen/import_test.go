package codegen

import (
	"path/filepath"
	"testing"
)

func TestImportPaths(t *testing.T) {
	var i0 = SimpleImport("simple")
	var i1 = NewImport("renamed", "direct")
	var i2 = NewImport("sub", filepath.Join("base", "sub"))
	var i3 = NewImport("remote", "example.com/something/remote")

	tests := []struct {
		i        *ImportSpec
		expected string
	}{
		{i0, `"simple"`},
		{i1, `renamed "direct"`},
		{i2, `sub "base/sub"`},
		{i3, `remote "example.com/something/remote"`},
	}

	for _, tc := range tests {
		t.Run(tc.i.Name, func(t *testing.T) {
			if tc.i.Code() != tc.expected {
				t.Errorf("unexpected import code: %q != %q", tc.i.Code(), tc.expected)
			}
		})
	}
}
