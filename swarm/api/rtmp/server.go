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
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	//"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"github.com/ethereum/go-ethereum/swarm/storage/streaming"
	streamingVizClient "github.com/ethereum/go-ethereum/swarm/streamingviz/client"

	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/format"
	"github.com/nareix/joy4/format/flv"
	joy4rtmp "github.com/nareix/joy4/format/rtmp"
)

//This is for flushing to http request handlers (joy4 concept)
type writeFlusher struct {
	httpflusher http.Flusher
	io.Writer
}

func (self writeFlusher) Flush() error {
	self.httpflusher.Flush()
	return nil
}

func init() {
	format.RegisterAll()
}

//Spin off a go routine that serves rtmp requests.  For now I think this only handles a single stream.
func StartRtmpServer(rtmpPort string, streamer *streaming.Streamer, forwarder storage.CloudStore, viz *streamingVizClient.Client) {
	if rtmpPort == "" {
		rtmpPort = "1935"
	}
	fmt.Println("Starting RTMP Server on port: ", rtmpPort)
	server := &joy4rtmp.Server{Addr: ":" + rtmpPort}

	server.HandlePlay = func(conn *joy4rtmp.Conn) {
		glog.V(logger.Info).Infof("Trying to play stream at %v", conn.URL)

		// Parse the streamID from the query param ?streamID=....
		strmID := conn.URL.Query()["streamID"][0]
		glog.V(logger.Info).Infof("Got streamID as %v", strmID)
		viz.LogConsume(strmID)
		stream, err := streamer.SubscribeToStream(strmID)

		if err != nil {
			glog.V(logger.Info).Infof("Error subscribing to stream %v", err)
			return
		}

		//Send subscribe request
		forwarder.Stream(strmID)

		//Copy chunks to outgoing connection
		go CopyFromChannel(conn, stream)
	}

	server.HandlePublish = func(conn *joy4rtmp.Conn) {
		// Create a new stream
		stream, _ := streamer.AddNewStream()
		glog.V(logger.Info).Infof("Added a new stream with id: %v", stream.ID)
		viz.LogBroadcast(string(stream.ID))

		//Send video to streamer channels
		go CopyToChannel(conn, stream)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("In handleFunc, Path: ", r.URL.Path)
		// glog.V(logger.Info).Infof("Trying to play stream at %v", conn.URL)

		// Parse the streamID from the query param ?streamID=....
		strmID := r.URL.Query()["streamID"][0]
		glog.V(logger.Info).Infof("Got streamID as %v", strmID)
		stream, err := streamer.SubscribeToStream(strmID)

		if err != nil {
			glog.V(logger.Info).Infof("Error subscribing to stream %v", err)
			return
		}

		//Send subscribe request
		forwarder.Stream(strmID)

		w.Header().Set("Content-Type", "video/x-flv")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.WriteHeader(200)
		flusher := w.(http.Flusher)
		flusher.Flush()

		muxer := flv.NewMuxerWriteFlusher(writeFlusher{httpflusher: flusher, Writer: w})
		//Cannot kick off a go routine here because the ResponseWriter is not a pointer (so a copy of the writer doesn't make any sense)
		CopyFromChannel(muxer, stream)
	})

	httpPortNum, _ := strconv.Atoi(rtmpPort)
	httpPort := strconv.Itoa(httpPortNum + 7000)
	go http.ListenAndServe(":"+httpPort, nil)
	server.ListenAndServe()
}

//Copy packets from channels in the streamer to our destination muxer
func CopyFromChannel(dst av.Muxer, stream *streaming.Stream) (err error) {
	chunk := <-stream.DstVideoChan
	// chunk := storage.ByteArrInVideoChunk(<-streamer.ByteArrChan)
	if err = dst.WriteHeader(chunk.HeaderStreams); err != nil {
		fmt.Println("Error writing header copying from channel")
		return
	}

	for {
		select {
		case chunk := <-stream.DstVideoChan:
			// fmt.Println("Copying from channel")
			if chunk.ID == streaming.EOFStreamMsgID {
				fmt.Println("Copying EOF from channel")
				err := dst.WriteTrailer()
				if err != nil {
					fmt.Println("Error writing trailer: ", err)
					return err
				}
			}
			err := dst.WritePacket(chunk.Packet)
			if err != nil {
				glog.V(logger.Error).Infof("Error writing packet to video player: %s", err)
				return err
			}
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
func CopyToChannel(src av.Demuxer, stream *streaming.Stream) (err error) {
	var streams []av.CodecData
	if streams, err = src.Streams(); err != nil {
		return
	}
	if err = CopyPacketsToChannel(src, streams, stream); err != nil {
		return
	}
	return
}

func CopyPacketsToChannel(src av.PacketReader, headerStreams []av.CodecData, stream *streaming.Stream) (err error) {
	for seq := int64(0); ; seq++ {
		var pkt av.Packet
		if pkt, err = src.ReadPacket(); err != nil {
			if err == io.EOF {
				chunk := &streaming.VideoChunk{
					ID:            streaming.EOFStreamMsgID,
					Seq:           seq,
					HeaderStreams: headerStreams,
					Packet:        pkt,
				}
				stream.SrcVideoChan <- chunk
				fmt.Println("Done with packet reading: ", err)

				// Close the channel so that the protocol.go loop
				// reading from the channel doesn't block
				close(stream.SrcVideoChan)
				break
			}
			return
		}

		chunk := &streaming.VideoChunk{
			ID:            streaming.DeliverStreamMsgID,
			Seq:           seq,
			HeaderStreams: headerStreams,
			Packet:        pkt,
		}

		select {
		case stream.SrcVideoChan <- chunk:
			if chunk.Seq%100 == 0 {
				fmt.Printf("sent video chunk: %d\n", chunk.Seq)
			}
		default:
		}
	}
	glog.V(logger.Info).Infof("Returning from the copyPacketsToChannel thread")
	return
}
