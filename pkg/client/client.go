package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"github.com/ozeidan/gosearch/internal/request"
)

func SearchRequest(searchQuery string, fuzzy bool, responseChan chan<- string) {
	defer close(responseChan)
	action := 0
	if fuzzy {
		action = request.FuzzySearch
	} else {
		action = request.PrefixSearch
	}
	req := request.Request{
		Action: action,
		Query:  searchQuery,
	}

	c, err := net.Dial("unix", request.SockAddr)
	if err != nil {
		fmt.Println("could not connect to the server. is the server running?")
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
