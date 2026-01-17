package backup

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewBackupManager(t *testing.T) {
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")

	// Override HOME for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	bm, err := NewBackupManager(profilesDir)
	if err != nil {
		t.Fatalf("NewBackupManager failed: %v", err)
	}

	if bm.profilesDir != profilesDir {
		t.Errorf("profilesDir = %s, want %s", bm.profilesDir, profilesDir)
	}

	expectedBackupDir := filepath.Join(tmpDir, ".cdp", "backups")
	if bm.backupDir != expectedBackupDir {
		t.Errorf("backupDir = %s, want %s", bm.backupDir, expectedBackupDir)
	}

	// Check backup directory was created
	if _, err := os.Stat(bm.backupDir); os.IsNotExist(err) {
		t.Error("backup directory was not created")
	}
}

func TestBackup(t *testing.T) {
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")

	// Override HOME for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create a test profile
	testProfileDir := filepath.Join(profilesDir, "test-profile")
	if err := os.MkdirAll(testProfileDir, 0755); err != nil {
		t.Fatalf("failed to create test profile: %v", err)
	}

	// Create some files in the profile
	files := map[string]string{
		"settings.json":  `{"key": "value"}`,
		".claude.json":   `{"auth": "token"}`,
		".metadata.json": `{"createdAt": "2024-01-01"}`,
	}
	for name, content := range files {
		path := filepath.Join(testProfileDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", name, err)
		}
	}

	bm, err := NewBackupManager(profilesDir)
	if err != nil {
		t.Fatalf("NewBackupManager failed: %v", err)
	}

	// Test backup
	backupPath, err := bm.Backup("test-profile")
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("backup file was not created")
	}

	// Verify it's a valid tar.gz
	verifyTarGz(t, backupPath, files)
}

func TestBackupNonExistentProfile(t *testing.T) {
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("failed to create profiles dir: %v", err)
	}

	// Override HOME for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	bm, err := NewBackupManager(profilesDir)
	if err != nil {
		t.Fatalf("NewBackupManager failed: %v", err)
	}

	_, err = bm.Backup("non-existent")
	if err == nil {
		t.Error("expected error for non-existent profile")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRestore(t *testing.T) {
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")

	// Override HOME for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create a test profile and back it up
	testProfileDir := filepath.Join(profilesDir, "test-profile")
	if err := os.MkdirAll(testProfileDir, 0755); err != nil {
		t.Fatalf("failed to create test profile: %v", err)
	}

	files := map[string]string{
		"settings.json":  `{"key": "value"}`,
		".claude.json":   `{"auth": "token"}`,
		".metadata.json": `{"createdAt": "2024-01-01"}`,
	}
	for name, content := range files {
		path := filepath.Join(testProfileDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", name, err)
		}
	}

	bm, err := NewBackupManager(profilesDir)
	if err != nil {
		t.Fatalf("NewBackupManager failed: %v", err)
	}

	backupPath, err := bm.Backup("test-profile")
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Remove the original profile
	if err := os.RemoveAll(testProfileDir); err != nil {
		t.Fatalf("failed to remove test profile: %v", err)
	}

	// Restore the profile
	profileName, err := bm.Restore(backupPath, false)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	if profileName != "test-profile" {
		t.Errorf("restored profile name = %s, want test-profile", profileName)
	}

	// Verify files were restored
	for name, content := range files {
		path := filepath.Join(testProfileDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read restored file %s: %v", name, err)
			continue
		}
		if string(data) != content {
			t.Errorf("file %s content = %s, want %s", name, string(data), content)
		}
	}
}

func TestRestoreWithOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")

	// Override HOME for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create a test profile and back it up
	testProfileDir := filepath.Join(profilesDir, "test-profile")
	if err := os.MkdirAll(testProfileDir, 0755); err != nil {
		t.Fatalf("failed to create test profile: %v", err)
	}

	originalContent := `{"original": "content"}`
	if err := os.WriteFile(filepath.Join(testProfileDir, "settings.json"), []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	bm, err := NewBackupManager(profilesDir)
	if err != nil {
		t.Fatalf("NewBackupManager failed: %v", err)
	}

	backupPath, err := bm.Backup("test-profile")
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Modify the profile
	newContent := `{"modified": "content"}`
	if err := os.WriteFile(filepath.Join(testProfileDir, "settings.json"), []byte(newContent), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// Try to restore without overwrite - should fail
	_, err = bm.Restore(backupPath, false)
	if err == nil {
		t.Error("expected error when restoring without overwrite")
	}

	// Restore with overwrite - should succeed
	_, err = bm.Restore(backupPath, true)
	if err != nil {
		t.Fatalf("Restore with overwrite failed: %v", err)
	}

	// Verify original content was restored
	data, err := os.ReadFile(filepath.Join(testProfileDir, "settings.json"))
	if err != nil {
		t.Fatalf("failed to read restored file: %v", err)
	}
	if string(data) != originalContent {
		t.Errorf("restored content = %s, want %s", string(data), originalContent)
	}
}

func TestList(t *testing.T) {
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("failed to create profiles dir: %v", err)
	}

	// Override HOME for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	bm, err := NewBackupManager(profilesDir)
	if err != nil {
		t.Fatalf("NewBackupManager failed: %v", err)
	}

	// Test empty list
	backups, err := bm.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("expected 0 backups, got %d", len(backups))
	}

	// Create a profile and back it up
	testProfileDir := filepath.Join(profilesDir, "test-profile")
	if err := os.MkdirAll(testProfileDir, 0755); err != nil {
		t.Fatalf("failed to create test profile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testProfileDir, "settings.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err = bm.Backup("test-profile")
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// List again
	backups, err = bm.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(backups) != 1 {
		t.Errorf("expected 1 backup, got %d", len(backups))
	}
	if backups[0].ProfileName != "test-profile" {
		t.Errorf("backup profile name = %s, want test-profile", backups[0].ProfileName)
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")

	// Override HOME for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create a profile and back it up
	testProfileDir := filepath.Join(profilesDir, "test-profile")
	if err := os.MkdirAll(testProfileDir, 0755); err != nil {
		t.Fatalf("failed to create test profile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testProfileDir, "settings.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	bm, err := NewBackupManager(profilesDir)
	if err != nil {
		t.Fatalf("NewBackupManager failed: %v", err)
	}

	backupPath, err := bm.Backup("test-profile")
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Delete the backup
	backupName := filepath.Base(backupPath)
	err = bm.Delete(backupName)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("backup file still exists after delete")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("failed to create profiles dir: %v", err)
	}

	// Override HOME for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	bm, err := NewBackupManager(profilesDir)
	if err != nil {
		t.Fatalf("NewBackupManager failed: %v", err)
	}

	err = bm.Delete("non-existent.tar.gz")
	if err == nil {
		t.Error("expected error when deleting non-existent backup")
	}
}

func TestCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")

	// Override HOME for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create a profile and back it up
	testProfileDir := filepath.Join(profilesDir, "test-profile")
	if err := os.MkdirAll(testProfileDir, 0755); err != nil {
		t.Fatalf("failed to create test profile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testProfileDir, "settings.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	bm, err := NewBackupManager(profilesDir)
	if err != nil {
		t.Fatalf("NewBackupManager failed: %v", err)
	}

	// Create a backup
	_, err = bm.Backup("test-profile")
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Cleanup with 30 days retention - nothing should be deleted
	deleted, err := bm.Cleanup(30)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}

	// Cleanup with 0 days retention - should delete the backup
	deleted, err = bm.Cleanup(0)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}
}

func TestGetBackupDir(t *testing.T) {
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")

	// Override HOME for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	bm, err := NewBackupManager(profilesDir)
	if err != nil {
		t.Fatalf("NewBackupManager failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".cdp", "backups")
	if bm.GetBackupDir() != expected {
		t.Errorf("GetBackupDir() = %s, want %s", bm.GetBackupDir(), expected)
	}
}

func TestRestoreInvalidFilename(t *testing.T) {
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("failed to create profiles dir: %v", err)
	}

	// Override HOME for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	bm, err := NewBackupManager(profilesDir)
	if err != nil {
		t.Fatalf("NewBackupManager failed: %v", err)
	}

	// Create a fake backup file with invalid name format
	invalidBackupPath := filepath.Join(bm.backupDir, "invalid.tar.gz")
	if err := os.WriteFile(invalidBackupPath, []byte("fake"), 0644); err != nil {
		t.Fatalf("failed to write fake backup: %v", err)
	}

	_, err = bm.Restore(invalidBackupPath, false)
	if err == nil {
		t.Error("expected error for invalid backup filename")
	}
}

func TestBackupWithSubdirectories(t *testing.T) {
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")

	// Override HOME for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create a test profile with subdirectories
	testProfileDir := filepath.Join(profilesDir, "test-profile")
	subDir := filepath.Join(testProfileDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create test profile subdirectory: %v", err)
	}

	// Create files
	files := map[string]string{
		"settings.json":       `{"key": "value"}`,
		"subdir/nested.json":  `{"nested": true}`,
	}
	for name, content := range files {
		path := filepath.Join(testProfileDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", name, err)
		}
	}

	bm, err := NewBackupManager(profilesDir)
	if err != nil {
		t.Fatalf("NewBackupManager failed: %v", err)
	}

	// Backup
	backupPath, err := bm.Backup("test-profile")
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Remove original
	if err := os.RemoveAll(testProfileDir); err != nil {
		t.Fatalf("failed to remove test profile: %v", err)
	}

	// Restore
	_, err = bm.Restore(backupPath, false)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Verify nested file was restored
	nestedPath := filepath.Join(testProfileDir, "subdir", "nested.json")
	data, err := os.ReadFile(nestedPath)
	if err != nil {
		t.Fatalf("failed to read restored nested file: %v", err)
	}
	if string(data) != `{"nested": true}` {
		t.Errorf("nested file content = %s, want %s", string(data), `{"nested": true}`)
	}
}

func TestListSorting(t *testing.T) {
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")

	// Override HOME for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create a profile
	testProfileDir := filepath.Join(profilesDir, "test-profile")
	if err := os.MkdirAll(testProfileDir, 0755); err != nil {
		t.Fatalf("failed to create test profile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testProfileDir, "settings.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	bm, err := NewBackupManager(profilesDir)
	if err != nil {
		t.Fatalf("NewBackupManager failed: %v", err)
	}

	// Create multiple backups with delay
	_, err = bm.Backup("test-profile")
	if err != nil {
		t.Fatalf("First backup failed: %v", err)
	}

	time.Sleep(1100 * time.Millisecond) // Wait to get different timestamp

	_, err = bm.Backup("test-profile")
	if err != nil {
		t.Fatalf("Second backup failed: %v", err)
	}

	// List backups - should be sorted newest first
	backups, err := bm.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(backups) != 2 {
		t.Fatalf("expected 2 backups, got %d", len(backups))
	}

	// Newest should be first
	if !backups[0].CreatedAt.After(backups[1].CreatedAt) {
		t.Error("backups not sorted by creation time (newest first)")
	}
}

// Helper function to verify tar.gz contents
func verifyTarGz(t *testing.T, path string, expectedFiles map[string]string) {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open backup: %v", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	found := make(map[string]bool)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to read tar: %v", err)
		}

		if expectedContent, ok := expectedFiles[header.Name]; ok {
			found[header.Name] = true

			content, err := io.ReadAll(tarReader)
			if err != nil {
				t.Fatalf("failed to read file content: %v", err)
			}

			if string(content) != expectedContent {
				t.Errorf("file %s content = %s, want %s", header.Name, string(content), expectedContent)
			}
		}
	}

	for name := range expectedFiles {
		if !found[name] {
			t.Errorf("file %s not found in backup", name)
		}
	}
}
