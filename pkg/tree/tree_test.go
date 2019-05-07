package tree

import (
	"reflect"
	"testing"
)

// TODO: looks weird
var baseTreeNode TreeNode = TreeNode{
	map[string]TreeNode{
		"home": TreeNode{
			map[string]TreeNode{
				"omar": TreeNode{
					map[string]TreeNode{
						"empty": TreeNode{},
						"documents": TreeNode{
							map[string]TreeNode{
								"documentA": TreeNode{},
								"documentB": TreeNode{},
							},
						},
					},
				},
			},
		},
	},
}

func TestTreeNode_GetChildren(t *testing.T) {
	type fields struct {
		children map[string]TreeNode
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			"simple_test",
			fields{baseTreeNode.children},
			args{"/home"},
			[]string{"omar"},
			false,
		},
		{
			"empty_test",
			fields{baseTreeNode.children},
			args{"/home/omar/empty"},
			[]string{},
			false,
		},
		{
			"multiple_test",
			fields{baseTreeNode.children},
			args{"/home/omar/documents"},
			[]string{"documentA", "documentB"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := TreeNode{
				children: tt.fields.children,
			}
			got, err := tree.GetChildren(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("TreeNode.GetChildren() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TreeNode.GetChildren() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestTreeNode_Add(t *testing.T) {
// 	type fields struct {
// 		children map[string]TreeNode
// 	}
// 	type args struct {
// 		path string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			t := &TreeNode{
// 				children: tt.fields.children,
// 			}
// 			if err := t.Add(tt.args.path); (err != nil) != tt.wantErr {
// 				t.Errorf("TreeNode.Add() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func TestTreeNode_DeleteAt(t *testing.T) {
	type fields struct {
		children map[string]TreeNode
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"working_test",
			fields{baseTreeNode.children},
			args{"/home/omar/documents"},
			false,
		},
		{
			"not_working_test",
			fields{baseTreeNode.children},
			args{"/home/omar/doesnotexist"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := &TreeNode{
				children: tt.fields.children,
			}
			if err := tree.DeleteAt(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("TreeNode.DeleteAt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want *TreeNode
	}{
		{
			"default_test",
			&TreeNode{map[string]TreeNode{}},
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
