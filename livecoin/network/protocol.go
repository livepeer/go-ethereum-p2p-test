// Protocol for the livecoin

package network

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/p2p"
)

const (
	ProtocolVersion    = 0
	ProtocolLength     = uint64(8)
	ProtocolMaxMsgSize = 10 * 1024 * 1024
	ProtocolNetworkId  = 65
)

type lvc struct {
	peer *p2p.Peer
	rw   p2p.MsgReadWriter
}

// The main protocol entrypoint. The Run: function will get invoked when peers connect
func Lvc() (p2p.Protocol, error) {
	glog.V(logger.Info).Infoln("Adding the Lvc protocol")
	return p2p.Protocol{
		Name:    "lvc",
		Version: ProtocolVersion,
		Length:  ProtocolLength,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			return runLvc(p, rw)
		},
	}, nil
}

func runLvc(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	glog.V(logger.Info).Infoln("Got a new peer:", p.String(), p.LocalAddr(), p.RemoteAddr())

	self := &lvc{
		peer: p,
		rw:   rw,
	}

	// Do the one time handshake
	err := self.handleHandshake()
	if err != nil {
		return err
	}

	// For now start sending some random publish and request video messages
	go self.sendRandomMessages()

	// Run a forever loop on this connection until an error is received
	for {
		err := self.handleAllMessages()
		if err != nil {
			return err
		}
	}

	return nil
}

func (self *lvc) handleAllMessages() error {
	msg, err := self.rw.ReadMsg()
	if err != nil {
		return err
	}

	defer msg.Discard()

	switch msg.Code {
	case handshakeMsg:
		return errors.New("Already handled a handshake from this peer.")

	case publishVideoMsg:
		var req publishVideoMsgData
		if err := msg.Decode(&req); err != nil {
			return errors.New("Couldn't decode publish video message.")
		}
		glog.V(logger.Info).Infof("Just received a publish video message", req.URL)

	case requestVideoMsg:
		var req requestVideoMsgData
		if err := msg.Decode(&req); err != nil {
			return errors.New("Couldn't decode request video message.")
		}
		glog.V(logger.Info).Infof("Just received a request video message", req.URL)
	default:
		return errors.New(fmt.Sprintf("Invalid message code %v", msg.Code))
	}

	return nil
}

func (self *lvc) handleHandshake() (err error) {
	greeting := fmt.Sprintf("Welcome to my house: %s", self.peer)
	handshake := &handshakeMsgData{
		ID:       "LIVE",
		Greeting: greeting,
	}

	err = p2p.Send(self.rw, handshakeMsg, handshake)
	if err != nil {
		return err
	}

	// Read their handshake
	var msg p2p.Msg
	msg, err = self.rw.ReadMsg()
	if err != nil {
		return err
	}

	if msg.Code != handshakeMsg {
		return errors.New("Should only receive a handshake message before receiving other messages")
	}

	var theirHandshake handshakeMsgData
	if err := msg.Decode(&theirHandshake); err != nil {
		return err
	}

	glog.V(logger.Info).Infof("Just received a livecoin handshake and the message was", theirHandshake.Greeting)
	return nil
}

func (self *lvc) sendRandomMessages() error {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 25; i++ {
		if i%2 == 0 {
			msgData := &publishVideoMsgData{
				ID:  uint64(i),
				URL: "http://youtube.com/videos/23",
			}
			err := p2p.Send(self.rw, publishVideoMsg, msgData)
			if err != nil {
				return err
			}
		} else {
			msgData := &requestVideoMsgData{
				ID:  uint64(i),
				URL: "http://vimeo.com/videos/3888",
			}
			err := p2p.Send(self.rw, requestVideoMsg, msgData)
			if err != nil {
				return err
			}
		}
		sleeptime := rand.Intn(10)
		time.Sleep(time.Duration(sleeptime) * time.Second)
	}
	return nil
}
