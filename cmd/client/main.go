package main

import (
	"flag"
	"fmt"

	"github.com/ozeidan/gosearch/pkg/client"
)

func main() {
	fuzzyFlag := flag.Bool("f", false, "use fuzzy searching")

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		return
	}

	query := flag.Arg(0)

	responseChan := make(chan string, 0)
	go client.SearchRequest(query, *fuzzyFlag, responseChan)

	for response := range responseChan {
		fmt.Printf(response)
	}
}
