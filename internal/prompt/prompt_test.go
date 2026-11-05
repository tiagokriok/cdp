package prompt

import (
	"strings"
	"testing"
)

func setup(t *testing.T, input string) {
	t.Helper()
	restore := SetInput(strings.NewReader(input))
	t.Cleanup(func() {
		restore()
		SetAssumeYes(false)
	})
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

func TestEmptyLineUsesDefaultAndEOFSaysNo(t *testing.T) {
	setup(t, "\n")
	if !Ask("a", true) {
		t.Error("empty line should return default yes")
	}
	if Ask("b", true) {
		t.Error("EOF should return no even when default is yes")
	}
	if Ask("c", false) {
		t.Error("EOF should return no")
	}
}

func TestLastLineWithoutNewlineIsRead(t *testing.T) {
	setup(t, "y")
	if !Ask("a", false) {
		t.Error("answer without trailing newline should be read")
	}
}

func TestConfirmAssumeYesSkipsInput(t *testing.T) {
	setup(t, "n\n")
	SetAssumeYes(true)
	if !Confirm("a", false) {
		t.Error("Confirm should auto-yes")
	}
	SetAssumeYes(false)
	if Confirm("b", true) {
		t.Error("input n should have been left unread and answer no")
	}
}

func TestAskIgnoresAssumeYes(t *testing.T) {
	setup(t, "y\n")
	SetAssumeYes(true)
	if Ask("remove?", false) {
		t.Error("Ask must keep default under --yes")
	}
}
