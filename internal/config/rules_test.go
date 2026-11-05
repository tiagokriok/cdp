package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchRule(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	work := filepath.Join(home, "dev", "work")
	deep := filepath.Join(work, "api", "deep")
	other := filepath.Join(home, "dev", "workshop")
	for _, d := range []string{deep, other, filepath.Join(work, "x")} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}
	link := filepath.Join(home, "link")
	if err := os.Symlink(work, link); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{Rules: []Rule{
		{Path: "~/dev/work", Profile: "work"},
		{Path: "~/dev/work/api", Profile: "api"},
		{Path: "~/missing", Profile: "ghost"},
	}}

	tests := []struct {
		name string
		cwd  string
		want string
	}{
		{"equal", work, "work"},
		{"under", filepath.Join(work, "x"), "work"},
		{"longest wins", deep, "api"},
		{"prefix sibling not matched", other, ""},
		{"symlinked cwd", filepath.Join(link, "api"), "api"},
		{"outside", home, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := cfg.MatchRule(tt.cwd)
			got := ""
			if r != nil {
				got = r.Profile
			}
			if got != tt.want {
				t.Errorf("MatchRule(%s) = %q, want %q", tt.cwd, got, tt.want)
			}
		})
	}
}

func TestAddRemoveRule(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, "proj")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := &Config{}

	cfg.AddRule("~/proj", "a")
	cfg.AddRule(dir, "b")
	if len(cfg.Rules) != 1 || cfg.Rules[0].Profile != "b" {
		t.Fatalf("same path should replace: %+v", cfg.Rules)
	}
	cfg.AddRule("~/other", "c")
	if len(cfg.Rules) != 2 {
		t.Fatalf("want 2 rules: %+v", cfg.Rules)
	}
	if !cfg.RemoveRule("~/proj") || len(cfg.Rules) != 1 {
		t.Fatalf("remove failed: %+v", cfg.Rules)
	}
	if cfg.RemoveRule("~/nope") {
		t.Error("removing unknown rule should return false")
	}
}

func TestRulesYAMLRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := Init(); err != nil {
		t.Fatal(err)
	}
	cfg, _ := Load()
	if len(cfg.Rules) != 0 {
		t.Fatal("legacy config has no rules")
	}
	cfg.AddRule("~/dev/work", "work")
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	got, _ := Load()
	if len(got.Rules) != 1 || got.Rules[0].Path != "~/dev/work" {
		t.Fatalf("round trip: %+v", got.Rules)
	}
}
