// Contains the metrics collected for LivePeer

package network

import (
	"github.com/ethereum/go-ethereum/metrics"
)

var (
	livepeerPacketSkipMeter   = metrics.NewMeter("livepeer/packets/skip")
	livepeerPacketBufferMeter = metrics.NewMeter("livepeer/packets/buffer")
	livepeerPacketReqTimer    = metrics.NewTimer("livepeer/packets/req")
	livepeerPacketInMeter     = metrics.NewMeter("livepeer/packets/in")

	livepeerStreamInMeter      = metrics.NewMeter("livepeer/streams/in")
	livepeerStreamTimeoutMeter = metrics.NewMeter("livepeer/streams/timeout")

	livepeerTestMeter = metrics.NewMeter("livepeer/test")
)
