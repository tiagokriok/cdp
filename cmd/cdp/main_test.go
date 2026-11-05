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
		{"no args", []string{"cdp"}, []string{"cdp"}},
		{"profile name", []string{"cdp", "work", "--continue"}, []string{"cdp", "switch", "work", "--continue"}},
		{"known command", []string{"cdp", "list"}, []string{"cdp", "list"}},
		{"flag", []string{"cdp", "--help"}, []string{"cdp", "--help"}},
		{"complete", []string{"cdp", "__complete", "sw", ""}, []string{"cdp", "__complete", "sw", ""}},
		{"complete no desc", []string{"cdp", "__completeNoDesc", "sw"}, []string{"cdp", "__completeNoDesc", "sw"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rewriteImplicitSwitch(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
