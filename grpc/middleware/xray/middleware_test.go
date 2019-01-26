package xray

import (
	"testing"
)

const (
	// udp host:port used to run test server
	udplisten = "127.0.0.1:62111"
)

func TestNewUnaryServer(t *testing.T) {
	cases := map[string]struct {
		Daemon  string
		Success bool
	}{
		"ok":     {udplisten, true},
		"not-ok": {"1002.0.0.0:62111", false},
	}
	for k, c := range cases {
		m, err := NewUnaryServer("", c.Daemon)
		if err == nil && !c.Success {
			t.Errorf("%s: expected failure but err is nil", k)
		}
		if err != nil && c.Success {
			t.Errorf("%s: unexpected error %s", k, err)
		}
		if m == nil && c.Success {
			t.Errorf("%s: middleware is nil", k)
		}
	}
}

func TestNewStreamServer(t *testing.T) {
	cases := map[string]struct {
		Daemon  string
		Success bool
	}{
		"ok":     {udplisten, true},
		"not-ok": {"1002.0.0.0:62111", false},
	}
	for k, c := range cases {
		m, err := NewStreamServer("", c.Daemon)
		if err == nil && !c.Success {
			t.Errorf("%s: expected failure but err is nil", k)
		}
		if err != nil && c.Success {
			t.Errorf("%s: unexpected error %s", k, err)
		}
		if m == nil && c.Success {
			t.Errorf("%s: middleware is nil", k)
		}
	}
}
