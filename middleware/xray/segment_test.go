package xray_test

import (
	"net"
	"regexp"
	"testing"

	"goa.design/goa/v3/middleware/xray"
	"goa.design/goa/v3/middleware/xray/xraytest"
)

func TestSegment_NewSubsegment(t *testing.T) {
	daemon := xraytest.NewDaemon(t)
	conn, err := net.Dial("udp", daemon.Addr)
	if err != nil {
		t.Fatalf("failed to connect to daemon - %s", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("close connection: %v", closeErr)
		}
	}()
	var (
		name   = "sub"
		s      = xray.NewSegment("", "", "", conn)
		before = s.StartTime
		ss     = s.NewSubsegment(name)
	)
	if ss.ID == "" {
		t.Errorf("subsegment ID not initialized")
	}
	if !regexp.MustCompile("[0-9a-f]{16}").MatchString(ss.ID) {
		t.Errorf("invalid subsegment ID, got %v", ss.ID)
	}
	if ss.Name != name {
		t.Errorf("invalid subsegment name, expected %s got %s", name, ss.Name)
	}
	if ss.StartTime < before {
		t.Errorf("invalid subsegment StartAt, expected at least %v, got %v", before, ss.StartTime)
	}
	if !ss.InProgress {
		t.Errorf("subsegment not in progress")
	}
	if ss.Parent != s {
		t.Errorf("invalid subsegment parent, expected %v, got %v", s, ss.Parent)
	}
}

func TestSegment_SubmitInProgress(t *testing.T) {
	t.Run("call twice then close -- second call is ignored", func(t *testing.T) {
		daemon := xraytest.NewDaemon(t)
		conn, err := net.Dial("udp", daemon.Addr)
		if err != nil {
			t.Fatalf("failed to connect to daemon - %s", err)
		}
		defer func() {
			if closeErr := conn.Close(); closeErr != nil {
				t.Errorf("close connection: %v", closeErr)
			}
		}()

		segment := xray.NewSegment("hello", xray.NewTraceID(), xray.NewID(), conn)

		// call SubmitInProgress() twice, then Close it
		messages := daemon.Read(t, 2, func() {
			segment.Namespace = "1"
			segment.SubmitInProgress()
			segment.Namespace = "2"
			segment.SubmitInProgress() // should have no effect
			segment.Namespace = "3"
			segment.Close()
		})

		// verify the In-Progress segment
		s := xraytest.ExtractSegment(t, messages[0])
		if !s.InProgress {
			t.Errorf("expected segment to be InProgress, but it's not")
		}
		if s.Namespace != "1" {
			t.Errorf("unexpected segment namespace, expected %q got %q", "1", s.Namespace)
		}

		// verify the final segment (the second In-Progress segment would not have been sent)
		s = xraytest.ExtractSegment(t, messages[1])
		if s.InProgress {
			t.Errorf("expected segment to not be InProgress, but it is")
		}
		if s.Namespace != "3" {
			t.Errorf("unexpected segment namespace, expected %q got %q", "3", s.Namespace)
		}
	})

	t.Run("calling after already Closed -- no effect", func(t *testing.T) {
		daemon := xraytest.NewDaemon(t)
		conn, err := net.Dial("udp", daemon.Addr)
		if err != nil {
			t.Fatalf("failed to connect to daemon - %s", err)
		}
		defer func() {
			if closeErr := conn.Close(); closeErr != nil {
				t.Errorf("close connection: %v", closeErr)
			}
		}()

		segment := xray.NewSegment("hello", xray.NewTraceID(), xray.NewID(), conn)

		// Close(), then call SubmitInProgress(), only expect 1 segment written
		messages := daemon.Read(t, 1, func() {
			segment.Namespace = "1"
			segment.Close()
			segment.Namespace = "2"
			segment.SubmitInProgress() // should have no effect
		})

		// verify the In-Progress segment
		s := xraytest.ExtractSegment(t, messages[0])
		if s.InProgress {
			t.Errorf("expected segment to be closed, but it is still InProgress")
		}
		if s.Namespace != "1" {
			t.Errorf("unexpected segment namespace, expected %q got %q", "1", s.Namespace)
		}
	})
}
