package main

import (
	"reflect"
	"testing"
)

func TestRewriteImplicitSwitch(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"no args", []string{"hat"}, []string{"hat"}},
		{"profile name", []string{"hat", "work", "--continue"}, []string{"hat", "switch", "work", "--continue"}},
		{"known command", []string{"hat", "list"}, []string{"hat", "list"}},
		{"rule command", []string{"hat", "rule", "list"}, []string{"hat", "rule", "list"}},
		{"doctor command", []string{"hat", "doctor"}, []string{"hat", "doctor"}},
		{"mcp command", []string{"hat", "mcp", "--all"}, []string{"hat", "mcp", "--all"}},
		{"implicit switch keeps --yes", []string{"hat", "work", "--yes"}, []string{"hat", "switch", "work", "--yes"}},
		{"leading --yes before command", []string{"hat", "--yes", "delete", "x"}, []string{"hat", "--yes", "delete", "x"}},
		{"short -y after command", []string{"hat", "delete", "x", "-y"}, []string{"hat", "delete", "x", "-y"}},
		{"flag", []string{"hat", "--help"}, []string{"hat", "--help"}},
		{"complete", []string{"hat", "__complete", "sw", ""}, []string{"hat", "__complete", "sw", ""}},
		{"complete no desc", []string{"hat", "__completeNoDesc", "sw"}, []string{"hat", "__completeNoDesc", "sw"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rewriteImplicitSwitch(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
