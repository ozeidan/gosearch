package client

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"

	"github.com/ozeidan/gosearch/internal/request"
)

type Option func(req *request.Request)

var ErrConnectionFailed = errors.New("could not connect to the server")

func Fuzzy(req *request.Request) {
	req.Settings.Action = request.FuzzySearch
}

func PrefixSearch(req *request.Request) {
	req.Settings.Action = request.PrefixSearch
}

func PathSearch(req *request.Request) {
	req.Settings.Action = request.PathSearch
}

func NoSort(req *request.Request) {
	req.Settings.NoSort = true
}

func ReverseSort(req *request.Request) {
	req.Settings.ReverseSort = true
}

func MaxResults(max int) Option {
	return func(req *request.Request) {
		req.Settings.MaxResults = max
	}
}

func SearchRequest(searchQuery string, options ...Option) (<-chan string, error) {
	responseChan := make(chan string, 0)

	req := new(request.Request)
	req.Query = searchQuery

	for _, option := range options {
		option(req)
	}

	c, err := net.Dial("unix", request.SockAddr)

	if err != nil {
		return nil, ErrConnectionFailed
	}

	err = json.NewEncoder(c).Encode(&req)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(responseChan)
		defer c.Close()
		reader := bufio.NewReader(c)
		for {
			bytes, err := reader.ReadBytes('\n')
			if err != nil {
				// TODO: handle this error
				return
			}
			responseChan <- string(bytes)
		}
	}()

	return responseChan, nil
}
