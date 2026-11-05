package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func sharedDir(t *testing.T) string {
	t.Helper()
	d, err := GetSharedDir()
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestSyncSharedLinksPerEntry(t *testing.T) {
	_, pm, cleanup := setupTestEnv(t)
	defer cleanup()
	if err := pm.CreateProfile("work", ""); err != nil {
		t.Fatal(err)
	}
	sd := sharedDir(t)
	writeFile(t, filepath.Join(sd, "skills", "a", "SKILL.md"), "a")
	writeFile(t, filepath.Join(sd, "skills", "b", "SKILL.md"), "b")
	writeFile(t, filepath.Join(sd, "agents", "x.md"), "x")
	p, _ := pm.GetProfile("work")
	writeFile(t, filepath.Join(p.Path, "skills", "b", "SKILL.md"), "own")

	if err := pm.SetSharedKinds("work", []string{"skills"}, true); err != nil {
		t.Fatal(err)
	}
	res, err := pm.SyncShared("work")
	if err != nil {
		t.Fatal(err)
	}

	if !isLinkTo(filepath.Join(p.Path, "skills", "a"), filepath.Join(sd, "skills", "a")) {
		t.Error("skills/a should be linked")
	}
	if b, _ := os.ReadFile(filepath.Join(p.Path, "skills", "b", "SKILL.md")); string(b) != "own" {
		t.Error("profile's own skills/b must win")
	}
	if !reflect.DeepEqual(res.Conflicts, []string{"skills/b"}) {
		t.Errorf("conflicts = %v", res.Conflicts)
	}
	if _, err := os.Lstat(filepath.Join(p.Path, "agents")); err == nil {
		t.Error("agents not opted in, must not be created")
	}

	res, err = pm.SyncShared("work")
	if err != nil || len(res.Linked) != 0 {
		t.Errorf("second sync should be idempotent, got %v %v", res.Linked, err)
	}
}

func TestSyncSharedClaudeMD(t *testing.T) {
	tests := []struct {
		name      string
		existing  bool
		wantLink  bool
		conflicts int
	}{
		{"absent -> linked", false, true, 0},
		{"present -> kept", true, false, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, pm, cleanup := setupTestEnv(t)
			defer cleanup()
			if err := pm.CreateProfile("work", ""); err != nil {
				t.Fatal(err)
			}
			p, _ := pm.GetProfile("work")
			sd := sharedDir(t)
			writeFile(t, filepath.Join(sd, "CLAUDE.md"), "shared")
			if tt.existing {
				writeFile(t, filepath.Join(p.Path, "CLAUDE.md"), "own")
			}
			if err := pm.SetSharedKinds("work", []string{"CLAUDE.md"}, true); err != nil {
				t.Fatal(err)
			}
			res, err := pm.SyncShared("work")
			if err != nil {
				t.Fatal(err)
			}
			if got := isLinkTo(filepath.Join(p.Path, "CLAUDE.md"), filepath.Join(sd, "CLAUDE.md")); got != tt.wantLink {
				t.Errorf("linked = %v", got)
			}
			if len(res.Conflicts) != tt.conflicts {
				t.Errorf("conflicts = %v", res.Conflicts)
			}
		})
	}
}

func TestSyncSharedPrunes(t *testing.T) {
	_, pm, cleanup := setupTestEnv(t)
	defer cleanup()
	if err := pm.CreateProfile("work", ""); err != nil {
		t.Fatal(err)
	}
	p, _ := pm.GetProfile("work")
	sd := sharedDir(t)
	writeFile(t, filepath.Join(sd, "skills", "a", "SKILL.md"), "a")
	writeFile(t, filepath.Join(sd, "skills", "gone", "SKILL.md"), "g")
	writeFile(t, filepath.Join(sd, "agents", "x.md"), "x")
	other := t.TempDir()
	writeFile(t, filepath.Join(other, "foreign", "f"), "f")
	writeFile(t, filepath.Join(p.Path, "skills", "own", "SKILL.md"), "own")
	if err := os.Symlink(filepath.Join(other, "foreign"), filepath.Join(p.Path, "skills", "foreign")); err != nil {
		t.Fatal(err)
	}

	pm.SetSharedKinds("work", []string{"skills", "agents"}, true)
	if _, err := pm.SyncShared("work"); err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(filepath.Join(sd, "skills", "gone"))
	if _, err := pm.SyncShared("work"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(p.Path, "skills", "gone")); err == nil {
		t.Error("dangling shared symlink should be pruned")
	}
	if _, err := os.Lstat(filepath.Join(p.Path, "skills", "foreign")); err != nil {
		t.Error("foreign symlink must be kept")
	}
	if _, err := os.Stat(filepath.Join(p.Path, "skills", "own", "SKILL.md")); err != nil {
		t.Error("own content must be kept")
	}

	if err := pm.SetSharedKinds("work", []string{"agents"}, false); err != nil {
		t.Fatal(err)
	}
	if _, err := pm.SyncShared("work"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(p.Path, "agents", "x.md")); err == nil {
		t.Error("unlink must prune agents symlink")
	}
	if _, err := os.Lstat(filepath.Join(p.Path, "skills", "a")); err != nil {
		t.Error("skills link must remain")
	}
	if _, err := os.Stat(filepath.Join(sd, "agents", "x.md")); err != nil {
		t.Error("shared content must never be deleted")
	}
}

func TestSyncSharedThroughSymlinkedProfile(t *testing.T) {
	cfg, pm, cleanup := setupTestEnv(t)
	defer cleanup()
	home, _ := os.UserHomeDir()
	real := filepath.Join(home, "real-claude")
	writeFile(t, filepath.Join(real, ".metadata.json"), `{"createdAt":"2024-01-01T00:00:00Z","usageCount":0}`)
	writeFile(t, filepath.Join(real, "settings.json"), "{}")
	if err := os.Symlink(real, filepath.Join(cfg.ProfilesDir, "default")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sharedDir(t), "rules", "r.md"), "r")
	if err := pm.SetSharedKinds("default", []string{"rules"}, true); err != nil {
		t.Fatal(err)
	}
	if _, err := pm.SyncShared("default"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(real, "rules", "r.md")); err != nil {
		t.Error("link should be created inside real dir")
	}
}

func TestSetSharedKindsValidation(t *testing.T) {
	_, pm, cleanup := setupTestEnv(t)
	defer cleanup()
	pm.CreateProfile("work", "")
	if err := pm.SetSharedKinds("work", []string{"plugins"}, true); err == nil {
		t.Error("plugins is not shareable")
	}
	pm.SetSharedKinds("work", []string{"skills", "skills", "rules"}, true)
	p, _ := pm.GetProfile("work")
	if !reflect.DeepEqual(p.Metadata.Shared, []string{"skills", "rules"}) {
		t.Errorf("shared = %v", p.Metadata.Shared)
	}
}

func TestAddToShared(t *testing.T) {
	_, pm, cleanup := setupTestEnv(t)
	defer cleanup()
	pm.CreateProfile("work", "")
	p, _ := pm.GetProfile("work")
	src := filepath.Join(p.Path, "skills", "mine")
	writeFile(t, filepath.Join(src, "SKILL.md"), "mine")
	sd := sharedDir(t)

	name, err := pm.AddToShared("skills", src)
	if err != nil {
		t.Fatal(err)
	}
	if name != "work" {
		t.Errorf("profile = %s", name)
	}
	if !isLinkTo(src, filepath.Join(sd, "skills", "mine")) {
		t.Error("original should be symlink to shared")
	}
	if b, _ := os.ReadFile(filepath.Join(sd, "skills", "mine", "SKILL.md")); string(b) != "mine" {
		t.Error("content should be moved")
	}
	p, _ = pm.GetProfile("work")
	if !reflect.DeepEqual(p.Metadata.Shared, []string{"skills"}) {
		t.Errorf("profile should opt in, got %v", p.Metadata.Shared)
	}
	if _, err := pm.SyncShared("work"); err != nil {
		t.Fatal(err)
	}
	if !isLinkTo(src, filepath.Join(sd, "skills", "mine")) {
		t.Error("sync must not prune the added link")
	}
}

func TestAddToSharedClaudeMD(t *testing.T) {
	_, pm, cleanup := setupTestEnv(t)
	defer cleanup()
	pm.CreateProfile("work", "")
	p, _ := pm.GetProfile("work")
	src := filepath.Join(p.Path, "CLAUDE.md")
	writeFile(t, src, "rules")
	if _, err := pm.AddToShared("CLAUDE.md", src); err != nil {
		t.Fatal(err)
	}
	if !isLinkTo(src, filepath.Join(sharedDir(t), "CLAUDE.md")) {
		t.Error("CLAUDE.md should be symlinked")
	}
}

func TestAddToSharedErrors(t *testing.T) {
	_, pm, cleanup := setupTestEnv(t)
	defer cleanup()
	pm.CreateProfile("work", "")
	p, _ := pm.GetProfile("work")
	sd := sharedDir(t)
	writeFile(t, filepath.Join(sd, "skills", "dup", "SKILL.md"), "s")
	writeFile(t, filepath.Join(p.Path, "skills", "dup", "SKILL.md"), "p")
	outside := t.TempDir()
	writeFile(t, filepath.Join(outside, "skills", "o", "SKILL.md"), "o")

	tests := []struct{ name, kind, path string }{
		{"destination exists", "skills", filepath.Join(p.Path, "skills", "dup")},
		{"not in a profile", "skills", filepath.Join(outside, "skills", "o")},
		{"missing path", "skills", filepath.Join(p.Path, "skills", "nope")},
		{"invalid kind", "plugins", filepath.Join(p.Path, "skills", "dup")},
		{"wrong kind dir", "agents", filepath.Join(p.Path, "skills", "dup")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := pm.AddToShared(tt.kind, tt.path); err == nil {
				t.Error("expected error")
			}
		})
	}
	if b, _ := os.ReadFile(filepath.Join(p.Path, "skills", "dup", "SKILL.md")); string(b) != "p" {
		t.Error("source must be untouched on error")
	}
}

func TestListShared(t *testing.T) {
	_, _, cleanup := setupTestEnv(t)
	defer cleanup()
	sd := sharedDir(t)
	writeFile(t, filepath.Join(sd, "skills", "a", "SKILL.md"), "a")
	writeFile(t, filepath.Join(sd, "CLAUDE.md"), "c")
	got, err := ListShared()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{"skills": {"a"}, "CLAUDE.md": {"CLAUDE.md"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v", got)
	}
}
