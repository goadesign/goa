package codegen

import (
	"reflect"
	"testing"
)

func TestRegisterPlugin(t *testing.T) {
	var (
		p1 = &plugin{name: "abc"}
		p2 = &plugin{name: "def"}

		pf1 = &plugin{name: "abc", first: true}

		pl1 = &plugin{name: "abc", last: true}

		pIns = &plugin{name: "cde"}
	)
	tests := []struct {
		name       string
		existingPs []*plugin
		expectedPs []*plugin
	}{
		{"no-plugins", []*plugin{}, []*plugin{pIns}},
		{"plugins-without-first", []*plugin{p1}, []*plugin{p1, pIns}},
		{"plugins-with-first", []*plugin{pf1, p2}, []*plugin{pf1, pIns, p2}},
		{"plugins-with-same-name", []*plugin{pf1, pIns, p2}, []*plugin{pf1, pIns, pIns, p2}},
		{"plugins-with-last", []*plugin{pf1, pl1}, []*plugin{pf1, pIns, pl1}},
		{"mixed", []*plugin{pf1, p1, p2}, []*plugin{pf1, p1, pIns, p2}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plugins = tc.existingPs
			RegisterPlugin(pIns.name, "", nil, nil)
			if !reflect.DeepEqual(plugins, tc.expectedPs) {
				t.Errorf("invalid plugin registration order")
			}
		})
	}
}

func TestRegisterPluginFirst(t *testing.T) {
	var (
		p1 = &plugin{name: "abc"}
		p2 = &plugin{name: "def"}

		pf1 = &plugin{name: "abc", first: true}
		pf2 = &plugin{name: "def", first: true}

		pl1 = &plugin{name: "abc", last: true}

		pIns = &plugin{name: "cde", first: true}
	)
	tests := []struct {
		name       string
		existingPs []*plugin
		expectedPs []*plugin
	}{
		{"no-plugins", []*plugin{}, []*plugin{pIns}},
		{"plugins-without-first", []*plugin{p1, p2}, []*plugin{pIns, p1, p2}},
		{"plugins-with-first", []*plugin{pf1, pf2}, []*plugin{pf1, pIns, pf2}},
		{"plugins-with-same-name", []*plugin{pf1, pIns}, []*plugin{pf1, pIns, pIns}},
		{"plugins-with-last", []*plugin{pf1, pl1}, []*plugin{pf1, pIns, pl1}},
		{"mixed", []*plugin{pf1, pf2, p1, p2}, []*plugin{pf1, pIns, pf2, p1, p2}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plugins = tc.existingPs
			RegisterPluginFirst(pIns.name, "", nil, nil)
			if !reflect.DeepEqual(plugins, tc.expectedPs) {
				t.Errorf("invalid plugin registration order")
			}
		})
	}
}

func TestRegisterPluginLast(t *testing.T) {
	var (
		p1 = &plugin{name: "abc"}
		p2 = &plugin{name: "def"}

		pl1 = &plugin{name: "abc", last: true}
		pl2 = &plugin{name: "def", last: true}

		pf1 = &plugin{name: "abc", first: true}

		pIns = &plugin{name: "cde", last: true}
	)
	tests := []struct {
		name       string
		existingPs []*plugin
		expectedPs []*plugin
	}{
		{"no-plugins", []*plugin{}, []*plugin{pIns}},
		{"plugins-without-last", []*plugin{p1, p2}, []*plugin{p1, p2, pIns}},
		{"plugins-with-last", []*plugin{pl1, pl2}, []*plugin{pl1, pIns, pl2}},
		{"plugins-with-same-name", []*plugin{pl1, pIns}, []*plugin{pl1, pIns, pIns}},
		{"plugins-with-first", []*plugin{pf1, pl2}, []*plugin{pf1, pIns, pl2}},
		{"mixed", []*plugin{pf1, p1, p2, pl1, pl2}, []*plugin{pf1, p1, p2, pl1, pIns, pl2}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plugins = tc.existingPs
			RegisterPluginLast(pIns.name, "", nil, nil)
			if !reflect.DeepEqual(plugins, tc.expectedPs) {
				t.Errorf("invalid plugin registration order")
			}
		})
	}
}
