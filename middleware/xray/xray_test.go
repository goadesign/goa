package xray

import (
	"context"
	"errors"
	"net"
	"regexp"
	"sync"
	"testing"
	"time"
)

func TestNewID(t *testing.T) {
	id := NewID()
	if len(id) != 16 {
		t.Errorf("invalid ID length, expected 16 got %d", len(id))
	}
	if !regexp.MustCompile("[0-9a-f]{16}").MatchString(id) {
		t.Errorf("invalid ID format, should be hexadecimal, got %s", id)
	}
	if id == NewID() {
		t.Errorf("ids not unique")
	}
}

func TestNewTraceID(t *testing.T) {
	id := NewTraceID()
	if len(id) != 35 {
		t.Errorf("invalid ID length, expected 35 got %d", len(id))
	}
	if !regexp.MustCompile("1-[0-9a-f]{8}-[0-9a-f]{16}").MatchString(id) {
		t.Errorf("invalid Trace ID format, got %s", id)
	}
	if id == NewTraceID() {
		t.Errorf("trace ids not unique")
	}
}

func TestConnect(t *testing.T) {
	t.Run("dial fails, returns error immediately", func(t *testing.T) {
		dialErr := errors.New("dialErr")
		_, err := Connect(context.Background(), time.Millisecond, func() (net.Conn, error) {
			return nil, dialErr
		})
		if err != dialErr {
			t.Fatalf("Unexpected err, got %q, expected %q", err, dialErr)
		}
	})
	t.Run("connection gets replaced by new one", func(t *testing.T) {
		var (
			firstConn  = &net.UDPConn{}
			secondConn = &net.UnixConn{}
			callCount  = 0
		)
		wgCheckFirstConnection := sync.WaitGroup{}
		wgCheckFirstConnection.Add(1)
		wgThirdDial := sync.WaitGroup{}
		wgThirdDial.Add(1)
		dial := func() (net.Conn, error) {
			callCount++
			if callCount == 1 {
				return firstConn, nil
			}
			wgCheckFirstConnection.Wait()
			if callCount == 3 {
				wgThirdDial.Done()
			}
			return secondConn, nil
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		conn, err := Connect(ctx, time.Millisecond, dial)
		if err != nil {
			t.Fatalf("Expected nil err but got: %v", err)
		}

		if c := conn(); c != firstConn {
			t.Fatalf("Unexpected first connection: got %#v, expected %#v", c, firstConn)
		}
		wgCheckFirstConnection.Done()

		// by the time the 3rd dial happens, we know conn() should be returning the second connection
		wgThirdDial.Wait()

		if c := conn(); c != secondConn {
			t.Fatalf("Unexpected second connection: got %#v, expected %#v", c, secondConn)
		}
	})
	t.Run("connection not replaced if dial errored", func(t *testing.T) {
		var (
			firstConn = &net.UDPConn{}
			callCount = 0
		)
		wgCheckFirstConnection := sync.WaitGroup{}
		wgCheckFirstConnection.Add(1)
		wgThirdDial := sync.WaitGroup{}
		wgThirdDial.Add(1)
		dial := func() (net.Conn, error) {
			callCount++
			if callCount == 1 {
				return firstConn, nil
			}
			wgCheckFirstConnection.Wait()
			if callCount == 3 {
				wgThirdDial.Done()
			}
			return nil, errors.New("dialErr")
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		conn, err := Connect(ctx, time.Millisecond, dial)
		if err != nil {
			t.Fatalf("Expected nil err but got: %v", err)
		}

		if c := conn(); c != firstConn {
			t.Fatalf("Unexpected first connection: got %#v, expected %#v", c, firstConn)
		}
		wgCheckFirstConnection.Done()

		// by the time the 3rd dial happens, we know the second dial was processed, and shouldn't have replaced conn()
		wgThirdDial.Wait()

		if c := conn(); c != firstConn {
			t.Fatalf("Connection unexpectedly replaced: got %#v, expected %#v", c, firstConn)
		}
	})
}
