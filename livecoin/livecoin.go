package livecoin

import (
	"time"

	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
)

type Livecoin struct {
	msg string
	running chan bool
	quit chan bool
}

// creates a new livecoin service instance
// implements node.Service
func NewLivecoin() (self *Livecoin, err error) {
	self = &Livecoin{
		msg: "Welcome to Livecoin",
		running: make(chan bool),
		quit: make(chan bool),
	}

	return
}

func (self *Livecoin) Start(net *p2p.Server) error {
	glog.V(logger.Info).Infoln("Starting the Livecoin Service")
	go self.livecoinLoop()
	go self.longloop()
	return nil
}

func (self *Livecoin) Stop() error {
	return nil
}

func (self *Livecoin) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "lvc",
			Version: "0.1",
			Service: &Info{},
			Public: true,
		},
	}
}

func (self *Livecoin) Protocols() []p2p.Protocol {
	proto, _ := Lvc()
	return []p2p.Protocol{proto}
}


const (
	ProtocolVersion            = 0
	ProtocolLength     = uint64(8)
	ProtocolMaxMsgSize = 10 * 1024 * 1024
	ProtocolNetworkId          = 65
)

// The main protocol entrypoint. The Run: function will get invoked when peers connect
func Lvc() (p2p.Protocol, error) {
	return p2p.Protocol{
		Name: "lvc",
		Version: ProtocolVersion,
		Length: ProtocolLength,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			return runLvc(p, rw)
		},
	}, nil
}

func runLvc(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	glog.V(logger.Info).Infoln("Got a new peer:", p.LocalAddr(), p.RemoteAddr())
	return nil
}
	

func (self *Livecoin) livecoinLoop() {
	for {
		glog.V(logger.Info).Infoln("In the livecoinloop")
		select {
		case <- self.running:
			glog.V(logger.Info).Infoln("Livecoin says", self.msg)
		case <- self.quit:
			return
		}
	}
}

func (self *Livecoin) longloop() {
	for i:= 0; i < 100; i++ {
		time.Sleep(1 * time.Second)
		self.running <- true
	}
	self.quit <- true
}

type Info struct {
}

func (self *Info) Add() (int, error) {
	return 0, nil
}
