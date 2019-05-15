package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"github.com/ozeidan/gosearch/internal/request"
)

func SearchRequest(searchQuery string, action int, noSort bool, responseChan chan<- string) {
	defer close(responseChan)

	req := request.Request{
		Query: searchQuery,
		Settings: request.Settings{
			Action: action,
			NoSort: noSort,
		},
	}

	c, err := net.Dial("unix", request.SockAddr)

	if err != nil {
		fmt.Println("could not connect to the server. is the server running?")
		fmt.Printf("err = %+v\n", err)
		return
	}
	defer c.Close()

	err = json.NewEncoder(c).Encode(&req)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(c)
	for {
		bytes, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}
		responseChan <- string(bytes)
	}
}
