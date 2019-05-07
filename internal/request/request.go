package request

import (
	"encoding/json"
	"log"
	"net"
	"os"
)

const SockAddr = "/tmp/gosearch.sock"

const (
	PrefixSearch = iota
	FuzzySearch
	IndexRefresh
)

type Request struct {
	Action          int         `json:"action"`
	Query           string      `json:"data"`
	ResponseChannel chan string `json:"-"`
}

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
		responseBytes := []byte(response)
		n, err := c.Write(responseBytes)
		if n != len(responseBytes) {
			log.Printf("wrote %d byte, expected to write %d", n, len(responseBytes))
		}
		if err != nil {
			break
		}

	}
}
