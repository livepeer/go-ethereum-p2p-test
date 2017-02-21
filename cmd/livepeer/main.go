package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	gitCommit string
	app       = utils.NewApp(gitCommit, "Livepeer")
)

var (
	LivepeerPortFlag = cli.StringFlag{
		Name:  "port",
		Usage: "Port that the local RTMP server is running on",
		Value: "1935",
	}
)

func init() {
	app.Action = livepeer
	app.HideVersion = true
	app.Copyright = "Copyright 2017 Doug and Eric"
	app.Flags = []cli.Flag{
		LivepeerPortFlag,
	}
	app.Commands = []cli.Command{
		{
			Action:      stream,
			Name:        "stream",
			Usage:       "Connect to a live stream by id",
			ArgsUsage:   " <streamID>",
			Description: "This command will use ffplay to play the given stream ID from the Livepeer network",
		},
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func livepeer(ctx *cli.Context) error {
	fmt.Println("Run livepeer stream --port <port> <streamID>")
	return nil
}

func stream(ctx *cli.Context) error {
	port := ctx.GlobalString(LivepeerPortFlag.Name)
	streamID := ctx.Args()[0]
	rtmpURL := fmt.Sprintf("rtmp://localhost:%v/stream/%v", port, streamID)

	cmd := exec.Command("ffplay", rtmpURL)
	err := cmd.Start()
	if err != nil {
		fmt.Println("Couldn't start the stream")
		os.Exit(1)
	}
	fmt.Println("Now streaming")
	err = cmd.Wait()
	fmt.Println("Finished the stream")
	return nil
}
