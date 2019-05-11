package main

import (
	"flag"
	"fmt"

	"github.com/ozeidan/gosearch/pkg/client"
)

func main() {
	fuzzyFlag := flag.Bool("f", false, "use fuzzy searching")
	noSortFlag := flag.Bool("nosort", false,
		"don't sort the result set for performance gains")

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		return
	}

	query := flag.Arg(0)

	responseChan := make(chan string, 0)
	go client.SearchRequest(query, *fuzzyFlag, *noSortFlag, responseChan)

	for response := range responseChan {
		fmt.Print(response)
	}
}
