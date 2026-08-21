package service

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	stest "goa.design/goa/v3/codegen/service/testdata"
)

// TestService_DedupEventMarkers verifies that when multiple streaming methods share the
// same result type the generated service code only emits a single event marker method.
func TestService_DedupEventMarkers(t *testing.T) {
	root := codegen.RunDSL(t, stest.StreamingDuplicateResultTypesDSL)
	services := mustServicesData(t, root)
	require.Len(t, root.Services, 1)

	files := Files("goa.design/goa/example", []*ServicesData{services})
	require.Greater(t, len(files), 0)

	// Generate the service.go content
	buf := new(bytes.Buffer)
	for _, s := range files[0].SectionTemplates[1:] {
		require.NoError(t, s.Write(buf))
	}
	code := buf.String()

	// Count occurrences of the service-level event marker method for SharedEvent
	// The marker has the shape: func (*SharedEvent) isdupStreamServiceEvent() {}
	occurrences := strings.Count(code, "func (*SharedEvent) isdupStreamServiceEvent()")
	require.Equal(t, 1, occurrences, "expected a single event marker for SharedEvent, got %d", occurrences)
}
