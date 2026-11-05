package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeJSONPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	def := filepath.Join(home, ".claude")
	custom := filepath.Join(home, "profiles", "work")
	for _, d := range []string{def, custom} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}
	link := filepath.Join(home, "profiles", "default")
	if err := os.Symlink(def, link); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name, path, want string
	}{
		{"custom profile", custom, filepath.Join(custom, ".claude.json")},
		{"default dir", def, filepath.Join(home, ".claude.json")},
		{"symlink to default", link, filepath.Join(home, ".claude.json")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClaudeJSONPath(tt.path); got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
