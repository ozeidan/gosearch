package database

import (
	"log"
	"sort"
	"time"

	trie "github.com/ozeidan/go-patricia/patricia"
	"github.com/ozeidan/gosearch/internal/request"
)

type sortResult struct {
	result  string
	skipped int
}

type bySkipped []sortResult

func (a bySkipped) Len() int      { return len(a) }
func (a bySkipped) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a bySkipped) Less(i, j int) bool {
	if a[i].skipped == a[j].skipped {
		return len(a[i].result) < len(a[j].result)
	}
	return a[i].skipped < a[j].skipped
}

type byLength []string

func (a byLength) Len() int           { return len(a) }
func (a byLength) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byLength) Less(i, j int) bool { return len(a[i]) < len(a[j]) }

func logStart(action string) time.Time {
	log.Println("starting to", action)
	return time.Now()
}

func logStop(start time.Time) {
	log.Println("done in", time.Now().Sub(start))
}

func queryIndex(req request.Request) {
	defer close(req.ResponseChannel)
	log.Println("querying", req.Query)
	prefix := trie.Prefix(req.Query)

	switch req.Settings.Action {
	case request.PrefixSearch:
		fallthrough
	case request.SubStringSearch:
		var visitFunc func(trie.Prefix, trie.VisitorFunc) error
		if req.Settings.Action == request.PrefixSearch {
			visitFunc = indexTrie.VisitSubtree
		} else {
			visitFunc = indexTrie.VisitSubstring
		}

		if req.Settings.NoSort {
			start := logStart("query and send")
			visitFunc(prefix, sendResults(req.ResponseChannel))
			logStop(start)
			return
		}

		var results []string
		start := logStart("query")
		visitFunc(prefix, func(prefix trie.Prefix, item trie.Item) error {
			list := item.([]indexedFile)
			for _, file := range list {
				results = append(results, file.pathNode.GetPath())
			}
			return nil
		})
		logStop(start)

		// normal sorting is from worst to best
		// so that the best result will show right
		// above the command prompt
		start = logStart("sort")
		if req.Settings.ReverseSort {
			sort.Sort(byLength(results))
		} else {
			sort.Sort(sort.Reverse(byLength(results)))
		}
		logStop(start)

		for _, result := range results {
			req.ResponseChannel <- result
		}
	case request.FuzzySearch:
		if req.Settings.NoSort {
			start := logStart("query and send")
			indexTrie.VisitFuzzy(
				prefix,
				func(prefix trie.Prefix, item trie.Item, skipped int) error {
					return sendResults(req.ResponseChannel)(prefix, item)
				})
			logStop(start)
			return
		}

		var results []sortResult
		visitor := func(prefix trie.Prefix, item trie.Item, skipped int) error {
			list := item.([]indexedFile)
			for _, file := range list {
				results = append(results, sortResult{file.pathNode.GetPath(), skipped})
			}
			return nil
		}
		start := logStart("query")
		indexTrie.VisitFuzzy(prefix, visitor)
		logStop(start)

		start = logStart("sort")
		if req.Settings.ReverseSort {
			sort.Sort(bySkipped(results))
		} else {
			sort.Sort(sort.Reverse(bySkipped(results)))
		}
		logStop(start)

		for _, result := range results {
			req.ResponseChannel <- result.result
		}
	}
}

func sendResults(channel chan string) trie.VisitorFunc {
	return func(prefix trie.Prefix, item trie.Item) error {
		list := item.([]indexedFile)
		for _, file := range list {
			channel <- file.pathNode.GetPath()
		}
		return nil
	}
}
