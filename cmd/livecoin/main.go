package main

import (
	"fmt"
	"os"
	"runtime"
	
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/internal/debug"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/livecoin"

	"gopkg.in/urfave/cli.v1"
)

const (
	clientIdentifier = "livecoin"
	versionString    = "0.1"
)

var (
	gitCommit string // Git SHA1 commit hash of the release (set via linker flags)
	app = utils.NewApp(gitCommit, "Livecoin")
	testBootNodes = []string {
		"enode://c5bf45b4acbe4d4fc6c06758ce642862396abdf0c7a18bce9bbf1d709af67f9b94f8de50d4d45600c8f5d1db4dfd2d2708648b922cd7bc76eaf74ef4f8d85e99@127.0.0.1:63450",
	}
)

var (
	LivecoinPortFlag = cli.StringFlag{
		Name: "livecoinport",
		Usage: "Livecoin local http api port",
	}
	LivecoinNetworkIdFlag = cli.IntFlag{
		Name: "livecoinnetworkid",
		Usage: "Network identifier (integer, default 65=livecoin privatenet)",
		Value: livecoin.NetworkId,
	}
)

func init() {
	app.Action = livecoinfunc
	app.HideVersion = true
	app.Copyright = "Copyright 2017 Doug and Eric"

	// Didn't bother setting up the app.Commands since there is only one

	app.Flags = []cli.Flag {
		LivecoinPortFlag,
		LivecoinNetworkIdFlag,
		utils.DataDirFlag,
		utils.ListenPortFlag,
	}
	app.Flags = append(app.Flags, debug.Flags...)
	app.Before = func(ctx *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return debug.Setup(ctx)
	}
	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func livecoinfunc(ctx *cli.Context) error {
	glog.V(logger.Info).Infoln("Livecoinfunc invoked")

	stack := utils.MakeNode(ctx, clientIdentifier, gitCommit)
	registerLivecoinService(ctx, stack)
	utils.StartNode(stack)
	networkId := ctx.GlobalUint64(LivecoinNetworkIdFlag.Name)

	// Add bootnodes as initial peers
	if (networkId == 65) {
		glog.V(logger.Info).Infoln("In network 65, injecting bootnodes")
		injectBootnodes(stack.Server(), testBootNodes)
	}

	stack.Wait()
	return nil
}

func registerLivecoinService(ctx *cli.Context, stack *node.Node) {
	boot := func(ctx *node.ServiceContext) (node.Service, error) {
		return livecoin.NewLivecoin()
	}

	if err := stack.Register(boot); err != nil {
		utils.Fatalf("Failed to register the Livecoin service: %v", err)
	}
}


func injectBootnodes(srv *p2p.Server, nodes []string) {
	for _, url := range nodes {
		n, err := discover.ParseNode(url)
		if err != nil {
			glog.Errorf("invalid bootnode %q", err)
			continue
		}
		srv.AddPeer(n)
	}
}
