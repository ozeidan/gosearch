package tree

import (
	"fmt"
	"strings"
)

// Node is a node of a directory tree
type Node struct {
	children map[string]Node
}

// ErrInvalidPath is returned when the path given to one of
// the tree operations does not exist
type ErrInvalidPath struct {
	path string
}

func (err ErrInvalidPath) Error() string {
	return fmt.Sprintf("accesing invalid path: %s", err.path)
}

// GetChildren returns the directoryies/files of a directory
// determiend by path
func (t Node) GetChildren(path string) ([]string, error) {
	parts := pathToParts(path)
	current := t

	for _, part := range parts {
		if child, ok := current.children[part]; ok {
			current = child
		} else {
			return nil, ErrInvalidPath{path}
		}
	}

	keys := make([]string, 0, len(current.children))
	for key := range current.children {
		keys = append(keys, key)
	}

	return keys, nil
}

// Add adds a directory to the directory tree
func (t *Node) Add(path string) {
	parts := pathToParts(path)
	current := *t

	for _, part := range parts {
		if child, ok := current.children[part]; ok {
			current = child
		} else {
			child = Node{make(map[string]Node, 0)}
			current.children[part] = child
			current = child
		}
	}
}

// DeleteAt deletes a directory and its subdirectories/files from the tree
func (t *Node) DeleteAt(path string) error {
	parts := pathToParts(path)
	current := *t
	for _, part := range parts[:len(parts)-1] {
		if child, ok := current.children[part]; ok {
			current = child
		} else {
			return ErrInvalidPath{path}
		}
	}

	if _, ok := current.children[parts[len(parts)-1]]; !ok {
		return ErrInvalidPath{path}
	}

	delete(current.children, parts[len(parts)-1])
	return nil
}

// New returns a new Node
func New() *Node {
	return &Node{make(map[string]Node, 0)}
}

func pathToParts(path string) []string {
	return strings.Split(path, "/")[1:]
}
