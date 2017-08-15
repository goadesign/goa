package xray

import (
	"net"
	"time"
)

// TestNetConn is a mock net.Conn
type TestNetConn struct {
	*TestClientExpectation
}

// Make sure TestNetConn implements net.Conn
var _ (net.Conn) = (*TestNetConn)(nil)

// NewTestNetConn creates a mock net.Conn which uses expectations set by the
// tests to implement the behavior.
func NewTestNetConn() *TestNetConn {
	return &TestNetConn{NewTestClientExpectation()}
}

// Read runs any preset expectation.
func (c *TestNetConn) Read(b []byte) (n int, err error) {
	if e := c.Expectation("Read"); e != nil {
		return e.(func(b []byte) (n int, err error))(b)
	}
	return 0, nil
}

// Write runs any preset expectation.
func (c *TestNetConn) Write(b []byte) (n int, err error) {
	if e := c.Expectation("Write"); e != nil {
		return e.(func(b []byte) (n int, err error))(b)
	}
	return 0, nil
}

// Close runs any preset expectation.
func (c *TestNetConn) Close() error {
	if e := c.Expectation("Close"); e != nil {
		return e.(func() error)()
	}
	return nil
}

// LocalAddr runs any preset expectation.
func (c *TestNetConn) LocalAddr() net.Addr {
	if e := c.Expectation("LocalAddr"); e != nil {
		return e.(func() net.Addr)()
	}
	return nil
}

// RemoteAddr runs any preset expectation.
func (c *TestNetConn) RemoteAddr() net.Addr {
	if e := c.Expectation("RemoteAddr"); e != nil {
		return e.(func() net.Addr)()
	}
	return nil
}

// SetDeadline runs any preset expectation.
func (c *TestNetConn) SetDeadline(t time.Time) error {
	if e := c.Expectation("SetDeadline"); e != nil {
		return e.(func(t time.Time) error)(t)
	}
	return nil
}

// SetReadDeadline runs any preset expectation.
func (c *TestNetConn) SetReadDeadline(t time.Time) error {
	if e := c.Expectation("SetReadDeadline"); e != nil {
		return e.(func(t time.Time) error)(t)
	}
	return nil
}

// SetWriteDeadline runs any preset expectation.
func (c *TestNetConn) SetWriteDeadline(t time.Time) error {
	if e := c.Expectation("SetWriteDeadline"); e != nil {
		return e.(func(t time.Time) error)(t)
	}
	return nil
}
