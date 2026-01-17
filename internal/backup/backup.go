package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupManager handles profile backup and restore operations
type BackupManager struct {
	backupDir   string
	profilesDir string
}

// BackupInfo contains information about a backup
type BackupInfo struct {
	Name        string
	ProfileName string
	Path        string
	CreatedAt   time.Time
	Size        int64
}

// NewBackupManager creates a new backup manager
func NewBackupManager(profilesDir string) (*BackupManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	backupDir := filepath.Join(homeDir, ".cdp", "backups")

	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	return &BackupManager{
		backupDir:   backupDir,
		profilesDir: profilesDir,
	}, nil
}

// Backup creates a backup of the specified profile
func (bm *BackupManager) Backup(profileName string) (string, error) {
	profilePath := filepath.Join(bm.profilesDir, profileName)

	// Check if profile exists
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("profile '%s' does not exist", profileName)
	}

	// Generate backup filename
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s-%s.tar.gz", profileName, timestamp)
	backupPath := filepath.Join(bm.backupDir, backupName)

	// Create backup file
	file, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk the profile directory and add files to tar
	err = filepath.Walk(profilePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(profilePath, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// If it's a file, write content
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		os.Remove(backupPath) // Clean up on error
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	return backupPath, nil
}

// Restore restores a profile from a backup
func (bm *BackupManager) Restore(backupPath string, overwrite bool) (string, error) {
	// Open backup file
	file, err := os.Open(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to read gzip: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Extract profile name from backup filename
	baseName := filepath.Base(backupPath)
	// Remove .tar.gz extension and timestamp
	parts := strings.Split(strings.TrimSuffix(baseName, ".tar.gz"), "-")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid backup filename format")
	}
	// Profile name is everything except the last two parts (date-time)
	profileName := strings.Join(parts[:len(parts)-2], "-")

	profilePath := filepath.Join(bm.profilesDir, profileName)

	// Check if profile already exists
	if _, err := os.Stat(profilePath); err == nil {
		if !overwrite {
			return "", fmt.Errorf("profile '%s' already exists. Use --overwrite to replace", profileName)
		}
		// Remove existing profile
		if err := os.RemoveAll(profilePath); err != nil {
			return "", fmt.Errorf("failed to remove existing profile: %w", err)
		}
	}

	// Create profile directory
	if err := os.MkdirAll(profilePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create profile directory: %w", err)
	}

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			os.RemoveAll(profilePath) // Clean up on error
			return "", fmt.Errorf("failed to read tar: %w", err)
		}

		targetPath := filepath.Join(profilePath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				os.RemoveAll(profilePath)
				return "", fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				os.RemoveAll(profilePath)
				return "", fmt.Errorf("failed to create parent directory: %w", err)
			}

			outFile, err := os.Create(targetPath)
			if err != nil {
				os.RemoveAll(profilePath)
				return "", fmt.Errorf("failed to create file: %w", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				os.RemoveAll(profilePath)
				return "", fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()
		}
	}

	return profileName, nil
}

// List returns all available backups
func (bm *BackupManager) List() ([]BackupInfo, error) {
	entries, err := os.ReadDir(bm.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Parse profile name from filename
		baseName := strings.TrimSuffix(entry.Name(), ".tar.gz")
		parts := strings.Split(baseName, "-")
		if len(parts) < 3 {
			continue
		}
		profileName := strings.Join(parts[:len(parts)-2], "-")

		// Parse timestamp
		timestamp := parts[len(parts)-2] + "-" + parts[len(parts)-1]
		createdAt, _ := time.Parse("20060102-150405", timestamp)

		backups = append(backups, BackupInfo{
			Name:        entry.Name(),
			ProfileName: profileName,
			Path:        filepath.Join(bm.backupDir, entry.Name()),
			CreatedAt:   createdAt,
			Size:        info.Size(),
		})
	}

	// Sort by creation time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// Delete removes a backup file
func (bm *BackupManager) Delete(backupName string) error {
	backupPath := filepath.Join(bm.backupDir, backupName)

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup '%s' does not exist", backupName)
	}

	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	return nil
}

// GetBackupDir returns the backup directory path
func (bm *BackupManager) GetBackupDir() string {
	return bm.backupDir
}

// Cleanup removes backups older than the specified number of days
func (bm *BackupManager) Cleanup(retentionDays int) (int, error) {
	backups, err := bm.List()
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	deleted := 0

	for _, backup := range backups {
		if backup.CreatedAt.Before(cutoff) {
			if err := bm.Delete(backup.Name); err == nil {
				deleted++
			}
		}
	}

	return deleted, nil
}
