package request

import (
	"encoding/json"
	"log"
	"net"
	"os"
)

// SockAddr is the path at which the unix domain socket is created
const SockAddr = "/tmp/gosearch.sock"

const (
	// PrefixSearch denotes searching for prefixes of file/directory names
	PrefixSearch = iota
	// FuzzySearch does a fuzzy search on file/directory names
	FuzzySearch
	// IndexRefresh refreshes the whole database over the files
	IndexRefresh
)

// Request holds the details of a request
// that was received over the unix domain socket
type Request struct {
	// Action describes the requested action
	Action int `json:"action"`
	// Query holds the string which is searched for
	Query string `json:"data"`
	// ResponseChannel is the channel
	// on which the database will send back the results
	ResponseChannel chan string `json:"-"`
}

// ListenAndServe starts listening for and accepting requests
// on a unix domain socket.
// requestReceiver is used for passing on the requests to the caller
func ListenAndServe(requestReceiver chan<- Request) {
	if err := os.RemoveAll(SockAddr); err != nil {
		log.Fatal(err)
	}

	l, err := net.Listen("unix", SockAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Println("accept error:", err)
			continue
		}

		serve(conn, requestReceiver)
	}
}

func serve(c net.Conn, requestReceiver chan<- Request) {
	defer c.Close()
	request := Request{}
	err := json.NewDecoder(c).Decode(&request)

	if err != nil {
		log.Println("failed to decode request:", err)
		return
	}

	request.ResponseChannel = make(chan string)
	requestReceiver <- request

	for response := range request.ResponseChannel {
		responseBytes := []byte(response + "\n")
		n, err := c.Write(responseBytes)
		if n != len(responseBytes) {
			log.Printf("wrote %d byte, expected to write %d", n, len(responseBytes))
		}
		if err != nil {
			break
		}

	}
}
