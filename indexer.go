package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/derekparker/trie"
	"github.com/karrick/godirwalk"
	"github.com/ozeidan/gosearch/request"
	"github.com/ozeidan/gosearch/tree"
)

func start(changeSender <-chan fileChange,
	requestSender <-chan request.Request) {
	initialIndex()

	for {
		select {
		case change := <-changeSender:
			fmt.Printf("received change: %+v\n", change)
			refreshDirectory(change.folderPath)
		case req := <-requestSender:
			queryIndex(req)
		}
	}
}

var filterError error = errors.New("directory filtered")
var filterStrings []string = []string{"/proc", "/home/omar/.cache", "/var/cache"}

var indexTrie *trie.Trie = trie.New()
var fileTree *tree.TreeNode = tree.New()

type indexedFile struct {
	path  string
	isDir bool
}

func initialIndex() {
	dirname := "/"
	addToIndexRecursively(dirname)
}

func refreshDirectory(path string) {
	// fmt.Println("refreshing path", path)
	newDirents, err := godirwalk.ReadDirents(path, nil)
	if err != nil {
		fmt.Println(err)
		panic("paniiiiiiiiic")
	}

	newNames := make([]string, 0, len(newDirents))
	nameDirents := make(map[string]godirwalk.Dirent, len(newNames))
	for _, dirent := range newDirents {
		newNames = append(newNames, dirent.Name())
		nameDirents[dirent.Name()] = *dirent
	}

	oldNames, err := fileTree.GetChildren(path)
	if err != nil {
		fmt.Println(err)
		panic("paniiiiiiiiic")
	}

	createdNames, deletedNames := sliceDifference(newNames, oldNames)
	// fmt.Printf("diff\n +: %+v[%d]\n -: %+v[%d]\n",
	// 	createdNames, len(createdNames), deletedNames, len(deletedNames))

	for _, name := range createdNames {
		dirent := nameDirents[name]
		pathName := filepath.Join(path, name)
		if isFiltered(pathName) {
			continue
		}
		// fmt.Println("adding", pathName, "to index")
		addToIndex(path, name, dirent)
	}

	for _, name := range deletedNames {
		pathName := filepath.Join(path, name)
		// fmt.Println("removing", pathName, "from index")
		deleteFromIndex(path, name)
		fileTree.DeleteAt(pathName)
	}
}

func sliceDifference(sliceA, sliceB []string) ([]string, []string) {
	mapA := sliceToSet(sliceA)
	mapB := sliceToSet(sliceB)

	for name, _ := range mapA {
		if _, ok := mapB[name]; ok {
			delete(mapA, name)
			delete(mapB, name)
		}
	}

	return setToSlice(mapA), setToSlice(mapB)
}

func sliceToSet(slice []string) map[string]bool {
	createMap := make(map[string]bool, len(slice))
	for _, name := range slice {
		createMap[name] = true
	}
	return createMap
}

func setToSlice(set map[string]bool) []string {
	createSlice := make([]string, 0, len(set))

	for key, _ := range set {
		createSlice = append(createSlice, key)
	}

	return createSlice
}

func addToIndex(path, name string, dirent godirwalk.Dirent) {
	pathName := filepath.Join(path, name)

	if dirent.IsDir() {
		addToIndexRecursively(pathName)
	} else {
		fileTree.Add(pathName)
		indexTrieAdd(name, indexedFile{pathName, false})
	}
}

func deleteFromIndex(path, name string) {
	pathName := filepath.Join(path, name)

	indexTrieDelete(name, path)

	children, err := fileTree.GetChildren(pathName)
	if err != nil {
		// fmt.Println("warning:", err)
		return
	}

	for _, child := range children {
		deleteFromIndex(pathName, child)
	}
}

func addToIndexRecursively(path string) {
	godirwalk.Walk(path, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if isFiltered(osPathname) {
				return filterError
			}

			name := de.Name()
			newFile := indexedFile{osPathname, de.IsDir()}

			indexTrieAdd(name, newFile)
			fileTree.Add(osPathname)

			return nil
		},
		Unsorted: true,
		ErrorCallback: func(_ string, err error) godirwalk.ErrorAction {
			if err == filterError {
				return godirwalk.SkipNode
			}
			// fmt.Println(err)
			return godirwalk.SkipNode
		},
	})
}

func isFiltered(path string) bool {
	for _, filterString := range filterStrings {
		if strings.HasPrefix(path, filterString) {
			return true
		}
	}

	return false
}

func indexTrieAdd(name string, index indexedFile) {
	if node, ok := indexTrie.Find(name); ok {
		fileList := node.Meta().([]indexedFile)
		fileList = append(fileList, index)
	} else {
		indexTrie.Add(name, []indexedFile{index})
	}
}

func indexTrieDelete(name, path string) {
	if node, ok := indexTrie.Find(name); ok {
		fileList := node.Meta().([]indexedFile)
		for i := 0; i < len(fileList); i++ {
			index := fileList[i]
			if index.path != path {
				continue
			}

			fileList[i] = fileList[len(fileList)-1]
			fileList = fileList[:len(fileList)-1]
			break
		}
	}
}

func queryIndex(req request.Request) {
	defer close(req.ResponseChannel)

	switch req.Action {
	case request.PrefixSearch:
		results := indexTrie.PrefixSearch(req.Query)
		sendResults(results, req.ResponseChannel)
	case request.FuzzySearch:
		results := indexTrie.FuzzySearch(req.Query)
		sendResults(results, req.ResponseChannel)
	}
}

func sendResults(results []string, channel chan string) {
	for _, result := range results {
		node, ok := indexTrie.Find(result)
		if !ok {
			continue
		}
		list := node.Meta().([]indexedFile)
		for _, file := range list {
			channel <- file.path
		}
	}
}
