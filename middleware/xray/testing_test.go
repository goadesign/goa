package xray

import (
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"
)

// ReadUDP verifies that exactly the expected number of messages are received.
func ReadUDP(t *testing.T, udplisten string, expectedMessages int, sender func()) []string {
	var (
		readChan = make(chan []string)
		msg      = make([]byte, 1024*32)
	)
	resAddr, err := net.ResolveUDPAddr("udp", udplisten)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.ListenUDP("udp", resAddr)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		listener.SetReadDeadline(time.Now().Add(time.Second))
		listener.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		var messages []string
		for {
			n, _, err := listener.ReadFrom(msg)
			if err != nil {
				if !strings.HasSuffix(err.Error(), "i/o timeout") {
					t.Errorf("expected final timeout error but got: %s", err)
				}
				break // we're done
			}
			messages = append(messages, string(msg[0:n]))
		}
		if len(messages) != expectedMessages {
			t.Errorf("unexpected number of messages, expected %d got %d. All messages:\n%s",
				expectedMessages, len(messages), strings.Join(messages, "\n"))
		}
		readChan <- messages
	}()

	sender()

	defer func() {
		if err := listener.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	return <-readChan
}

// ExtractSegment returns the unmarshalled segment JSON from a ReadUDP response.
func ExtractSegment(t *testing.T, js string) *Segment {
	t.Helper()

	var s *Segment
	elems := strings.Split(js, "\n")
	if len(elems) != 2 {
		t.Fatalf("invalid number of lines, expected 2 got %d: %v", len(elems), elems)
	}
	if elems[0] != UDPHeader[:len(UDPHeader)-1] {
		t.Errorf("invalid header, got %s", elems[0])
	}
	err := json.Unmarshal([]byte(elems[1]), &s)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
