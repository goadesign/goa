package goa

import (
	"fmt"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	expected := fmt.Sprintf("v%d.%d.%d", Major, Minor, Build)
	if Suffix != "" {
		expected += "-" + Suffix
	}
	if got := Version(); got != expected {
		t.Errorf("invalid version format, %s", got)
	}
}

func TestCompatible(t *testing.T) {
	t.Run("well-formed", func(t *testing.T) {
		testdata := []string{
			Version(),
			fmt.Sprintf("v%d.0.0", Major),
			fmt.Sprintf("v%d.1.2", Major),
			fmt.Sprintf("v%d.10.10", Major),
			fmt.Sprintf("v%d.10.10-wip", Major),
			fmt.Sprintf("v%d.10.10-test", Major),
		}
		for i, v := range testdata {
			ok, err := Compatible(v)
			if err != nil {
				t.Errorf("unexpected error, %v", err)
			}
			if !ok {
				t.Errorf("expected false, but true, %d:%v", i, v)
			}
		}
	})
	t.Run("invalid version string format", func(t *testing.T) {
		testdata := []string{
			"v1",
			"v1.",
			"v..",
			"v1..",
			"v1.2",
			"v1.2.",
			"v1.2..",
		}
		for i, v := range testdata {
			_, err := Compatible(v)
			if err == nil {
				t.Errorf("expected error, but nil, %d:%v", i, v)
				continue
			}
			if !strings.HasPrefix(err.Error(), "invalid version string format ") {
				t.Errorf("unexpected error, %d:%v, %v", i, v, err)
			}
		}
	})
	t.Run("different major version number", func(t *testing.T) {
		testdata := []string{
			fmt.Sprintf("v%d.0.0", Major-1),
			fmt.Sprintf("v%d.0.0", Major+1),
		}
		for i, v := range testdata {
			ok, err := Compatible(v)
			if err != nil {
				t.Errorf("unexpected error, %d:%v, %v", i, v, err)
				continue
			}
			if ok {
				t.Errorf("expected false, but true, %d:%v", i, v)
			}
		}
	})
}
