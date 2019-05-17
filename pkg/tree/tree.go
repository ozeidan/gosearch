package tree

import (
	"fmt"
	"strings"
)

// Node is a node of a directory tree
type Node struct {
	children []*Node
	name     string
	parent   *Node
}

// ErrInvalidPath is returned when the path given to one of
// the tree operations does not exist
type ErrInvalidPath struct {
	path string
}

func (err ErrInvalidPath) Error() string {
	return fmt.Sprintf("accesing invalid path: %s", err.path)
}

func (t Node) findFile(name string) (*Node, bool) {
	for _, c := range t.children {
		if c.name == name {
			return c, true
		}
	}
	return nil, false
}

func (t *Node) deleteFile(name string) bool {
	for i := range t.children {
		if t.children[i].name == name {
			t.children[i] = t.children[len(t.children)-1]
			t.children = t.children[:len(t.children)-1]
			return true
		}
	}
	return false
}

// GetChildren returns the directoryies/files of a directory
// determiend by path
func (t *Node) GetChildren(path string) ([]string, error) {
	parts := pathToParts(path)
	current := t

	for _, part := range parts {
		if child, ok := current.findFile(part); ok {
			current = child
		} else {
			return nil, ErrInvalidPath{path}
		}
	}

	keys := make([]string, 0, len(current.children))
	for _, child := range current.children {
		keys = append(keys, child.name)
	}

	return keys, nil
}

// Add adds a directory to the directory tree
func (t *Node) Add(path string) *Node {
	parts := pathToParts(path)
	current := t

	for _, part := range parts {
		if child, ok := current.findFile(part); ok {
			current = child
		} else {
			newPart := make([]byte, len(part))
			copy(newPart, []byte(part))
			child = &Node{make([]*Node, 0), string(newPart), current}
			current.children = append(current.children, child)
			current = child
		}
	}

	return current
}

// DeleteAt deletes a directory and its subdirectories/files from the tree
func (t *Node) DeleteAt(path string) error {
	parts := pathToParts(path)
	current := *t
	for _, part := range parts[:len(parts)-1] {
		if child, ok := current.findFile(part); ok {
			current = *child
		} else {
			return ErrInvalidPath{path}
		}
	}

	ok := current.deleteFile(parts[len(parts)-1])
	if !ok {
		return ErrInvalidPath{path}
	}
	return nil
}

func (t *Node) GetPath() string {
	parts := make([]string, 0, 10) // faster?
	current := t

	for ; current.parent != nil; current = current.parent {
		parts = append(parts, current.name)
	}

	var builder strings.Builder
	for i := len(parts) - 1; i >= 0; i-- {
		builder.WriteString("/")
		builder.WriteString(parts[i])
	}

	return builder.String()
}

// New returns a new Node
func New() *Node {
	return &Node{make([]*Node, 0), "", nil}
}

func pathToParts(path string) []string {
	return strings.Split(path, "/")[1:]
}
