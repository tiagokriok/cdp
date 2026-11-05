package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestListPlugins(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "settings.json"), `{"enabledPlugins":{"a@m":true,"b@m":false}}`)
	writeFile(t, filepath.Join(dir, "plugins", "installed_plugins.json"),
		`{"version":2,"plugins":{"a@m":[{"scope":"user","version":"1.2.0","extra":1}],"c@m":{"version":"0.1"},"d@m":[]}}`)

	got, err := ListPlugins(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []PluginInfo{
		{ID: "a@m", Enabled: true, Installed: true, Version: "1.2.0"},
		{ID: "b@m"},
		{ID: "c@m", Installed: true, Version: "0.1"},
		{ID: "d@m", Installed: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v", got)
	}
}

func TestListPluginsEmptyProfile(t *testing.T) {
	got, err := ListPlugins(t.TempDir())
	if err != nil || len(got) != 0 {
		t.Errorf("got %v %v", got, err)
	}
}

func TestSetPluginEnabledPreservesKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	writeFile(t, path, `{"model":"opus","permissions":{"allow":["x"]},"enabledPlugins":{"a@m":true}}`)

	if err := SetPluginEnabled(dir, "b@m", true); err != nil {
		t.Fatal(err)
	}
	if err := SetPluginEnabled(dir, "a@m", false); err != nil {
		t.Fatal(err)
	}
	m := readJSON(t, path)
	if m["model"] != "opus" || m["permissions"] == nil {
		t.Errorf("other keys lost: %v", m)
	}
	want := map[string]any{"a@m": false, "b@m": true}
	if !reflect.DeepEqual(m["enabledPlugins"], want) {
		t.Errorf("enabledPlugins = %v", m["enabledPlugins"])
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "\n  \"model\"") {
		t.Errorf("expected 2-space indent:\n%s", data)
	}
}

func TestSetPluginEnabledCreatesSettings(t *testing.T) {
	dir := t.TempDir()
	if err := SetPluginEnabled(dir, "a@m", true); err != nil {
		t.Fatal(err)
	}
	m := readJSON(t, filepath.Join(dir, "settings.json"))
	if !reflect.DeepEqual(m["enabledPlugins"], map[string]any{"a@m": true}) {
		t.Errorf("got %v", m)
	}
}

func TestPluginsInvalidInputNeverOverwrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	writeFile(t, path, `{broken`)

	if err := SetPluginEnabled(dir, "a@m", true); err == nil || !strings.Contains(err.Error(), "settings.json") {
		t.Errorf("expected clear error, got %v", err)
	}
	if b, _ := os.ReadFile(path); string(b) != `{broken` {
		t.Error("invalid file must not be overwritten")
	}
	if _, err := ListPlugins(dir); err == nil {
		t.Error("list should error on invalid JSON")
	}
	if err := SetPluginEnabled(t.TempDir(), "noat", true); err == nil {
		t.Error("id without @marketplace must be rejected")
	}
}

func TestCopyPlugin(t *testing.T) {
	src := map[string]string{
		"settings": `{"extraKnownMarketplaces":{"m":{"source":{"source":"github","repo":"o/r"}}},"enabledPlugins":{"a@m":true}}`,
	}
	known := `{"m":{"source":{"source":"git","url":"https://x/y.git"},"installLocation":"/abs/path","lastUpdated":"t"}}`

	tests := []struct {
		name       string
		fromSet    string
		fromKnown  string
		toSet      string
		wantSource map[string]any
		wantErr    string
	}{
		{"from settings", src["settings"], "", `{"model":"opus"}`, map[string]any{"source": "github", "repo": "o/r"}, ""},
		{"from known_marketplaces", `{}`, known, ``, map[string]any{"source": "git", "url": "https://x/y.git"}, ""},
		{"target keeps its marketplace", src["settings"], "", `{"extraKnownMarketplaces":{"m":{"source":{"source":"github","repo":"mine/own"}}}}`, map[string]any{"source": "github", "repo": "mine/own"}, ""},
		{"source not found", `{}`, `{}`, ``, nil, "marketplace 'm'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, to := t.TempDir(), t.TempDir()
			writeFile(t, filepath.Join(from, "settings.json"), tt.fromSet)
			if tt.fromKnown != "" {
				writeFile(t, filepath.Join(from, "plugins", "known_marketplaces.json"), tt.fromKnown)
			}
			if tt.toSet != "" {
				writeFile(t, filepath.Join(to, "settings.json"), tt.toSet)
			}
			err := CopyPlugin(from, to, "a@m")
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			m := readJSON(t, filepath.Join(to, "settings.json"))
			mk := m["extraKnownMarketplaces"].(map[string]any)["m"].(map[string]any)
			if !reflect.DeepEqual(mk["source"], tt.wantSource) {
				t.Errorf("source = %v", mk["source"])
			}
			if m["enabledPlugins"].(map[string]any)["a@m"] != true {
				t.Error("plugin should be enabled")
			}
			if tt.toSet == `{"model":"opus"}` && m["model"] != "opus" {
				t.Error("other keys lost")
			}
		})
	}
}

func TestDiffPlugins(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(a, "settings.json"), `{"enabledPlugins":{"x@m":true,"both@m":true,"off@m":false}}`)
	writeFile(t, filepath.Join(b, "settings.json"), `{"enabledPlugins":{"y@m":true,"both@m":true,"off@m":true}}`)
	onlyA, onlyB, err := DiffPlugins(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(onlyA, []string{"x@m"}) || !reflect.DeepEqual(onlyB, []string{"off@m", "y@m"}) {
		t.Errorf("onlyA=%v onlyB=%v", onlyA, onlyB)
	}
}
