package config

import (
	"os"
	"path/filepath"
)

func ClaudeJSONPath(profilePath string) string {
	if IsDefaultClaudeDir(profilePath) {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ClaudeConfigFile)
		}
	}
	return filepath.Join(profilePath, ClaudeConfigFile)
}
