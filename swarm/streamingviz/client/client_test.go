package client

import (
	"testing"
)

// Need to be running a server on port 8585 to run the below.
// For now treat it as an example usage
func TestVizClient(t *testing.T) {
	client := NewClient("A")
	client.LogPeers([]string{"B", "C"})
	client.LogBroadcast("stream1")

	client2 := NewClient("B")
	client2.LogPeers([]string{"A", "D"})

	client3 := NewClient("D")
	client3.LogPeers([]string{"B"})
	client3.LogConsume("stream1")

	client2.LogRelay("stream1")
}
