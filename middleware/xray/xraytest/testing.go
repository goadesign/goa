// Package xraytest contains test helpers for package xray that are used by
// transport-specific X-Ray middleware tests.
package xraytest

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"goa.design/goa/v3/middleware/xray"
)

type (
	// Daemon owns a bound localhost UDP listener on an OS-assigned free port for
	// one X-Ray test or subtest. Keeping the socket bound across client dialing
	// and segment reads avoids port collisions and Windows reserved-port errors.
	Daemon struct {
		// Addr is the bound "127.0.0.1:<port>" address that X-Ray clients dial.
		Addr string
		// listener is the bound UDP socket receiving emitted X-Ray datagrams.
		listener *net.UDPConn
	}
)

// NewDaemon binds a UDP listener on 127.0.0.1:0 and closes it when t finishes.
func NewDaemon(t *testing.T) *Daemon {
	t.Helper()
	listener, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 0,
	})
	if err != nil {
		t.Fatalf("listen UDP on free port: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := listener.Close(); closeErr != nil && !errors.Is(closeErr, net.ErrClosed) {
			t.Errorf("close UDP test daemon: %v", closeErr)
		}
	})
	return &Daemon{
		Addr:     listener.LocalAddr().String(),
		listener: listener,
	}
}

// ReadUDP verifies that exactly the expected number of messages are received
// on udplisten. Prefer NewDaemon so the OS assigns an unreserved port and keeps
// the socket bound for the duration of the test.
func ReadUDP(t *testing.T, udplisten string, expectedMessages int, sender func()) []string {
	t.Helper()
	resAddr, err := net.ResolveUDPAddr("udp", udplisten)
	if err != nil {
		t.Fatal(err)
	}
	var listener *net.UDPConn
	for range 5 {
		listener, err = net.ListenUDP("udp", resAddr)
		if err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if listener == nil || err != nil {
		t.Fatal(err)
	}
	res := readUDPConn(t, listener, expectedMessages, sender)
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return res
}

// ExtractSegment returns the unmarshalled segment JSON from a ReadUDP response.
func ExtractSegment(t *testing.T, js string) *xray.Segment {
	t.Helper()

	var s *xray.Segment
	elems := strings.Split(js, "\n")
	if len(elems) != 2 {
		t.Fatalf("invalid number of lines, expected 2 got %d: %v", len(elems), elems)
	}
	if elems[0] != xray.UDPHeader[:len(xray.UDPHeader)-1] {
		t.Errorf("invalid header, got %s", elems[0])
	}
	err := json.Unmarshal([]byte(elems[1]), &s)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// Read runs sender and verifies that expectedMessages UDP datagrams arrive on d.
func (d *Daemon) Read(t *testing.T, expectedMessages int, sender func()) []string {
	t.Helper()
	return readUDPConn(t, d.listener, expectedMessages, sender)
}

// readUDPConn invokes sender while reading up to expectedMessages datagrams
// from listener. If sender exits early via t.Fatal or panic, the deferred
// drain cancels the read deadline and joins the reader goroutine before t exits.
func readUDPConn(t *testing.T, listener *net.UDPConn, expectedMessages int, sender func()) []string {
	t.Helper()
	var (
		readChan = make(chan []string, 1)
		msg      = make([]byte, 1024*32)
	)
	go func() {
		defer close(readChan)
		if err := listener.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			t.Error(err)
		}
		var messages []string
		for {
			n, _, err := listener.ReadFrom(msg)
			if err != nil {
				if !errors.Is(err, os.ErrDeadlineExceeded) &&
					!errors.Is(err, net.ErrClosed) &&
					!strings.HasSuffix(err.Error(), "i/o timeout") {
					t.Errorf("expected final timeout error but got: %s", err)
				}
				break
			}
			messages = append(messages, string(msg[0:n]))
			if len(messages) == expectedMessages {
				break
			}
		}
		if len(messages) != expectedMessages && !t.Failed() {
			t.Errorf("unexpected number of messages, expected %d got %d. All messages:\n%s",
				expectedMessages, len(messages), strings.Join(messages, "\n"))
		}
		readChan <- messages
	}()

	completed := false
	defer func() {
		if !completed {
			if err := listener.SetReadDeadline(time.Now()); err != nil && !errors.Is(err, net.ErrClosed) {
				t.Errorf("cancel UDP read deadline: %v", err)
			}
			<-readChan
		}
	}()

	sender()
	res := <-readChan
	completed = true
	return res
}
