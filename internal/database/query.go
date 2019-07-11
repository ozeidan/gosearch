package database

import (
	"errors"
	"log"
	"sort"
	"time"

	"github.com/ozeidan/gosearch/internal/request"
	trie "gopkg.in/ozeidan/fuzzy-patricia.v3/patricia"
)

type resulter interface {
	Result(index int) string
	sort.Interface
}

type sortResult struct {
	result  string
	skipped int
}

type bySkipped []sortResult

func (s bySkipped) Len() int      { return len(s) }
func (s bySkipped) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s bySkipped) Less(i, j int) bool {
	if s[i].skipped == s[j].skipped {
		return len(s[i].result) < len(s[j].result)
	}
	return s[i].skipped < s[j].skipped
}
func (s bySkipped) Result(index int) string {
	return s[index].result
}

type byLength []string

func (l byLength) Len() int           { return len(l) }
func (l byLength) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l byLength) Less(i, j int) bool { return len(l[i]) < len(l[j]) }
func (l byLength) Result(index int) string {
	return l[index]
}

func queryIndex(req request.Request) {
	defer close(req.ResponseChannel)
	log.Println("querying", req.Query)
	prefix := trie.Prefix(req.Query)

	var results resulter
	var err error

	start := logStart("query")
	switch req.Settings.Action {
	case request.PrefixSearch:
		tempResults := byLength{}
		err = indexTrie.VisitSubtree(prefix, func(prefix trie.Prefix, item trie.Item) error {
			list := item.([]indexedFile)
			for _, file := range list {
				tempResults = append(tempResults,
					file.pathNode.GetPath())
			}
			return checkDone(req)
		})

		results = byLength(tempResults)
	case request.SubStringSearch:
		tempResults := byLength{}
		err = indexTrie.VisitSubstring(prefix, req.Settings.CaseInsensitive,
			func(prefix trie.Prefix, item trie.Item) error {
				list := item.([]indexedFile)
				for _, file := range list {
					tempResults = append(tempResults,
						file.pathNode.GetPath())
				}
				return checkDone(req)
			})

		results = byLength(tempResults)
	case request.FuzzySearch:
		tempResults := []sortResult{}
		err = indexTrie.VisitFuzzy(prefix, req.Settings.CaseInsensitive,
			func(prefix trie.Prefix, item trie.Item, skipped int) error {
				list := item.([]indexedFile)
				for _, file := range list {
					tempResults = append(tempResults,
						sortResult{file.pathNode.GetPath(), skipped})
				}
				return checkDone(req)
			})

		results = bySkipped(tempResults)
	}
	logStop(start)

	if err != nil {
		log.Println(err)
		return
	}

	if !req.Settings.NoSort {
		start = logStart("sort")
		if req.Settings.ReverseSort {
			sort.Sort(results)
		} else {
			sort.Sort(sort.Reverse(results))
		}
		logStop(start)
	}

	sendResults(results, req)
}

func sendResults(results resulter, req request.Request) {
	maxResults := req.Settings.MaxResults
	if maxResults == 0 || maxResults > results.Len() {
		maxResults = results.Len()
	}

	var startIndex int

	if req.Settings.ReverseSort {
		startIndex = 0
	} else {
		startIndex = results.Len() - maxResults
	}

	for i := startIndex; i < startIndex+maxResults; i++ {
		select {
		case req.ResponseChannel <- results.Result(i):
		case <-req.Done:
			return
		}
	}
}

func checkDone(req request.Request) error {
	select {
	case <-req.Done:
		return errors.New("request aborted by client")
	default:
		return nil
	}
}

func logStart(action string) time.Time {
	log.Println("starting to", action)
	return time.Now()
}

func logStop(start time.Time) {
	log.Println("done in", time.Now().Sub(start))
}
