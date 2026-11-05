package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tiagokriok/hat/internal/config"
)

const secretsA = `{
  "mcpServers": {
    "local": {"command": "npx", "args": ["-y", "pkg", "--token=ARGSECRET"], "env": {"API_KEY": "ENVSECRET", "OTHER": "x"}},
    "remote": {"type": "http", "url": "https://mcp.example.com/PATHSECRET/v1?key=URLSECRET", "headers": {"Authorization": "Bearer HEADERSECRET"}},
    "stream": {"type": "sse", "url": "https://sse.example.com/"},
    "same": {"command": "a"},
    "onlya": {"command": "z"}
  },
  "projects": {
    "/work/app": {"mcpServers": {"proj": {"command": "node", "env": {"P": "PROJSECRET"}}}},
    "/work/empty": {"mcpServers": {}}
  }
}`

const secretsB = `{
  "mcpServers": {
    "local": {"command": "npx", "args": ["-y", "pkg", "--token=OTHERSECRET"], "env": {"API_KEY": "ENVSECRET2"}},
    "remote": {"type": "http", "url": "https://mcp.example.com/PATHSECRET/v1?key=URLSECRET", "headers": {"Authorization": "Bearer HEADERSECRET"}},
    "same": {"command": "a"},
    "onlyb": {"type": "http", "url": "https://b.example.com"}
  }
}`

func setupMCP(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := config.Init(); err != nil {
		t.Fatal(err)
	}
	cfg, _ := config.Load()
	pm := config.NewProfileManager(cfg)
	for _, n := range []string{"a", "b", "empty"} {
		if err := pm.CreateProfile(n, ""); err != nil {
			t.Fatal(err)
		}
	}
	for n, content := range map[string]string{"a": secretsA, "b": secretsB} {
		p := filepath.Join(home, ".claude-profiles", n, ".claude.json")
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	cfg.CurrentProfile = "a"
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	return home
}

func assertNoSecrets(t *testing.T, out string) {
	t.Helper()
	for _, s := range []string{"SECRET"} {
		if strings.Contains(out, s) {
			t.Errorf("output leaks secret material:\n%s", out)
		}
	}
}

func TestHandleMCP_ListsUserScope(t *testing.T) {
	setupMCP(t)
	var err error
	out := captureStdout(t, func() { err = HandleMCP("", false) })
	if err != nil {
		t.Fatal(err)
	}
	assertNoSecrets(t, out)
	for _, want := range []string{
		"local", "stdio", "npx", "env: API_KEY, OTHER",
		"remote", "http", "https://mcp.example.com", "headers: Authorization",
		"stream", "sse", "https://sse.example.com",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "/work/app") {
		t.Errorf("project servers should need --all:\n%s", out)
	}
}

func TestHandleMCP_All(t *testing.T) {
	setupMCP(t)
	var err error
	out := captureStdout(t, func() { err = HandleMCP("a", true) })
	if err != nil {
		t.Fatal(err)
	}
	assertNoSecrets(t, out)
	if !strings.Contains(out, "/work/app") || !strings.Contains(out, "proj") {
		t.Errorf("missing project servers:\n%s", out)
	}
	if strings.Contains(out, "/work/empty") {
		t.Errorf("empty projects should be hidden:\n%s", out)
	}
}

func TestHandleMCP_NoServers(t *testing.T) {
	setupMCP(t)
	var err error
	out := captureStdout(t, func() { err = HandleMCP("empty", false) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "no MCP servers") {
		t.Errorf("output:\n%s", out)
	}
}

func TestHandleMCP_Errors(t *testing.T) {
	home := setupMCP(t)
	if err := HandleMCP("ghost", false); err == nil {
		t.Error("unknown profile should fail")
	}
	cfg, _ := config.Load()
	cfg.CurrentProfile = ""
	_ = cfg.Save()
	if err := HandleMCP("", false); err == nil {
		t.Error("no current profile should fail")
	}
	bad := filepath.Join(home, ".claude-profiles", "empty", ".claude.json")
	_ = os.WriteFile(bad, []byte("{not json"), 0644)
	if err := HandleMCP("empty", false); err == nil {
		t.Error("invalid json should fail")
	}
}

func TestHandleMCPDiff(t *testing.T) {
	setupMCP(t)
	var err error
	out := captureStdout(t, func() { err = HandleMCPDiff("a", "b") })
	if err != nil {
		t.Fatal(err)
	}
	assertNoSecrets(t, out)
	for _, want := range []string{"Only in a", "onlya", "Only in b", "onlyb", "Different", "local (args, env)"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"same", "remote"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("identical server %q listed:\n%s", unwanted, out)
		}
	}
}

func TestHandleMCPDiff_Identical(t *testing.T) {
	setupMCP(t)
	var err error
	out := captureStdout(t, func() { err = HandleMCPDiff("a", "a") })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "identical") {
		t.Errorf("output:\n%s", out)
	}
}

func TestHandleMCPDiff_UnknownProfile(t *testing.T) {
	setupMCP(t)
	if err := HandleMCPDiff("a", "ghost"); err == nil {
		t.Error("unknown profile should fail")
	}
}

func TestRedactURL(t *testing.T) {
	cases := map[string]string{
		"https://u:p@mcp.example.com:8443/tok/sse?k=v#f": "https://mcp.example.com:8443",
		"http://localhost:3000/mcp":                      "http://localhost:3000",
		"ftp://host/tok":                                 "(redacted url)",
		"mailto:tok@example.com":                         "(redacted url)",
		"https:///tok":                                   "(redacted url)",
		"not a url":                                      "(redacted url)",
		"https://%zz/tok":                                "(redacted url)",
	}
	for in, want := range cases {
		got := redactURL(in)
		if got != want {
			t.Errorf("redactURL(%q) = %q, want %q", in, got, want)
		}
		if strings.Contains(got, "tok") && !strings.Contains(want, "tok") {
			t.Errorf("leak in %q", got)
		}
	}
}
