package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	NodeID   string
	Endpoint string
}

func NewClient(nodeID string) *Client {
	return &Client{
		NodeID:   nodeID,
		Endpoint: "http://localhost:8585/event", // Default. Override if you'd like to change
	}
}

func (self *Client) LogPeers(peers []string) {
	data := self.initData("peers")
	data["peers"] = peers
	self.postEvent(data)
}

func (self *Client) LogBroadcast(streamID string) {
	data := self.initData("broadcast")
	data["streamId"] = streamID
	self.postEvent(data)
}

func (self *Client) LogConsume(streamID string) {
	data := self.initData("consume")
	data["streamId"] = streamID
	self.postEvent(data)
}

func (self *Client) LogRelay(streamID string) {
	data := self.initData("relay")
	data["streamId"] = streamID
	self.postEvent(data)
}

func (self *Client) LogDone(streamID string) {
	data := self.initData("done")
	data["streamId"] = streamID
	self.postEvent(data)
}

func (self *Client) initData(eventName string) (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["name"] = eventName
	data["node"] = self.NodeID
	return
}

func (self *Client) postEvent(data map[string]interface{}) {
	enc, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", self.Endpoint, bytes.NewBuffer(enc))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Couldn't connect to the event server", err)
		return
	}
	defer resp.Body.Close()
}
