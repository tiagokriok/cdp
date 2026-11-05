package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var excludedEntries = map[string]bool{
	MetadataFileName:   true,
	"projects":         true,
	"debug":            true,
	"todos":            true,
	"shell-snapshots":  true,
	"file-history":     true,
	"session-env":      true,
	"sessions":         true,
	"statsig":          true,
	"telemetry":        true,
	"cache":            true,
	"image-cache":      true,
	"paste-cache":      true,
	"plans":            true,
	"tasks":            true,
	"backups":          true,
	"ide":              true,
	"history.jsonl":    true,
	"stats-cache.json": true,
}

var pluginPathFiles = []string{"installed_plugins.json", "known_marketplaces.json"}

func IsExcludedEntry(name string) bool {
	return excludedEntries[name] || strings.HasPrefix(name, ".claude.json.backup")
}

func copyProfileTree(src, dst string) (copied int, symlinks []string, err error) {
	var walk func(srcDir, dstDir, rel string) error
	walk = func(srcDir, dstDir, rel string) error {
		entries, err := os.ReadDir(srcDir)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", srcDir, err)
		}
		for _, entry := range entries {
			name := entry.Name()
			if rel == "" && IsExcludedEntry(name) {
				continue
			}
			srcPath := filepath.Join(srcDir, name)
			dstPath := filepath.Join(dstDir, name)
			info, err := entry.Info()
			if err != nil {
				return fmt.Errorf("failed to stat %s: %w", srcPath, err)
			}
			switch {
			case info.Mode()&os.ModeSymlink != 0:
				symlinks = append(symlinks, filepath.Join(rel, name))
			case info.IsDir():
				if err := os.MkdirAll(dstPath, info.Mode().Perm()); err != nil {
					return fmt.Errorf("failed to create %s: %w", dstPath, err)
				}
				if err := walk(srcPath, dstPath, filepath.Join(rel, name)); err != nil {
					return err
				}
			case info.Mode().IsRegular():
				if err := copyFile(srcPath, dstPath, info.Mode().Perm()); err != nil {
					return err
				}
				copied++
			}
		}
		return nil
	}
	err = walk(src, dst, "")
	return copied, symlinks, err
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", src, err)
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", dst, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return fmt.Errorf("failed to write file %s: %w", dst, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("failed to write file %s: %w", dst, err)
	}
	return os.Chmod(dst, perm)
}

func rewritePluginPaths(srcDirs []string, dst string) error {
	for _, name := range pluginPathFiles {
		path := filepath.Join(dst, "plugins", name)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to read %s: %w", path, err)
		}
		text := string(data)
		for _, src := range srcDirs {
			text = strings.ReplaceAll(text, src+"/plugins/", dst+"/plugins/")
		}
		if text == string(data) {
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(text), info.Mode().Perm()); err != nil {
			return fmt.Errorf("failed to rewrite %s: %w", path, err)
		}
	}
	return nil
}

func sourcePrefixes(src string) []string {
	prefixes := []string{src}
	if resolved, err := filepath.EvalSymlinks(src); err == nil && resolved != src {
		prefixes = append(prefixes, resolved)
	}
	return prefixes
}
