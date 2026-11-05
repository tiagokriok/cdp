package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tiagokriok/hat/internal/config"
	"github.com/tiagokriok/hat/internal/executor"
)

const fakeClaude = `#!/bin/sh
if [ "$1" = "--version" ]; then echo "9.9.9 (Claude Code)"; exit 0; fi
case "$HAT_PROFILE" in
  out) echo '{"loggedIn":false}'; exit 1;;
  junk) echo "something odd"; exit 1;;
  slow) exec sleep 5;;
esac
echo "{\"loggedIn\":true,\"authMethod\":\"claude.ai\",\"email\":\"$HAT_PROFILE@example.com\",\"orgName\":\"Org $HAT_PROFILE\",\"configDirectory\":\"$CLAUDE_CONFIG_DIR\"}"
`

func setupDoctor(t *testing.T, profiles ...string) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, v := range []string{"ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN", "CLAUDE_CODE_OAUTH_TOKEN"} {
		t.Setenv(v, "")
	}
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "claude"), []byte(fakeClaude), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+":"+os.Getenv("PATH"))
	if err := config.Init(); err != nil {
		t.Fatal(err)
	}
	cfg, _ := config.Load()
	pm := config.NewProfileManager(cfg)
	for _, n := range profiles {
		if err := pm.CreateProfile(n, ""); err != nil {
			t.Fatal(err)
		}
	}
	return home
}

func TestHandleDoctor_Healthy(t *testing.T) {
	setupDoctor(t, "work", "out", "junk")
	var err error
	out := captureStdout(t, func() { err = HandleDoctor() })
	if err != nil {
		t.Fatalf("err = %v\n%s", err, out)
	}
	for _, want := range []string{
		"9.9.9 (Claude Code)",
		"logged in: work@example.com (Org work) via claude.ai",
		"not logged in",
		"auth status unavailable: something odd",
		"no auth env overrides",
		"all checks passed",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestHandleDoctor_Timeout(t *testing.T) {
	setupDoctor(t, "slow")
	old := doctorTimeout
	doctorTimeout = 200 * time.Millisecond
	t.Cleanup(func() { doctorTimeout = old })

	var err error
	out := captureStdout(t, func() { err = HandleDoctor() })
	if err != nil {
		t.Fatalf("timeout should only warn: %v", err)
	}
	if !strings.Contains(out, "timed out") {
		t.Errorf("output:\n%s", out)
	}
}

func TestHandleDoctor_Problems(t *testing.T) {
	home := setupDoctor(t, "work")
	t.Setenv("ANTHROPIC_API_KEY", "sk-test")

	work := filepath.Join(home, ".claude-profiles", "work")
	if err := os.Symlink(filepath.Join(home, "nowhere"), filepath.Join(work, "dangling")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(work, "skills"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(home, "nowhere"), filepath.Join(work, "skills", "gone")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(work, "settings.json")); err != nil {
		t.Fatal(err)
	}

	cfg, _ := config.Load()
	cfg.AddRule(filepath.Join(home, "missing-dir"), "work")
	cfg.AddRule(home, "ghost")
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	var err error
	out := captureStdout(t, func() { err = HandleDoctor() })
	if !errors.Is(err, ErrDoctorFailed) {
		t.Fatalf("err = %v, want ErrDoctorFailed", err)
	}
	for _, want := range []string{
		"ANTHROPIC_API_KEY set in environment",
		"broken symlink: " + filepath.Join(work, "dangling"),
		"broken symlink: " + filepath.Join(work, "skills", "gone"),
		"required file 'settings.json' is missing",
		"path does not exist",
		"ghost: profile does not exist",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestHandleDoctor_DefaultProfileStaleFile(t *testing.T) {
	home := setupDoctor(t)
	def := filepath.Join(home, ".claude")
	if err := os.MkdirAll(def, 0755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{filepath.Join(def, ".claude.json"), filepath.Join(home, ".claude.json")} {
		if err := os.WriteFile(f, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	cfg, _ := config.Load()
	pm := config.NewProfileManager(cfg)
	if err := pm.CreateProfile("seed", ""); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(home, ".claude-profiles", "default")
	if err := os.Symlink(def, link); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{".metadata.json": `{"createdAt":"2024-01-01T00:00:00Z","usageCount":0}`, "settings.json": "{}"} {
		if err := os.WriteFile(filepath.Join(def, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	var err error
	out := captureStdout(t, func() { err = HandleDoctor() })
	if err != nil {
		t.Fatalf("err = %v\n%s", err, out)
	}
	if !strings.Contains(out, "is ignored (stale)") {
		t.Errorf("missing stale warning:\n%s", out)
	}
}

func TestHandleDoctor_NoClaude(t *testing.T) {
	setupDoctor(t, "work")
	t.Setenv("PATH", t.TempDir())
	oldLocations := executor.SystemLocations
	executor.SystemLocations = nil
	t.Cleanup(func() { executor.SystemLocations = oldLocations })

	var err error
	out := captureStdout(t, func() { err = HandleDoctor() })
	if !errors.Is(err, ErrDoctorFailed) {
		t.Fatalf("missing claude should fail, err=%v\n%s", err, out)
	}
	if !strings.Contains(out, "claude binary not found") {
		t.Errorf("output:\n%s", out)
	}
}

func TestHandleDoctor_NotInitialized(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var err error
	captureStdout(t, func() { err = HandleDoctor() })
	if !errors.Is(err, ErrDoctorFailed) {
		t.Errorf("err = %v", err)
	}
}
