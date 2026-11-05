package backup

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

type tarEntry struct {
	name     string
	typeflag byte
	linkname string
	body     string
}

func writeTarGz(t *testing.T, path string, entries []tarEntry) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		h := &tar.Header{Name: e.name, Typeflag: e.typeflag, Linkname: e.linkname, Mode: 0644, Size: int64(len(e.body))}
		if e.typeflag == tar.TypeDir {
			h.Mode = 0755
		}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if e.body != "" {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	tw.Close()
	gz.Close()
}

func newTestManager(t *testing.T) (*BackupManager, string) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	bm, err := NewBackupManager(filepath.Join(tmp, "profiles"))
	if err != nil {
		t.Fatal(err)
	}
	return bm, tmp
}

func TestRestoreRejectsUnsafeEntries(t *testing.T) {
	tests := []struct {
		name  string
		entry tarEntry
	}{
		{"parent traversal", tarEntry{name: "../evil", typeflag: tar.TypeReg, body: "x"}},
		{"nested traversal", tarEntry{name: "a/../../evil", typeflag: tar.TypeReg, body: "x"}},
		{"absolute path", tarEntry{name: "/tmp/hat-evil-abs", typeflag: tar.TypeReg, body: "x"}},
		{"symlink outside", tarEntry{name: "link", typeflag: tar.TypeSymlink, linkname: "../../outside"}},
		{"symlink absolute", tarEntry{name: "link", typeflag: tar.TypeSymlink, linkname: "/etc/passwd"}},
		{"hardlink outside", tarEntry{name: "link", typeflag: tar.TypeLink, linkname: "../../outside"}},
		{"hardlink absolute", tarEntry{name: "link", typeflag: tar.TypeLink, linkname: "/etc/passwd"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm, tmp := newTestManager(t)
			archive := filepath.Join(bm.backupDir, "work-20240101-120000.tar.gz")
			writeTarGz(t, archive, []tarEntry{tt.entry})

			if _, err := bm.Restore(archive, false); err == nil {
				t.Fatal("expected error")
			}
			if _, err := os.Lstat(filepath.Join(tmp, "profiles", "evil")); err == nil {
				t.Error("file escaped profile dir")
			}
			if _, err := os.Lstat(filepath.Join(tmp, "profiles", "work")); err == nil {
				t.Error("partial profile not cleaned up")
			}
		})
	}
}

func TestRestoreAllowsSafeEntries(t *testing.T) {
	bm, tmp := newTestManager(t)
	archive := filepath.Join(bm.backupDir, "work-20240101-120000.tar.gz")
	writeTarGz(t, archive, []tarEntry{
		{name: "./", typeflag: tar.TypeDir},
		{name: "sub/", typeflag: tar.TypeDir},
		{name: "sub/a.json", typeflag: tar.TypeReg, body: "{}"},
		{name: "link", typeflag: tar.TypeSymlink, linkname: "sub/a.json"},
	})
	if _, err := bm.Restore(archive, false); err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "profiles", "work", "sub", "a.json")); err != nil {
		t.Error(err)
	}
}

func TestRestoreRejectsInvalidProfileName(t *testing.T) {
	for _, base := range []string{"a.b-20240101-120000.tar.gz", "we ird-20240101-120000.tar.gz"} {
		t.Run(base, func(t *testing.T) {
			bm, tmp := newTestManager(t)
			archive := filepath.Join(bm.backupDir, base)
			writeTarGz(t, archive, []tarEntry{{name: "f", typeflag: tar.TypeReg, body: "x"}})
			if _, err := bm.Restore(archive, false); err == nil {
				t.Fatal("expected error")
			}
			entries, _ := os.ReadDir(filepath.Join(tmp, "profiles"))
			if len(entries) != 0 {
				t.Errorf("profile dir created: %v", entries)
			}
		})
	}
}
