// Messages for the livecoin protocol

package network

import (
	"fmt"
)

const (
	handshakeMsg = iota
	publishVideoMsg
	requestVideoMsg
)

type handshakeMsgData struct {
	ID       string
	Greeting string
}

func (self *handshakeMsgData) String() string {
	return fmt.Sprintf("%v says: %v", self.ID, self.Greeting)
}

type publishVideoMsgData struct {
	ID  uint64
	URL string
}

type requestVideoMsgData struct {
	ID  uint64
	URL string
}
