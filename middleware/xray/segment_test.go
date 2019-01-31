package xray

import (
	"regexp"
	"sync"
	"testing"
)

func TestNewSubsegment(t *testing.T) {
	var (
		name   = "sub"
		s      = &Segment{Mutex: &sync.Mutex{}}
		before = now()
		ss     = s.NewSubsegment(name)
	)
	if s.counter != 1 {
		t.Errorf("counter not incremented after call to Subsegment")
	}
	if len(s.Subsegments) != 1 {
		t.Fatalf("invalid count of subsegments, expected 1 got %d", len(s.Subsegments))
	}
	if s.Subsegments[0] != ss {
		t.Errorf("invalid subsegments element, expected %v - got %v", name, s.Subsegments[0])
	}
	if ss.ID == "" {
		t.Errorf("subsegment ID not initialized")
	}
	if !regexp.MustCompile("[0-9a-f]{16}").MatchString(ss.ID) {
		t.Errorf("invalid subsegment ID, got %v", ss.ID)
	}
	if ss.Name != name {
		t.Errorf("invalid subsegemnt name, expected %s got %s", name, ss.Name)
	}
	if ss.StartTime < before {
		t.Errorf("invalid subsegment StartAt, expected at least %v, got %v", before, ss.StartTime)
	}
	if !ss.InProgress {
		t.Errorf("subsegemnt not in progress")
	}
	if ss.Parent != s {
		t.Errorf("invalid subsegment parent, expected %v, got %v", s, ss.Parent)
	}
}
