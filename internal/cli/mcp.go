package cli

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/tiagokriok/hat/internal/config"
	"github.com/tiagokriok/hat/internal/ui"
)

type mcpServer map[string]any

type mcpConfig struct {
	Servers  map[string]mcpServer `json:"mcpServers"`
	Projects map[string]struct {
		Servers map[string]mcpServer `json:"mcpServers"`
	} `json:"projects"`
}

func loadMCPConfig(cfg *config.Config, profileName string) (*mcpConfig, error) {
	profile, err := config.NewProfileManager(cfg).GetProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("profile '%s' does not exist", profileName)
	}
	data, err := os.ReadFile(config.ClaudeJSONPath(profile.Path))
	if os.IsNotExist(err) {
		return &mcpConfig{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read .claude.json: %w", err)
	}
	var mc mcpConfig
	if err := json.Unmarshal(data, &mc); err != nil {
		return nil, fmt.Errorf("failed to parse .claude.json of profile '%s': %w", profileName, err)
	}
	return &mc, nil
}

func (s mcpServer) kind() string {
	if t, ok := s["type"].(string); ok && t != "" {
		return t
	}
	if _, ok := s["command"]; ok {
		return "stdio"
	}
	if _, ok := s["url"]; ok {
		return "http"
	}
	return "unknown"
}

func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Opaque != "" || u.Host == "" {
		return "(redacted url)"
	}
	return u.Scheme + "://" + u.Host
}

func (s mcpServer) target() string {
	if c, ok := s["command"].(string); ok {
		return c
	}
	if u, ok := s["url"].(string); ok {
		return redactURL(u)
	}
	return ""
}

func (s mcpServer) keysOf(field string) string {
	m, ok := s[field].(map[string]any)
	if !ok || len(m) == 0 {
		return ""
	}
	return field + ": " + strings.Join(slices.Sorted(maps.Keys(m)), ", ")
}

func printServers(servers map[string]mcpServer) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, name := range slices.Sorted(maps.Keys(servers)) {
		s := servers[name]
		var extras []string
		for _, f := range []string{"env", "headers"} {
			if k := s.keysOf(f); k != "" {
				extras = append(extras, k)
			}
		}
		fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\n", name, s.kind(), s.target(), strings.Join(extras, "; "))
	}
	_ = tw.Flush()
}

func ResolveProfileName(cfg *config.Config, name string) (string, error) {
	if name != "" {
		return name, nil
	}
	if cfg.CurrentProfile == "" {
		return "", fmt.Errorf("no current profile, pass a profile name")
	}
	return cfg.CurrentProfile, nil
}

func HandleMCP(profileName string, all bool) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	profileName, err = ResolveProfileName(cfg, profileName)
	if err != nil {
		return err
	}
	mc, err := loadMCPConfig(cfg, profileName)
	if err != nil {
		return err
	}

	projectNames := make([]string, 0, len(mc.Projects))
	if all {
		for p, v := range mc.Projects {
			if len(v.Servers) > 0 {
				projectNames = append(projectNames, p)
			}
		}
		sort.Strings(projectNames)
	}

	if len(mc.Servers) == 0 && len(projectNames) == 0 {
		ui.Info(fmt.Sprintf("no MCP servers in profile '%s'", profileName))
		return nil
	}

	if len(mc.Servers) > 0 {
		ui.Header(fmt.Sprintf("User-scope MCP servers (%s)", profileName))
		printServers(mc.Servers)
	}
	for _, p := range projectNames {
		fmt.Println()
		ui.Header(fmt.Sprintf("Project %s", p))
		printServers(mc.Projects[p].Servers)
	}
	return nil
}

func differingFields(a, b mcpServer) []string {
	keys := map[string]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	var diff []string
	for k := range keys {
		if !reflect.DeepEqual(a[k], b[k]) {
			diff = append(diff, k)
		}
	}
	sort.Strings(diff)
	return diff
}

func HandleMCPDiff(a, b string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	ma, err := loadMCPConfig(cfg, a)
	if err != nil {
		return err
	}
	mb, err := loadMCPConfig(cfg, b)
	if err != nil {
		return err
	}

	var onlyA, onlyB, different []string
	for _, n := range slices.Sorted(maps.Keys(ma.Servers)) {
		sb, ok := mb.Servers[n]
		if !ok {
			onlyA = append(onlyA, n)
			continue
		}
		if fields := differingFields(ma.Servers[n], sb); len(fields) > 0 {
			different = append(different, fmt.Sprintf("%s (%s)", n, strings.Join(fields, ", ")))
		}
	}
	for _, n := range slices.Sorted(maps.Keys(mb.Servers)) {
		if _, ok := ma.Servers[n]; !ok {
			onlyB = append(onlyB, n)
		}
	}

	if len(onlyA)+len(onlyB)+len(different) == 0 {
		ui.Success(fmt.Sprintf("MCP servers of '%s' and '%s' are identical", a, b))
		return nil
	}
	for _, section := range []struct {
		title string
		items []string
	}{
		{"Only in " + a, onlyA},
		{"Only in " + b, onlyB},
		{"Different", different},
	} {
		if len(section.items) == 0 {
			continue
		}
		ui.Header(section.title)
		for _, item := range section.items {
			fmt.Printf("  %s\n", item)
		}
	}
	return nil
}
