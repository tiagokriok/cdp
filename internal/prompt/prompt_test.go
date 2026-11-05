package prompt

import (
	"strings"
	"testing"
)

func setup(t *testing.T, input string) {
	t.Helper()
	restore := SetInput(strings.NewReader(input))
	t.Cleanup(restore)
}

func TestSharedReaderAnswersPromptsInOrder(t *testing.T) {
	setup(t, "y\nn\nyes\n")
	got := []bool{Ask("a", false), Ask("b", true), Ask("c", false)}
	want := []bool{true, false, true}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("answer %d = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestEmptyLineAndEOFUseDefault(t *testing.T) {
	setup(t, "\n")
	if !Ask("a", true) {
		t.Error("empty line should return default yes")
	}
	if !Ask("b", true) {
		t.Error("EOF should return default yes")
	}
	if Ask("c", false) {
		t.Error("EOF should return default no")
	}
}
