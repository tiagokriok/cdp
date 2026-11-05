package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func exists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}

func TestIsExcludedEntry(t *testing.T) {
	tests := map[string]bool{
		"projects": true, "history.jsonl": true, ".metadata.json": true,
		".claude.json.backup.123": true, "stats-cache.json": true,
		"skills": false, "agents": false, "plugins": false, ".claude.json": false, "CLAUDE.md": false,
	}
	for name, want := range tests {
		if got := IsExcludedEntry(name); got != want {
			t.Errorf("IsExcludedEntry(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestCopyProfileTree(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	writeTree(t, src, map[string]string{
		".claude.json":                "{}",
		"settings.json":               "{}",
		"skills/foo/SKILL.md":         "skill",
		"agents/a.md":                 "agent",
		"plugins/cache/x/plugin.json": "nested cache kept",
		"projects/p/session.jsonl":    "history",
		"cache/c":                     "cache",
		"history.jsonl":               "h",
		".claude.json.backup.1":       "b",
		".metadata.json":              "{}",
		"todos/t.json":                "t",
	})
	if err := os.WriteFile(filepath.Join(src, ".credentials.json"), []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/etc/passwd", filepath.Join(src, "skills", "link")); err != nil {
		t.Fatal(err)
	}

	copied, symlinks, err := copyProfileTree(src, dst)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{".claude.json", "settings.json", "skills/foo/SKILL.md", "agents/a.md", "plugins/cache/x/plugin.json", ".credentials.json"} {
		if !exists(filepath.Join(dst, want)) {
			t.Errorf("expected %s copied", want)
		}
	}
	for _, not := range []string{"projects", "cache", "history.jsonl", ".claude.json.backup.1", ".metadata.json", "todos", "skills/link"} {
		if exists(filepath.Join(dst, not)) {
			t.Errorf("expected %s not copied", not)
		}
	}
	if copied != 6 {
		t.Errorf("copied = %d, want 6", copied)
	}
	if len(symlinks) != 1 || symlinks[0] != filepath.Join("skills", "link") {
		t.Errorf("symlinks = %v", symlinks)
	}
	info, _ := os.Stat(filepath.Join(dst, ".credentials.json"))
	if info.Mode().Perm() != 0600 {
		t.Errorf("credentials mode = %v, want 0600", info.Mode().Perm())
	}
}

func TestRewritePluginPaths(t *testing.T) {
	src, dst := "/old/.claude", t.TempDir()
	writeTree(t, dst, map[string]string{
		"plugins/installed_plugins.json":  `{"installPath":"/old/.claude/plugins/cache/a","projectPath":"/work/proj"}`,
		"plugins/known_marketplaces.json": `{"installLocation":"/old/.claude/plugins/marketplaces/m"}`,
		"plugins/other.json":              `{"p":"/old/.claude/plugins/x"}`,
	})
	if err := rewritePluginPaths([]string{src}, dst); err != nil {
		t.Fatal(err)
	}
	read := func(n string) string {
		b, _ := os.ReadFile(filepath.Join(dst, "plugins", n))
		return string(b)
	}
	if got := read("installed_plugins.json"); !strings.Contains(got, dst+"/plugins/cache/a") || !strings.Contains(got, "/work/proj") || strings.Contains(got, "/old/") {
		t.Errorf("installed_plugins.json = %s", got)
	}
	if got := read("known_marketplaces.json"); !strings.Contains(got, dst+"/plugins/marketplaces/m") {
		t.Errorf("known_marketplaces.json = %s", got)
	}
	if got := read("other.json"); !strings.Contains(got, "/old/.claude") {
		t.Errorf("other.json should be untouched: %s", got)
	}
}

func TestRewritePluginPaths_NoPluginsDir(t *testing.T) {
	if err := rewritePluginPaths([]string{"/a"}, t.TempDir()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func withStdin(t *testing.T, input string, fn func()) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	w.WriteString(input)
	w.Close()
	orig := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = orig }()
	fn()
}

func TestImportProfile_CopiesSubdirsAndRewritesPlugins(t *testing.T) {
	cfg, pm, cleanup := setupTestEnv(t)
	defer cleanup()

	src := t.TempDir()
	writeTree(t, src, map[string]string{
		".claude.json":                   "{}",
		"settings.json":                  "{}",
		"skills/s/SKILL.md":              "skill",
		"projects/p/x.jsonl":             "history",
		"plugins/installed_plugins.json": `{"installPath":"` + src + `/plugins/cache/a"}`,
	})

	var err error
	withStdin(t, "y\nn\n", func() { err = pm.ImportProfile(src, "imp", "d") })
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(cfg.ProfilesDir, "imp")
	if !exists(filepath.Join(dest, "skills", "s", "SKILL.md")) {
		t.Error("skills not copied")
	}
	if exists(filepath.Join(dest, "projects")) {
		t.Error("projects should be excluded")
	}
	b, _ := os.ReadFile(filepath.Join(dest, "plugins", "installed_plugins.json"))
	if !strings.Contains(string(b), dest+"/plugins/cache/a") || strings.Contains(string(b), src) {
		t.Errorf("plugin paths not rewritten: %s", b)
	}
}

func TestCloneProfile_CopiesSubdirs(t *testing.T) {
	cfg, pm, cleanup := setupTestEnv(t)
	defer cleanup()

	if err := pm.CreateProfile("a", ""); err != nil {
		t.Fatal(err)
	}
	a := filepath.Join(cfg.ProfilesDir, "a")
	writeTree(t, a, map[string]string{"agents/x.md": "agent", "projects/p/h": "h"})

	if err := pm.CloneProfile("a", "b"); err != nil {
		t.Fatal(err)
	}
	b := filepath.Join(cfg.ProfilesDir, "b")
	if !exists(filepath.Join(b, "agents", "x.md")) {
		t.Error("agents not cloned")
	}
	if exists(filepath.Join(b, "projects")) {
		t.Error("projects should be excluded")
	}
}
