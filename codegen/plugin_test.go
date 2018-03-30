package codegen

import (
	"reflect"
	"testing"
)

func TestRegisterPlugin(t *testing.T) {
	var (
		abcP = &plugin{name: "abc"}
		defP = &plugin{name: "def"}

		abcPFirst = &plugin{name: "abc", first: true}

		abcPLast = &plugin{name: "abc", last: true}

		pToInsert = &plugin{name: "cde"}
	)
	tests := []struct {
		name       string
		existingPs []*plugin
		expectedPs []*plugin
	}{
		{"no-plugins", []*plugin{}, []*plugin{pToInsert}},
		{"plugins-without-first", []*plugin{abcP}, []*plugin{abcP, pToInsert}},
		{"plugins-with-first", []*plugin{abcPFirst, defP}, []*plugin{abcPFirst, pToInsert, defP}},
		{"plugins-with-same-name", []*plugin{abcPFirst, pToInsert, defP}, []*plugin{abcPFirst, pToInsert, pToInsert, defP}},
		{"plugins-with-last", []*plugin{abcPFirst, abcPLast}, []*plugin{abcPFirst, pToInsert, abcPLast}},
		{"mixed", []*plugin{abcPFirst, abcP, defP}, []*plugin{abcPFirst, abcP, pToInsert, defP}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plugins = tc.existingPs
			RegisterPlugin(pToInsert.name, "", nil)
			if !reflect.DeepEqual(plugins, tc.expectedPs) {
				t.Errorf("invalid plugin registration order")
			}
		})
	}
}

func TestRegisterPluginFirst(t *testing.T) {
	var (
		abcP = &plugin{name: "abc"}
		defP = &plugin{name: "def"}

		abcPFirst = &plugin{name: "abc", first: true}
		defPFirst = &plugin{name: "def", first: true}

		abcPLast = &plugin{name: "abc", last: true}

		pToInsert = &plugin{name: "cde", first: true}
	)
	tests := []struct {
		name       string
		existingPs []*plugin
		expectedPs []*plugin
	}{
		{"no-plugins", []*plugin{}, []*plugin{pToInsert}},
		{"plugins-without-first", []*plugin{abcP, defP}, []*plugin{pToInsert, abcP, defP}},
		{"plugins-with-first", []*plugin{abcPFirst, defPFirst}, []*plugin{abcPFirst, pToInsert, defPFirst}},
		{"plugins-with-same-name", []*plugin{abcPFirst, pToInsert}, []*plugin{abcPFirst, pToInsert, pToInsert}},
		{"plugins-with-last", []*plugin{abcPFirst, abcPLast}, []*plugin{abcPFirst, pToInsert, abcPLast}},
		{"mixed", []*plugin{abcPFirst, defPFirst, abcP, defP}, []*plugin{abcPFirst, pToInsert, defPFirst, abcP, defP}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plugins = tc.existingPs
			RegisterPluginFirst(pToInsert.name, "", nil)
			if !reflect.DeepEqual(plugins, tc.expectedPs) {
				t.Errorf("invalid plugin registration order")
			}
		})
	}
}

func TestRegisterPluginLast(t *testing.T) {
	var (
		abcP = &plugin{name: "abc"}
		defP = &plugin{name: "def"}

		abcPLast = &plugin{name: "abc", last: true}
		defPLast = &plugin{name: "def", last: true}

		abcPFirst = &plugin{name: "abc", first: true}

		pToInsert = &plugin{name: "cde", last: true}
	)
	tests := []struct {
		name       string
		existingPs []*plugin
		expectedPs []*plugin
	}{
		{"no-plugins", []*plugin{}, []*plugin{pToInsert}},
		{"plugins-without-last", []*plugin{abcP, defP}, []*plugin{abcP, defP, pToInsert}},
		{"plugins-with-last", []*plugin{abcPLast, defPLast}, []*plugin{abcPLast, pToInsert, defPLast}},
		{"plugins-with-same-name", []*plugin{abcPLast, pToInsert}, []*plugin{abcPLast, pToInsert, pToInsert}},
		{"plugins-with-first", []*plugin{abcPFirst, defPLast}, []*plugin{abcPFirst, pToInsert, defPLast}},
		{"mixed", []*plugin{abcPFirst, abcP, defP, abcPLast, defPLast}, []*plugin{abcPFirst, abcP, defP, abcPLast, pToInsert, defPLast}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plugins = tc.existingPs
			RegisterPluginLast(pToInsert.name, "", nil)
			if !reflect.DeepEqual(plugins, tc.expectedPs) {
				t.Errorf("invalid plugin registration order")
			}
		})
	}
}
