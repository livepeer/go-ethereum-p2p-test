package storage

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/codec/aacparser"
	"github.com/nareix/joy4/codec/h264parser"
)

//The streamer brokers the video streams.
type Streamer struct {
	SrcVideoChan chan *VideoChunk
	DstVideoChan chan *VideoChunk
	ByteArrChan  chan []byte
}

func NewStreamer() (*Streamer, error) {
	return &Streamer{
		SrcVideoChan: make(chan *VideoChunk, 10),
		DstVideoChan: make(chan *VideoChunk, 10),
		ByteArrChan:  make(chan []byte),
	}, nil
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
