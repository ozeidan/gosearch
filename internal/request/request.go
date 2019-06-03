package request

import (
	"encoding/json"
	"log"
	"net"
	"os"
)

// SockAddr is the path at which the unix domain socket is created
const SockAddr = "/run/gosearch.sock"

const (
	SubStringSearch = iota
	// PrefixSearch denotes searching for prefixes of file/directory names
	PrefixSearch
	// FuzzySearch does a fuzzy search on file/directory names
	FuzzySearch
	// IndexRefresh refreshes the whole database over the files
	IndexRefresh
)

// Request holds the details of a request
// that was received over the unix domain socket
type Request struct {
	// Query holds the string which is searched for
	Query string `json:"data"`
	// Settings holds some query settings
	Settings Settings `json:"settings"`
	// ResponseChannel is the channel
	// on which the database will send back the results
	ResponseChannel chan string `json:"-"`
	// Done is used to signal to the database
	// that no more results are needed
	Done chan struct{} `json:"-"`
}

type Settings struct {
	// Action describes the requested action
	Action int `json:"action"`
	// Maximal amount of results to be transmitted, 0 means unlimited
	MaxResults int `json:"max_results"`
	// Don't sort the results when fuzzy searching
	NoSort      bool `json:"no_sort"`
	ReverseSort bool `json:"reverse_sort"`
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

	setSocketPermissions()
	if err != nil {
		log.Fatal("couldn't set socket permissions properly", err)
	}

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Println("accept error:", err)
			continue
		}

		serve(conn, requestReceiver)
	}
}

func setSocketPermissions() error {
	// group, err := user.LookupGroup("users")
	// if err != nil {
	// 	return err
	// }
	// gid, err := strconv.Atoi(group.Gid)
	// if err != nil {
	// 	return err
	// }
	// err = os.Chown(SockAddr, -1, gid)
	// if err != nil {
	// 	return err
	// }

	err := os.Chmod(SockAddr, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func serve(c net.Conn, requestReceiver chan<- Request) {
	defer c.Close()
	request := Request{}
	err := json.NewDecoder(c).Decode(&request)

	if err != nil {
		// TODO: send error back
		log.Println("failed to decode request:", err)
		return
	}

	request.ResponseChannel = make(chan string)
	request.Done = make(chan struct{})
	requestReceiver <- request

	for response := range request.ResponseChannel {
		responseBytes := []byte(response + "\n")
		n, err := c.Write(responseBytes)
		if n != len(responseBytes) {
			log.Printf("wrote %d bytes, expected to write %d", n, len(responseBytes))
		}
		if err != nil {
			log.Println("failed to write to unix domain socket:", err)
			request.Done <- struct{}{}
			break
		}

	}
}
