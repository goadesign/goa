// This file verifies X-Ray UDP test daemon port allocation, segment capture,
// and backward-compatible ReadUDP execution.
package xraytest_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/middleware/xray"
	"goa.design/goa/v3/middleware/xray/xraytest"
)

// TestDaemonReadAndExtractSegment verifies that NewDaemon binds a free
// localhost UDP port and captures emitted X-Ray segments without port races.
func TestDaemonReadAndExtractSegment(t *testing.T) {
	daemon := xraytest.NewDaemon(t)
	conn, err := net.Dial("udp", daemon.Addr)
	require.NoError(t, err)
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("close connection: %v", closeErr)
		}
	}()

	seg := xray.NewSegment("svc", xray.NewTraceID(), xray.NewID(), conn)
	messages := daemon.Read(t, 1, func() {
		seg.Close()
	})
	require.Len(t, messages, 1)
	decoded := xraytest.ExtractSegment(t, messages[0])
	assert.Equal(t, "svc", decoded.Name)
	assert.False(t, decoded.InProgress)
}

// TestReadUDPDeprecatedZeroMessages verifies that deprecated ReadUDP remains
// functional on an ephemeral localhost address.
func TestReadUDPDeprecatedZeroMessages(t *testing.T) {
	messages := xraytest.ReadUDP(t, "127.0.0.1:0", 0, func() {})
	assert.Empty(t, messages)
}
