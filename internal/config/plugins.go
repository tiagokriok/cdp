package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	enabledPluginsKey     = "enabledPlugins"
	extraMarketplacesKey  = "extraKnownMarketplaces"
	installedPluginsFile  = "installed_plugins.json"
	knownMarketplacesFile = "known_marketplaces.json"
	pluginsDirName        = "plugins"
)

type PluginInfo struct {
	ID        string
	Enabled   bool
	Installed bool
	Version   string
}

func readJSONMap(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("invalid JSON in %s: %w", path, err)
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

func writeJSONMap(path string, m map[string]any) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

func subMap(m map[string]any, key string) map[string]any {
	if v, ok := m[key].(map[string]any); ok {
		return v
	}
	return map[string]any{}
}

func validatePluginID(id string) (marketplace string, err error) {
	name, mk, ok := strings.Cut(id, "@")
	if !ok || name == "" || mk == "" {
		return "", fmt.Errorf("invalid plugin '%s', expected plugin@marketplace", id)
	}
	return mk, nil
}

func settingsPath(profilePath string) string {
	return filepath.Join(profilePath, ClaudeSettingsFile)
}

func installedVersions(profilePath string) (map[string]string, error) {
	installed, err := readJSONMap(filepath.Join(profilePath, pluginsDirName, installedPluginsFile))
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for id, v := range subMap(installed, "plugins") {
		version := ""
		switch e := v.(type) {
		case []any:
			if len(e) > 0 {
				version = versionOf(e[0])
			}
		default:
			version = versionOf(e)
		}
		out[id] = version
	}
	return out, nil
}

func versionOf(v any) string {
	if m, ok := v.(map[string]any); ok {
		if s, ok := m["version"].(string); ok {
			return s
		}
	}
	return ""
}

func ListPlugins(profilePath string) ([]PluginInfo, error) {
	settings, err := readJSONMap(settingsPath(profilePath))
	if err != nil {
		return nil, err
	}
	installed, err := installedVersions(profilePath)
	if err != nil {
		return nil, err
	}
	byID := map[string]*PluginInfo{}
	get := func(id string) *PluginInfo {
		if byID[id] == nil {
			byID[id] = &PluginInfo{ID: id}
		}
		return byID[id]
	}
	for id, v := range subMap(settings, enabledPluginsKey) {
		enabled, _ := v.(bool)
		get(id).Enabled = enabled
	}
	for id, version := range installed {
		p := get(id)
		p.Installed = true
		p.Version = version
	}
	out := make([]PluginInfo, 0, len(byID))
	for _, p := range byID {
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func SetPluginEnabled(profilePath, id string, enabled bool) error {
	if _, err := validatePluginID(id); err != nil {
		return err
	}
	settings, err := readJSONMap(settingsPath(profilePath))
	if err != nil {
		return err
	}
	plugins := subMap(settings, enabledPluginsKey)
	plugins[id] = enabled
	settings[enabledPluginsKey] = plugins
	return writeJSONMap(settingsPath(profilePath), settings)
}

func CopyPlugin(fromPath, toPath, id string) error {
	marketplace, err := validatePluginID(id)
	if err != nil {
		return err
	}
	from, err := readJSONMap(settingsPath(fromPath))
	if err != nil {
		return err
	}
	to, err := readJSONMap(settingsPath(toPath))
	if err != nil {
		return err
	}

	entry, ok := subMap(from, extraMarketplacesKey)[marketplace]
	if !ok {
		known, err := readJSONMap(filepath.Join(fromPath, pluginsDirName, knownMarketplacesFile))
		if err != nil {
			return err
		}
		if src, found := subMap(known, marketplace)["source"]; found {
			entry, ok = map[string]any{"source": src}, true
		}
	}
	if !ok {
		return fmt.Errorf("marketplace '%s' source not found in source profile (settings.json extraKnownMarketplaces or plugins/%s)", marketplace, knownMarketplacesFile)
	}

	marketplaces := subMap(to, extraMarketplacesKey)
	if _, exists := marketplaces[marketplace]; !exists {
		marketplaces[marketplace] = entry
	}
	to[extraMarketplacesKey] = marketplaces
	plugins := subMap(to, enabledPluginsKey)
	plugins[id] = true
	to[enabledPluginsKey] = plugins
	return writeJSONMap(settingsPath(toPath), to)
}

func enabledSet(profilePath string) (map[string]bool, error) {
	settings, err := readJSONMap(settingsPath(profilePath))
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for id, v := range subMap(settings, enabledPluginsKey) {
		if enabled, _ := v.(bool); enabled {
			out[id] = true
		}
	}
	return out, nil
}

func DiffPlugins(aPath, bPath string) (onlyA, onlyB []string, err error) {
	a, err := enabledSet(aPath)
	if err != nil {
		return nil, nil, err
	}
	b, err := enabledSet(bPath)
	if err != nil {
		return nil, nil, err
	}
	for id := range a {
		if !b[id] {
			onlyA = append(onlyA, id)
		}
	}
	for id := range b {
		if !a[id] {
			onlyB = append(onlyB, id)
		}
	}
	sort.Strings(onlyA)
	sort.Strings(onlyB)
	return onlyA, onlyB, nil
}
