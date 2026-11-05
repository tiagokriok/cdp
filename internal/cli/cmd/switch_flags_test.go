package cmd

import (
	"reflect"
	"testing"
)

func TestExtractHatFlags(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantArgs  []string
		wantNoRun bool
		wantYes   bool
	}{
		{"none", []string{"work", "--continue"}, []string{"work", "--continue"}, false, false},
		{"no-run", []string{"work", "--no-run"}, []string{"work"}, true, false},
		{"yes long", []string{"work", "--yes", "--continue"}, []string{"work", "--continue"}, false, true},
		{"yes short", []string{"work", "-y"}, []string{"work"}, false, true},
		{"both", []string{"--yes", "work", "--no-run", "-p"}, []string{"work", "-p"}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArgs, noRun, yes := extractHatFlags(tt.args)
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) || noRun != tt.wantNoRun || yes != tt.wantYes {
				t.Errorf("got %v %v %v", gotArgs, noRun, yes)
			}
		})
	}
}
