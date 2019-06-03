package tree

import (
	"reflect"
	"testing"
)

var files = []string{
	"/home/user/Documents/file1",
	"/home",
	"/home/user/Downloads/file2",
	"/home/user/empty",
	"/home/user/Desktop",
	"/home/user/Desktop/file3",
	"/home/user/Desktop/file4",
}

func buildTree() *Node {
	tree := New()
	for _, file := range files {
		tree.Add(file)
	}
	return tree
}

func TestNode_GetChildren(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			"simple_test",
			args{"/home"},
			[]string{"user"},
			false,
		},
		{
			"empty_test",
			args{"/home/user/empty"},
			[]string{},
			false,
		},
		{
			"multiple_test",
			args{"/home/user/Desktop"},
			[]string{"file3", "file4"},
			false,
		},
		{
			"invalid_path",
			args{"/home/user/doesntexist"},
			[]string{},
			true,
		},
	}
	tree := buildTree()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tree.GetChildren(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Node.GetChildren() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Node.GetChildren() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNode_AddGetPath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "new_addition",
			args: args{
				path: "/usr",
			},
		},
		{
			name: "new_file",
			args: args{
				path: "/home/user/Downloads/newfile",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := buildTree()
			newNode := tree.Add(tt.args.path)

			gotPath := newNode.GetPath()

			if gotPath != tt.args.path {
				t.Errorf("Node.GetPath() error, wanted %s, got %s", tt.args.path, gotPath)
			}
		})
	}
}

func TestNode_DeleteAt(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			"working",
			"/home/user/Documents",
			false,
		},
		{
			"file_not_found",
			"/home/user/doesnotexist",
			true,
		},
		{
			"invalid_path",
			"/home/user/invalid/path/err",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := buildTree()
			if err := tree.DeleteAt(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("Node.DeleteAt() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := tree.DeleteAt(tt.path); err == nil {
				t.Errorf("Node.DeleteAt() error = nil, not deleted properly")
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want *Node
	}{
		{
			"default_test",
			&Node{[]*Node{}, "", nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
