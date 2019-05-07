package tree

import (
	"reflect"
	"testing"
)

// TODO: looks weird
var baseNode = Node{
	map[string]Node{
		"home": Node{
			map[string]Node{
				"omar": Node{
					map[string]Node{
						"empty": Node{},
						"documents": Node{
							map[string]Node{
								"documentA": Node{},
								"documentB": Node{},
							},
						},
					},
				},
			},
		},
	},
}

func TestNode_GetChildren(t *testing.T) {
	type fields struct {
		children map[string]Node
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
			fields{baseNode.children},
			args{"/home"},
			[]string{"omar"},
			false,
		},
		{
			"empty_test",
			fields{baseNode.children},
			args{"/home/omar/empty"},
			[]string{},
			false,
		},
		{
			"multiple_test",
			fields{baseNode.children},
			args{"/home/omar/documents"},
			[]string{"documentA", "documentB"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := Node{
				children: tt.fields.children,
			}
			got, err := tree.GetChildren(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Node.GetChildren() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Node.GetChildren() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestNode_Add(t *testing.T) {
// 	type fields struct {
// 		children map[string]Node
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
// 			t := &Node{
// 				children: tt.fields.children,
// 			}
// 			if err := t.Add(tt.args.path); (err != nil) != tt.wantErr {
// 				t.Errorf("Node.Add() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func TestNode_DeleteAt(t *testing.T) {
	type fields struct {
		children map[string]Node
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
			fields{baseNode.children},
			args{"/home/omar/documents"},
			false,
		},
		{
			"not_working_test",
			fields{baseNode.children},
			args{"/home/omar/doesnotexist"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := &Node{
				children: tt.fields.children,
			}
			if err := tree.DeleteAt(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Node.DeleteAt() error = %v, wantErr %v", err, tt.wantErr)
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
			&Node{map[string]Node{}},
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
