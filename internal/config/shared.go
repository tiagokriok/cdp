package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	SharedDirName  = "shared"
	SharedClaudeMD = "CLAUDE.md"
)

var sharedDirKinds = []string{"skills", "agents", "commands", "output-styles", "rules"}

type SyncResult struct {
	Linked    []string
	Conflicts []string
	Pruned    []string
}

func SharedKinds() []string {
	return append(append([]string{}, sharedDirKinds...), SharedClaudeMD)
}

func ValidateSharedKind(kind string) error {
	for _, k := range SharedKinds() {
		if k == kind {
			return nil
		}
	}
	return fmt.Errorf("invalid kind '%s', valid: %s", kind, strings.Join(SharedKinds(), ", "))
}

func GetSharedDir() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, SharedDirName), nil
}

func ListShared() (map[string][]string, error) {
	sd, err := GetSharedDir()
	if err != nil {
		return nil, err
	}
	out := map[string][]string{}
	for _, kind := range sharedDirKinds {
		entries, err := os.ReadDir(filepath.Join(sd, kind))
		if err != nil {
			continue
		}
		for _, e := range entries {
			out[kind] = append(out[kind], e.Name())
		}
	}
	if _, err := os.Lstat(filepath.Join(sd, SharedClaudeMD)); err == nil {
		out[SharedClaudeMD] = []string{SharedClaudeMD}
	}
	return out, nil
}

func (pm *ProfileManager) SetSharedKinds(name string, kinds []string, enable bool) error {
	for _, k := range kinds {
		if err := ValidateSharedKind(k); err != nil {
			return err
		}
	}
	profile, err := pm.GetProfile(name)
	if err != nil {
		return err
	}
	changing := map[string]bool{}
	for _, k := range kinds {
		changing[k] = true
	}
	seen := map[string]bool{}
	var next []string
	for _, k := range profile.Metadata.Shared {
		if !seen[k] && !(changing[k] && !enable) {
			next = append(next, k)
			seen[k] = true
		}
	}
	if enable {
		for _, k := range kinds {
			if !seen[k] {
				next = append(next, k)
				seen[k] = true
			}
		}
	}
	profile.Metadata.Shared = next
	return pm.saveMetadata(profile.Path, profile.Metadata)
}

func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode()&os.ModeSymlink != 0
}

func isLinkTo(link, target string) bool {
	got, err := os.Readlink(link)
	return err == nil && got == target
}

func pointsIntoShared(link, sharedDir string) bool {
	target, err := os.Readlink(link)
	if err != nil {
		return false
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(link), target)
	}
	return strings.HasPrefix(filepath.Clean(target), sharedDir+string(filepath.Separator))
}

func (pm *ProfileManager) SyncShared(name string) (SyncResult, error) {
	var res SyncResult
	profile, err := pm.GetProfile(name)
	if err != nil {
		return res, err
	}
	sd, err := GetSharedDir()
	if err != nil {
		return res, err
	}
	opted := map[string]bool{}
	for _, k := range profile.Metadata.Shared {
		opted[k] = true
	}

	for _, kind := range sharedDirKinds {
		dir := filepath.Join(profile.Path, kind)
		if err := pruneDir(dir, kind, sd, opted[kind], &res); err != nil {
			return res, err
		}
		if opted[kind] {
			if err := linkDir(dir, filepath.Join(sd, kind), kind, &res); err != nil {
				return res, err
			}
		}
	}
	return res, syncClaudeMD(profile.Path, sd, opted[SharedClaudeMD], &res)
}

func syncClaudeMD(profilePath, sd string, opted bool, res *SyncResult) error {
	dst := filepath.Join(profilePath, SharedClaudeMD)
	src := filepath.Join(sd, SharedClaudeMD)
	if isSymlink(dst) && pointsIntoShared(dst, sd) {
		if _, err := os.Stat(dst); err != nil || !opted {
			if err := os.Remove(dst); err != nil {
				return fmt.Errorf("failed to prune %s: %w", dst, err)
			}
			res.Pruned = append(res.Pruned, SharedClaudeMD)
		}
	}
	if !opted {
		return nil
	}
	if _, err := os.Stat(src); err != nil {
		return nil
	}
	if _, err := os.Lstat(dst); err == nil {
		if !isLinkTo(dst, src) {
			res.Conflicts = append(res.Conflicts, SharedClaudeMD)
		}
		return nil
	}
	if err := os.Symlink(src, dst); err != nil {
		return fmt.Errorf("failed to link %s: %w", dst, err)
	}
	res.Linked = append(res.Linked, SharedClaudeMD)
	return nil
}

func pruneDir(dir, kind, sd string, keep bool, res *SyncResult) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		if !isSymlink(path) || !pointsIntoShared(path, sd) {
			continue
		}
		if _, err := os.Stat(path); err == nil && keep {
			continue
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to prune %s: %w", path, err)
		}
		res.Pruned = append(res.Pruned, kind+"/"+e.Name())
	}
	return nil
}

func linkDir(dir, sharedKindDir, kind string, res *SyncResult) error {
	entries, err := os.ReadDir(sharedKindDir)
	if err != nil || len(entries) == 0 {
		return nil
	}
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		if shared, err := filepath.EvalSymlinks(sharedKindDir); err == nil && resolved == shared {
			return nil
		}
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", dir, err)
	}
	for _, e := range entries {
		src := filepath.Join(sharedKindDir, e.Name())
		dst := filepath.Join(dir, e.Name())
		if _, err := os.Lstat(dst); err == nil {
			if !isLinkTo(dst, src) {
				res.Conflicts = append(res.Conflicts, kind+"/"+e.Name())
			}
			continue
		}
		if err := os.Symlink(src, dst); err != nil {
			return fmt.Errorf("failed to link %s: %w", dst, err)
		}
		res.Linked = append(res.Linked, kind+"/"+e.Name())
	}
	return nil
}

func (pm *ProfileManager) SyncAllShared() (map[string]SyncResult, error) {
	profiles, err := pm.ListProfiles()
	if err != nil {
		return nil, err
	}
	out := map[string]SyncResult{}
	for _, p := range profiles {
		res, err := pm.SyncShared(p.Name)
		if err != nil {
			return out, fmt.Errorf("%s: %w", p.Name, err)
		}
		out[p.Name] = res
	}
	return out, nil
}

func (pm *ProfileManager) AddToShared(kind, path string) (string, error) {
	if err := ValidateSharedKind(kind); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return "", fmt.Errorf("path not found: %s", abs)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("%s is already a symlink", abs)
	}

	owner := filepath.Dir(abs)
	dest := filepath.Join(sharedRoot(), kind, filepath.Base(abs))
	if kind == SharedClaudeMD {
		if filepath.Base(abs) != SharedClaudeMD {
			return "", fmt.Errorf("%s is not a %s file", abs, SharedClaudeMD)
		}
		dest = filepath.Join(sharedRoot(), SharedClaudeMD)
	} else {
		if filepath.Base(owner) != kind {
			return "", fmt.Errorf("%s is not inside a '%s' directory", abs, kind)
		}
		owner = filepath.Dir(owner)
	}
	profile, err := pm.profileAt(owner)
	if err != nil {
		return "", err
	}
	if _, err := os.Lstat(dest); err == nil {
		return "", fmt.Errorf("%s already exists in shared", dest)
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return "", fmt.Errorf("failed to create shared dir: %w", err)
	}
	if err := os.Rename(abs, dest); err != nil {
		return "", fmt.Errorf("failed to move into shared: %w", err)
	}
	if err := os.Symlink(dest, abs); err != nil {
		os.Rename(dest, abs)
		return "", fmt.Errorf("failed to link back: %w", err)
	}
	return profile.Name, pm.SetSharedKinds(profile.Name, []string{kind}, true)
}

func sharedRoot() string {
	sd, _ := GetSharedDir()
	return sd
}

func (pm *ProfileManager) profileAt(dir string) (*Profile, error) {
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve %s: %w", dir, err)
	}
	profiles, err := pm.ListProfiles()
	if err != nil {
		return nil, err
	}
	for i := range profiles {
		if r, err := filepath.EvalSymlinks(profiles[i].Path); err == nil && r == resolved {
			return &profiles[i], nil
		}
	}
	return nil, fmt.Errorf("%s is not inside a hat profile", dir)
}
