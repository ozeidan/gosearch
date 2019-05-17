package main

import (
	"flag"
	"fmt"

	"github.com/ozeidan/gosearch/pkg/client"
)

func main() {
	fuzzyFlag := flag.Bool("f", false, "use fuzzy searching")
	prefixFlag := flag.Bool("p", false, "do a prefix search (faster)")
	noSortFlag := flag.Bool("nosort", false,
		"don't sort the result set for performance gains when fuzzy searching")
	reverseSortFlag := flag.Bool("r", false, "reverse the sort order")

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

	options := []client.Option{}

	if *fuzzyFlag {
		options = append(options, client.Fuzzy)
	}
	if *prefixFlag {
		options = append(options, client.PrefixSearch)
	}
	if *noSortFlag {
		options = append(options, client.NoSort)
	}
	if *reverseSortFlag {
		options = append(options, client.ReverseSort)
	}

	responseChan, err := client.SearchRequest(query, options...)

	if err == client.ErrConnectionFailed {
		fmt.Println(err)
		fmt.Println("is the server running?")
		return
	}

	for response := range responseChan {
		fmt.Print(response)
	}
}