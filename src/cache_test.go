package main

import (
	"testing"
)

func Test_getItemTypeByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "login type",
			args: args{
				name: "login",
			},
			want: 1,
		},
		{
			name: "node type",
			args: args{
				name: "node",
			},
			want: 2,
		},
		{
			name: "card type",
			args: args{
				name: "card",
			},
			want: 3,
		},
		{
			name: "identity type",
			args: args{
				name: "identity",
			},
			want: 4,
		},
		{
			name: "unknown type",
			args: args{
				name: "none",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getItemTypeByName(tt.args.name)
			if got != tt.want {
				t.Errorf("getItemTypeByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isItemIdFound(t *testing.T) {
	type args struct {
		itemId []string
		item   Item
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test note",
			args: args{
				itemId: []string{"note"},
				item: Item{
					Type: 2,
				},
			},
			want: true,
		},
		{
			name: "test login note",
			args: args{
				itemId: []string{"login", "note"},
				item: Item{
					Type: 4,
				},
			},
			want: false,
		},
		{
			name: "test card",
			args: args{
				itemId: []string{"card"},
				item: Item{
					Type: 3,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isItemIdFound(tt.args.itemId, tt.args.item); got != tt.want {
				t.Errorf("isItemIdFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
