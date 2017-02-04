package streaming

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/codec/aacparser"
	"github.com/nareix/joy4/codec/h264parser"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
)

// The ID for a stream, consists of the concatenation of the
// NodeID and a unique ID string of the
type StreamID string

func MakeStreamID(nodeID common.Hash, id string) StreamID {
	return StreamID(fmt.Sprintf("%x%v", nodeID[:], id))
}

// Given a stream ID, return it's origin nodeID and the unique stream ID
func (self *StreamID) SplitComponents() (common.Hash, string) {
	strStreamID := string(*self)
	originComponentLength := common.HashLength * 2 // 32 bytes == 64 hexadecimal digits
	return common.HexToHash(strStreamID[:originComponentLength]), strStreamID[originComponentLength:]
}

// A stream represents one stream
type Stream struct {
	SrcVideoChan chan *VideoChunk
	DstVideoChan chan *VideoChunk
	ByteArrChan  chan []byte
	ID           StreamID
}

// The streamer brookers the video streams
type Streamer struct {
	Streams     map[StreamID]*Stream
	SelfAddress common.Hash
}

func NewStreamer(selfAddress common.Hash) (*Streamer, error) {
	glog.V(logger.Info).Infof("Setting up new streamer with self address: %x", selfAddress[:])
	return &Streamer{
		Streams:     make(map[StreamID]*Stream),
		SelfAddress: selfAddress,
	}, nil
}

func (self *Streamer) SubscribeToStream(id string) (stream *Stream, err error) {
	streamID := StreamID(id) //MakeStreamID(nodeID, id)
	glog.V(logger.Info).Infof("Subscribing to stream with ID: %v", streamID)
	return self.saveStreamForId(streamID)
}

func (self *Streamer) AddNewStream() (stream *Stream, err error) {
	streamID := MakeStreamID(self.SelfAddress, "teststream")
	glog.V(logger.Info).Infof("Adding new stream with ID: %v", streamID)
	return self.saveStreamForId(streamID)
}

func (self *Streamer) saveStreamForId(streamID StreamID) (stream *Stream, err error) {
	if self.Streams[streamID] != nil {
		return nil, errors.New("Stream with this ID already exists")
	}

	self.Streams[streamID] = &Stream{
		SrcVideoChan: make(chan *VideoChunk, 10),
		DstVideoChan: make(chan *VideoChunk, 10),
		ByteArrChan:  make(chan []byte),
		ID:           streamID,
	}

	return self.Streams[streamID], nil
}

func (self *Streamer) GetStream(nodeID common.Hash, id string) (stream *Stream, err error) {
	return self.Streams[MakeStreamID(nodeID, id)], nil
}

func VideoChunkToByteArr(chunk VideoChunk) []byte {
	var buf bytes.Buffer
	gob.Register(VideoChunk{})
	gob.Register(h264parser.CodecData{})
	gob.Register(aacparser.CodecData{})
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(chunk)
	if err != nil {
		fmt.Println("Error converting bytearr to chunk: ", err)
	}
	return buf.Bytes()
}

func ByteArrInVideoChunk(arr []byte) VideoChunk {
	var buf bytes.Buffer
	gob.Register(VideoChunk{})
	gob.Register(h264parser.CodecData{})
	gob.Register(aacparser.CodecData{})
	gob.Register(av.Packet{})

	buf.Write(arr)
	var chunk VideoChunk
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(&chunk)
	if err != nil {
		fmt.Println("Error converting bytearr to chunk: ", err)
	}
	return chunk
}

func TestChunkEncoding(chunk VideoChunk) {
	// var buf bytes.Buffer
	// var newbuf bytes.Buffer
	// gob.Register(VideoChunk{})
	// gob.Register(h264parser.CodecData{})
	// gob.Register(aacparser.CodecData{})
	// enc := gob.NewEncoder(&buf)
	// dec := gob.NewDecoder(&newbuf)
	// err := enc.Encode(chunk)
	// if err != nil {
	// 	fmt.Println("Error converting bytearr to chunk: ", err)
	// }

	// newbuf.Write(buf.Bytes())
	// var newChunk VideoChunk
	// dec.Decode(&newChunk)

	// return buf.Bytes()
	bytes := VideoChunkToByteArr(chunk)
	newChunk := ByteArrInVideoChunk(bytes)
	fmt.Println("chunk: ", chunk)
	fmt.Println("newchunk: ", newChunk)
}
