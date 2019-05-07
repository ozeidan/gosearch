package tree

import (
	"fmt"
	"strings"
)

type TreeNode struct {
	children map[string]TreeNode
}

type InvalidPathErr struct {
	path string
}

func (err InvalidPathErr) Error() string {
	return fmt.Sprintf("accesing invalid path: %s", err.path)
}

func (t TreeNode) GetChildren(path string) ([]string, error) {
	parts := pathToParts(path)
	current := t

	for _, part := range parts {
		if child, ok := current.children[part]; ok {
			current = child
		} else {
			return nil, InvalidPathErr{path}
		}
	}

	keys := make([]string, 0, len(current.children))
	for key := range current.children {
		keys = append(keys, key)
	}

	return keys, nil
}

func (t *TreeNode) Add(path string) {
	parts := pathToParts(path)
	current := *t

	for _, part := range parts {
		if child, ok := current.children[part]; ok {
			current = child
		} else {
			child = TreeNode{make(map[string]TreeNode, 0)}
			current.children[part] = child
			current = child
		}
	}
}

func (t *TreeNode) DeleteAt(path string) error {
	parts := pathToParts(path)
	current := *t
	for _, part := range parts[:len(parts)-1] {
		if child, ok := current.children[part]; ok {
			current = child
		} else {
			return InvalidPathErr{path}
		}
	}

	if _, ok := current.children[parts[len(parts)-1]]; !ok {
		return InvalidPathErr{path}
	}

	delete(current.children, parts[len(parts)-1])
	return nil
}

func pathToParts(path string) []string {
	return strings.Split(path, "/")[1:]
}

func New() *TreeNode {
	return &TreeNode{make(map[string]TreeNode, 0)}
}
