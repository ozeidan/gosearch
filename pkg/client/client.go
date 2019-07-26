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
func SubStringSearch(req *request.Request) {
	req.Settings.Action = request.SubStringSearch
}

func NoSort(req *request.Request) {
	req.Settings.NoSort = true
}

func ReverseSort(req *request.Request) {
	req.Settings.ReverseSort = true
}

func CaseInsensitive(req *request.Request) {
	req.Settings.CaseInsensitive = true
}

func MaxResults(max int) Option {
	return func(req *request.Request) {
		req.Settings.MaxResults = max
	}
}

func SearchRequest(searchQuery string, done <-chan struct{}, options ...Option) (<-chan string, error) {
	responseChan := make(chan string, 0)
	lines := make(chan string)

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

	reader := bufio.NewReader(c)

	go func() {
		defer close(lines)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			lines <- line
		}
	}()

	go func() {
		defer close(responseChan)
		defer c.Close()

		for {
			select {
			case line, ok := <-lines:
				if !ok {
					return
				}

				responseChan <- line
			case <-done:
				return
			}
		}
	}()

	return responseChan, nil
}
