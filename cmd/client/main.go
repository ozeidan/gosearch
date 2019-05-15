package main

import (
	"flag"
	"fmt"

	"github.com/ozeidan/gosearch/internal/request"
	"github.com/ozeidan/gosearch/pkg/client"
)

func main() {
	fuzzyFlag := flag.Bool("f", false, "use fuzzy searching")
	prefixFlag := flag.Bool("p", false, "do a prefix search (faster)")
	noSortFlag := flag.Bool("nosort", false,
		"don't sort the result set for performance gains when fuzzy searching")

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		return
	}

	if *fuzzyFlag && *prefixFlag {
		flag.Usage()
		return
	}

	query := flag.Arg(0)

	responseChan := make(chan string, 0)

	action := request.SubStringSearch
	if *fuzzyFlag {
		action = request.FuzzySearch
	} else if *prefixFlag {
		action = request.PrefixSearch
	} else {
	}

	go client.SearchRequest(query, action, *noSortFlag, responseChan)

	for response := range responseChan {
		fmt.Print(response)
	}
}
