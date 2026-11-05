package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Rule struct {
	Path    string `yaml:"path"`
	Profile string `yaml:"profile"`
}

func ResolveRulePath(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	if real, err := filepath.EvalSymlinks(p); err == nil {
		p = real
	}
	return p
}

func isUnder(dir, root string) bool {
	return dir == root || strings.HasPrefix(dir, strings.TrimSuffix(root, string(filepath.Separator))+string(filepath.Separator))
}

func (c *Config) MatchRule(cwd string) *Rule {
	dir := ResolveRulePath(cwd)
	var best *Rule
	bestLen := -1
	for i := range c.Rules {
		root := ResolveRulePath(c.Rules[i].Path)
		if isUnder(dir, root) && len(root) > bestLen {
			best, bestLen = &c.Rules[i], len(root)
		}
	}
	return best
}

func (c *Config) AddRule(path, profile string) {
	key := ResolveRulePath(path)
	for i := range c.Rules {
		if ResolveRulePath(c.Rules[i].Path) == key {
			c.Rules[i] = Rule{Path: path, Profile: profile}
			return
		}
	}
	c.Rules = append(c.Rules, Rule{Path: path, Profile: profile})
}

func (c *Config) RemoveRule(path string) bool {
	key := ResolveRulePath(path)
	for i := range c.Rules {
		if ResolveRulePath(c.Rules[i].Path) == key {
			c.Rules = slices.Delete(c.Rules, i, i+1)
			return true
		}
	}
	return false
}
