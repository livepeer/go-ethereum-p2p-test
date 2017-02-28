package network

import "github.com/ethereum/go-ethereum/swarm/storage/streaming"

type StreamDB struct {
	DownstreamRequesters        map[streaming.StreamID][]*peer
	UpstreamTranscodeRequesters map[streaming.StreamID]*peer
}

func NewStreamDB() *StreamDB {
	return &StreamDB{
		DownstreamRequesters:        make(map[streaming.StreamID][]*peer),
		UpstreamTranscodeRequesters: make(map[streaming.StreamID]*peer),
	}
}

func (self *StreamDB) AddDownstreamPeer(streamID streaming.StreamID, p *peer) {
	self.DownstreamRequesters[streamID] = append(self.DownstreamRequesters[streamID], p)
}

func (self *StreamDB) AddUpstreamTranscodeRequester(transcodeID streaming.StreamID, p *peer) {
	self.UpstreamTranscodeRequesters[transcodeID] = p
}
