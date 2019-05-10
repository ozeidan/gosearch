package database

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/karrick/godirwalk"
	trie "github.com/ozeidan/go-patricia/patricia"
	"github.com/ozeidan/gosearch/internal/config"
	"github.com/ozeidan/gosearch/internal/fanotify"
	"github.com/ozeidan/gosearch/internal/request"
	"github.com/ozeidan/gosearch/pkg/tree"
)

// Start starts the indexing and listens for file changes and requests
// changeSender is used to get file change messages from the caller
// requestSender is used to get request messages from the caller
func Start(changeSender <-chan fanotify.FileChange,
	requestSender <-chan request.Request) {
	initialIndex()

	for {
		select {
		case change := <-changeSender:
			fmt.Printf("received change: %+v\n", change)
			refreshDirectory(change.FolderPath)
		case req := <-requestSender:
			queryIndex(req)
		}
	}
}

var errFilter = errors.New("directory filtered")

var indexTrie *trie.Trie = trie.NewTrie()
var fileTree *tree.Node = tree.New()

type indexedFile struct {
	pathNode *tree.Node
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
		if config.IsPathFiltered(pathName) {
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

	for name := range mapA {
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

	for key := range set {
		createSlice = append(createSlice, key)
	}

	return createSlice
}

func addToIndex(path, name string, dirent godirwalk.Dirent) {
	pathName := filepath.Join(path, name)

	if dirent.IsDir() {
		addToIndexRecursively(pathName)
	} else {
		newNode := fileTree.Add(pathName)
		indexTrieAdd(name, indexedFile{newNode})
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
			if config.IsPathFiltered(osPathname) {
				return errFilter
			}

			newNode := fileTree.Add(osPathname)
			name := de.Name()
			newFile := indexedFile{newNode}
			indexTrieAdd(name, newFile)

			return nil
		},
		Unsorted: true,
		ErrorCallback: func(_ string, err error) godirwalk.ErrorAction {
			if err == errFilter {
				return godirwalk.SkipNode
			}
			// fmt.Println(err)
			return godirwalk.SkipNode
		},
	})
}

func indexTrieAdd(name string, index indexedFile) {
	prefix := trie.Prefix(name)
	if item := indexTrie.Get(prefix); item != nil {
		fileList := item.([]indexedFile)
		fileList = append(fileList, index)
	} else {
		indexTrie.Insert(prefix, []indexedFile{index})
	}
}

func indexTrieDelete(name, path string) {
	prefix := trie.Prefix(name)
	if item := indexTrie.Get(prefix); item != nil {
		fileList := item.([]indexedFile)
		for i := 0; i < len(fileList); i++ {
			index := fileList[i]
			if index.pathNode.GetPath() != path {
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
	prefix := trie.Prefix(req.Query)

	switch req.Action {
	case request.PrefixSearch:
		indexTrie.VisitSubtree(
			prefix,
			sendResults(req.ResponseChannel))
	case request.FuzzySearch:
		indexTrie.VisitFuzzy(
			prefix,
			sendResults(req.ResponseChannel))
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
