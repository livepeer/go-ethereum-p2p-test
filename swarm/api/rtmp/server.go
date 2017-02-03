//Adding the RTMP server.  This will put up a RTMP endpoint when starting up Swarm.
//It's a simple RTMP server that will take a video stream and play it right back out.
//After bringing up the Swarm node with RTMP enabled, try it out using:
//
//ffmpeg -re -i bunny.mp4 -c copy -f flv rtmp://localhost/movie
//ffplay rtmp://localhost/movie

package rtmp

import (
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/swarm/storage"
	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/format"
	joy4rtmp "github.com/nareix/joy4/format/rtmp"
)

func init() {
	format.RegisterAll()
}

//Spin off a go routine that serves rtmp requests.  For now I think this only handles a single stream.
func StartRtmpServer(rtmpPort string, streamer *storage.Streamer, forwarder storage.CloudStore) {
	if rtmpPort == "" {
		rtmpPort = "1935"
	}
	fmt.Println("Starting RTMP Server on port: ", rtmpPort)
	server := &joy4rtmp.Server{Addr: ":" + rtmpPort}
	server.HandlePlay = func(conn *joy4rtmp.Conn) {
		//Send subscribe request
		key := []byte("teststream")
		forwarder.Stream(key)

		//Copy chunks to outgoing connection
		go CopyFromChannel(conn, streamer)
	}

	server.HandlePublish = func(conn *joy4rtmp.Conn) {
		//Send video to streamer channels
		go CopyToChannel(conn, streamer)
	}

	server.ListenAndServe()
}

//Copy packets from channels in the streamer to our destination muxer
func CopyFromChannel(dst av.Muxer, streamer *storage.Streamer) (err error) {
	chunk := <-streamer.DstVideoChan
	// chunk := storage.ByteArrInVideoChunk(<-streamer.ByteArrChan)
	if err = dst.WriteHeader(chunk.HeaderStreams); err != nil {
		fmt.Println("Error writing header copying from channel")
		return
	}

	for {
		select {
		case chunk := <-streamer.DstVideoChan:
			// fmt.Println("Copying from channel")
			if chunk.ID == 300 {
				fmt.Println("Copying EOF from channel")
				err := dst.WriteTrailer()
				if err != nil {
					fmt.Println("Error writing trailer: ", err)
				}
			}
			dst.WritePacket(chunk.Packet)
			// This doesn't work because default will just end the stream too quickly.
			// There is a design trade-off here: if we want the stream to automatically continue after some kind of
			// interruption, then we cannot end the stream.  Maybe we can do it after like... 10 mins of inactivity,
			// but it's quite common for livestream sources to have some difficulties and stop for minutes at a time.
			// default:
			// 	fmt.Println("CopyFromChannel Finished")
			// 	return
		}
	}
}

//Copy packets from our source demuxer to the streamer channels.  For now we put the header in every packet.  We can
//optimize for packet size later.
func CopyToChannel(src av.Demuxer, streamer *storage.Streamer) (err error) {
	var streams []av.CodecData
	if streams, err = src.Streams(); err != nil {
		return
	}

	for {
		var pkt av.Packet
		if pkt, err = src.ReadPacket(); err != nil {
			if err == io.EOF {
				chunk := &storage.VideoChunk{
					ID:            300,
					HeaderStreams: streams,
					Packet:        pkt,
				}
				streamer.SrcVideoChan <- chunk
				fmt.Println("Done with packet reading: %s", err)
				break
			}
			return
		}

		chunk := &storage.VideoChunk{
			ID:            200,
			HeaderStreams: streams,
			Packet:        pkt,
		}

		select {
		case streamer.SrcVideoChan <- chunk:
			fmt.Println("sent video chunk")
		default:
		}
	}

	return
}
