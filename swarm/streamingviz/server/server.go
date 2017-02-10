package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/ethereum/go-ethereum/swarm/streamingviz"
)

func handler(w http.ResponseWriter, r *http.Request) {
	abs, _ := filepath.Abs("./swarm/streamingviz/server/static/index.html")
	view, err := template.ParseFiles(abs)

	//data := getData()
	network := getNetwork()
	data := networkToData(network, "teststream")

	if err != nil {
		fmt.Fprintf(w, "error: %v", err)
	} else {
		view.Execute(w, data)
	}
}

func handleJson(w http.ResponseWriter, r *http.Request) {
	abs, _ := filepath.Abs("./swarm/streamingviz/server/static/data.json")
	view, _ := ioutil.ReadFile(abs)
	fmt.Fprintf(w, "%s", view)
}

func main() {
	http.HandleFunc("/data.json", handleJson)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8585", nil)
}

func getNetwork() *streamingviz.Network {
	sID := "teststream"
	network := streamingviz.NewNetwork()

	// Set up peers
	network.ReceivePeersForNode("A", []string{"B", "D"})
	network.ReceivePeersForNode("B", []string{"A", "D"})
	network.ReceivePeersForNode("C", []string{"D"})

	network.StartBroadcasting("A", sID)
	network.StartConsuming("C", sID)
	network.StartRelaying("D", sID)
	network.StartConsuming("B", sID)
	network.DoneWithStream("B", sID)

	return network
}

func networkToData(network *streamingviz.Network, streamID string) interface{} {
	/*type Node struct {
		ID string
		Group int
	}

	type Link struct {
		Source string
		Target string
		Value int
	}*/

	res := make(map[string]interface{})
	nodes := make([]map[string]interface{}, 0)

	for _, v := range network.Nodes {
		nodes = append(nodes, map[string]interface{}{
			"id":    v.ID,
			"group": v.Group[streamID],
		})
	}

	links := make([]map[string]interface{}, 0)

	for _, v := range network.Links {
		links = append(links, map[string]interface{}{
			"source": v.Source.ID,
			"target": v.Target.ID,
			"value":  2, //v.Value[streamID],
		})
	}

	res["nodes"] = nodes
	res["links"] = links

	b, _ := json.Marshal(res)
	fmt.Println(fmt.Sprintf("The output network is: %s", b))

	var genResult interface{}

	json.Unmarshal(b, &genResult)
	return genResult
}

func getData() map[string]interface{} {
	return map[string]interface{}{
		"nodes": []map[string]interface{}{
			{
				"id":    "A",
				"group": 1,
			},
			{
				"id":    "B",
				"group": 2,
			},
		},
		"links": []map[string]interface{}{
			{
				"source": "A",
				"target": "B",
				"value":  1,
			},
		},
	}
}
