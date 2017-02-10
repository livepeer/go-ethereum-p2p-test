package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/ethereum/go-ethereum/swarm/streamingviz"
)

func handler(w http.ResponseWriter, r *http.Request) {
	abs, _ := filepath.Abs("./swarm/streamingviz/server/static/index.html")
	view, err := template.ParseFiles(abs) //ioutil.ReadFile(abs)
	if err != nil {
		fmt.Fprintf(w, "error: %v", err)
	} else {
		view.Execute(w, view)
		//fmt.Fprintf(w, "%s", view)
		//fmt.Fprintf(w, "Hi there. This %s is great!", r.URL.Path[1:])
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

func getNetwork() string {
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

	return network.String()
}
