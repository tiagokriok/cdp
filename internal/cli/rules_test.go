package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tiagokriok/hat/internal/config"
)

func setupRules(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := config.Init(); err != nil {
		t.Fatal(err)
	}
	cfg, _ := config.Load()
	pm := config.NewProfileManager(cfg)
	for _, n := range []string{"work", "personal"} {
		if err := pm.CreateProfile(n, ""); err != nil {
			t.Fatal(err)
		}
	}
	return home
}

func TestHandleRuleAdd(t *testing.T) {
	home := setupRules(t)
	proj := filepath.Join(home, "proj")
	_ = os.MkdirAll(proj, 0755)

	if err := HandleRuleAdd(proj, "ghost"); err == nil {
		t.Error("unknown profile should fail")
	}
	if err := HandleRuleAdd(proj, "work"); err != nil {
		t.Fatal(err)
	}
	if err := HandleRuleAdd(proj, "personal"); err != nil {
		t.Fatal(err)
	}
	cfg, _ := config.Load()
	if len(cfg.Rules) != 1 || cfg.Rules[0].Profile != "personal" {
		t.Fatalf("rules = %+v", cfg.Rules)
	}

	chdir(t, proj)
	if err := HandleRuleAdd(".", "work"); err != nil {
		t.Fatal(err)
	}
	cfg, _ = config.Load()
	if len(cfg.Rules) != 1 || cfg.Rules[0].Profile != "work" || !filepath.IsAbs(cfg.Rules[0].Path) {
		t.Fatalf("'.' should replace with absolute path: %+v", cfg.Rules)
	}
}

func TestHandleRuleListAndRemove(t *testing.T) {
	home := setupRules(t)
	a := filepath.Join(home, "a")
	b := filepath.Join(home, "b")
	_ = os.MkdirAll(a, 0755)
	_ = os.MkdirAll(b, 0755)
	_ = HandleRuleAdd(a, "work")
	_ = HandleRuleAdd(b, "personal")
	chdir(t, a)

	out := captureStdout(t, func() {
		if err := HandleRuleList(); err != nil {
			t.Fatal(err)
		}
	})
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("output = %q", out)
	}
	if !strings.Contains(lines[0], "work") || !strings.Contains(lines[1], "personal") {
		t.Errorf("output = %q", out)
	}
	if !strings.Contains(lines[0], "▶") || strings.Contains(lines[1], "▶") {
		t.Errorf("only the matching rule should be marked: %q", out)
	}

	if err := HandleRuleRemove(a); err != nil {
		t.Fatal(err)
	}
	if err := HandleRuleRemove(a); err == nil {
		t.Error("removing twice should fail")
	}
	cfg, _ := config.Load()
	if len(cfg.Rules) != 1 {
		t.Fatalf("rules = %+v", cfg.Rules)
	}
}

func TestRunInteractiveMenu_UsesRule(t *testing.T) {
	home := setupRules(t)
	proj := filepath.Join(home, "proj", "sub")
	_ = os.MkdirAll(proj, 0755)
	_ = HandleRuleAdd(filepath.Join(home, "proj"), "work")
	chdir(t, proj)

	out := captureStdout(t, func() {
		if err := RunInteractiveMenu(true); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "Using profile work (rule: ") {
		t.Errorf("output = %q", out)
	}
	cfg, _ := config.Load()
	if cfg.CurrentProfile != "work" {
		t.Errorf("current = %q", cfg.CurrentProfile)
	}
}
