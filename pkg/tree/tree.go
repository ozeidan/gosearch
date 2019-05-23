package tree

import (
	"fmt"
	"strings"

	"github.com/ozeidan/go-patricia/patricia"
)

// Node is a node of a directory tree
type Node struct {
	children []*Node
	name     string
	parent   *Node
	mask     uint64
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
	current := t
	// parts := pathToParts(path)

	var start int
	var end int
	var idx int
	var part string

	for end != len(path) {
		start = end
		idx = strings.Index(path[end+1:], "/")
		if idx == -1 {
			end = len(path)
		} else {
			end = end + idx + 1
		}

		current.mask |= makePrefixMask(path[start:])
		part = path[start+1 : end]

		if child, ok := current.findFile(part); ok {
			current = child
		} else {
			newPart := make([]byte, len(part))
			copy(newPart, []byte(part))
			child = &Node{make([]*Node, 0), string(newPart), current, 0}
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

	var mask uint64
	for _, c := range current.children {
		mask |= c.mask
	}
	current.mask = mask

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

func (t Node) walk(path string, visitor func(part string) error) error {
	err := visitor(path)
	if err != nil {
		return err
	}

	for _, c := range t.children {
		newPath := path + "/" + c.name
		c.walk(newPath, visitor)
	}

	return nil
}

const upperBits = 0xFFFFFFC00
const lowerBits = 0x3FFFFFF000000000

func makePrefixMask(key string) uint64 {
	var mask uint64 = 0
	for _, b := range key {
		if b >= '0' && b <= '9' {
			// 0-9 bits: 0-9
			b -= 48
		} else if b >= 'A' && b <= 'Z' {
			// A-Z bits: 10-35
			b -= 55
		} else if b >= 'a' && b <= 'z' {
			// a-z bits: 36-61
			b -= 61
		} else if b == '.' {
			b = 62
		} else if b == '-' {
			b = 63
		} else {
			continue
		}
		mask |= uint64(1) << uint64(b)
	}

	return mask
}

func caseInsensitiveMask(mask uint64) uint64 {
	mask |= (mask & upperBits) << uint64(26)
	mask |= (mask & lowerBits) >> uint64(26)
	return mask
}

type potentialSubtree struct {
	idx     int
	skipped int
	part    string
	node    *Node
}

func (t Node) VisitFuzzy(bquery patricia.Prefix,
	caseInsensitive bool,
	visitor patricia.FuzzyVisitorFunc) error {
	var (
		m          uint64
		cmp        uint64
		i          int
		matchCount int
		skipped    int
		p          potentialSubtree
	)

	query := string(bquery)

	potential := []potentialSubtree{potentialSubtree{node: &t, part: "", idx: 0}}
	for l := len(potential); l > 0; l = len(potential) {
		i = l - 1
		p = potential[i]

		potential = potential[:i]

		matchCount, skipped = fuzzyMatchCount(p.node.name, query[p.idx:], p.idx, caseInsensitive)
		p.idx += matchCount

		if p.idx != 0 {
			p.skipped += skipped
		}

		if p.idx == len(query) {
			fullName := p.part + "/" + p.node.name

			err := p.node.walk(fullName, func(path string) error {
				err := visitor(patricia.Prefix(path), struct{}{}, p.skipped)
				if err != nil {
					return err
				}
				return nil
			})

			if err != nil {
				return err
			}

			continue
		}

		m = makePrefixMask(query[p.idx:])

		if caseInsensitive {
			cmp = caseInsensitiveMask(p.node.mask)
		} else {
			cmp = p.node.mask
		}

		if (cmp & m) != m {
			continue
		}

		for _, c := range p.node.children {
			var newPart string
			if p.part == "/" {
				newPart = "/" + p.node.name
			} else {
				newPart = p.part + "/" + p.node.name
			}
			potential = append(potential, potentialSubtree{
				node:    c,
				part:    newPart,
				idx:     p.idx,
				skipped: p.skipped,
			})
		}
	}

	return nil
}

func fuzzyMatchCount(part, partialQuery string, idx int, caseInsensitive bool) (count, skipped int) {
	if caseInsensitive {
		part = strings.ToLower(part)
		partialQuery = strings.ToLower(partialQuery)
	}
	for i := 0; i < len(part); i++ {
		if part[i] != partialQuery[count] {
			if count+idx > 0 {
				skipped++
			}
			continue
		}

		count++
		if count >= len(partialQuery) {
			return
		}
	}
	return
}

// New returns a new Node
func New() *Node {
	return &Node{make([]*Node, 0), "", nil, 0}
}

func pathToParts(path string) []string {
	return strings.Split(path, "/")[1:]
}
