package livecoin

import (
	"time"

	"github.com/ethereum/go-ethereum/livecoin/network"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
)

type Livecoin struct {
	msg     string
	running chan bool
	quit    chan bool
}

// creates a new livecoin service instance
// implements node.Service
func NewLivecoin() (self *Livecoin, err error) {
	self = &Livecoin{
		msg:     "Welcome to Livecoin",
		running: make(chan bool),
		quit:    make(chan bool),
	}

	return
}

func (self *Livecoin) Start(net *p2p.Server) error {
	glog.V(logger.Info).Infoln("Starting the Livecoin Service")
	go self.livecoinLoop(net)
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
			Version:   "0.1",
			Service:   &Info{},
			Public:    true,
		},
	}
}

func (self *Livecoin) Protocols() []p2p.Protocol {
	proto, _ := network.Lvc()
	return []p2p.Protocol{proto}
}

func (self *Livecoin) livecoinLoop(net *p2p.Server) {
	for {
		glog.V(logger.Info).Infoln("In the livecoinloop")
		select {
		case <-self.running:
			//glog.V(logger.Info).Infoln("Livecoin says", self.msg, net.Peers())
			glog.V(logger.Info).Infoln("Peers:")
			for _, val := range net.Peers() {
				glog.V(logger.Info).Infoln(val)
			}
		case <-self.quit:
			return
		}
	}
}

func (self *Livecoin) longloop() {
	for i := 0; i < 100; i++ {
		time.Sleep(2 * time.Second)
		self.running <- true
	}
	self.quit <- true
}

type Info struct {
}

func (self *Info) Add(x, y int) (int, error) {
	return x + y, nil
}
