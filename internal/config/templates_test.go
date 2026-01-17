package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTemplateManager(t *testing.T) {
	tm := NewTemplateManager()
	if tm == nil {
		t.Fatal("NewTemplateManager() returned nil")
	}
	if tm.customTemplatesDir == "" {
		t.Error("customTemplatesDir should not be empty")
	}
}

func TestListTemplates(t *testing.T) {
	tm := NewTemplateManager()
	templates, err := tm.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}

	// Should have at least the built-in templates
	if len(templates) < 2 {
		t.Errorf("ListTemplates() returned %d templates, want at least 2", len(templates))
	}

	// Check for expected templates
	hasRestrictive := false
	hasPermissive := false
	for _, name := range templates {
		if name == "restrictive" {
			hasRestrictive = true
		}
		if name == "permissive" {
			hasPermissive = true
		}
	}

	if !hasRestrictive {
		t.Error("ListTemplates() missing 'restrictive' template")
	}
	if !hasPermissive {
		t.Error("ListTemplates() missing 'permissive' template")
	}
}

func TestLoadTemplate_Restrictive(t *testing.T) {
	tm := NewTemplateManager()
	template, err := tm.LoadTemplate("restrictive")
	if err != nil {
		t.Fatalf("LoadTemplate('restrictive') error = %v", err)
	}

	if template.Name != "restrictive" {
		t.Errorf("template.Name = %q, want 'restrictive'", template.Name)
	}

	if template.Content == nil {
		t.Error("template.Content should not be nil")
	}

	// Check that permissions key exists
	if _, ok := template.Content["permissions"]; !ok {
		t.Error("template.Content should have 'permissions' key")
	}
}

func TestLoadTemplate_Permissive(t *testing.T) {
	tm := NewTemplateManager()
	template, err := tm.LoadTemplate("permissive")
	if err != nil {
		t.Fatalf("LoadTemplate('permissive') error = %v", err)
	}

	if template.Name != "permissive" {
		t.Errorf("template.Name = %q, want 'permissive'", template.Name)
	}
}

func TestLoadTemplate_NotFound(t *testing.T) {
	tm := NewTemplateManager()
	_, err := tm.LoadTemplate("nonexistent")
	if err == nil {
		t.Error("LoadTemplate('nonexistent') should return error")
	}
}

func TestTemplateExists(t *testing.T) {
	tm := NewTemplateManager()

	if !tm.TemplateExists("restrictive") {
		t.Error("TemplateExists('restrictive') should return true")
	}

	if !tm.TemplateExists("permissive") {
		t.Error("TemplateExists('permissive') should return true")
	}

	if tm.TemplateExists("nonexistent") {
		t.Error("TemplateExists('nonexistent') should return false")
	}
}

func TestApplyTemplate(t *testing.T) {
	// Create temp directory for test profile
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "test-profile")
	os.MkdirAll(profilePath, 0755)

	// Create empty settings file
	settingsPath := filepath.Join(profilePath, ClaudeSettingsFile)
	os.WriteFile(settingsPath, []byte("{}"), 0644)

	tm := NewTemplateManager()
	err := tm.ApplyTemplate(profilePath, "restrictive")
	if err != nil {
		t.Fatalf("ApplyTemplate() error = %v", err)
	}

	// Read the settings file and verify template was applied
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings file: %v", err)
	}

	content := string(data)
	if len(content) < 10 {
		t.Error("Settings file should contain template content")
	}
}

func TestApplyTemplate_InvalidTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "test-profile")
	os.MkdirAll(profilePath, 0755)

	tm := NewTemplateManager()
	err := tm.ApplyTemplate(profilePath, "nonexistent")
	if err == nil {
		t.Error("ApplyTemplate() with invalid template should return error")
	}
}

func TestCreateProfileWithTemplate(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Initialize config
	if err := Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	pm := NewProfileManager(cfg)

	// Create profile with template
	err = pm.CreateProfileWithTemplate("test-profile", "Test description", "restrictive")
	if err != nil {
		t.Fatalf("CreateProfileWithTemplate() error = %v", err)
	}

	// Verify profile was created
	profile, err := pm.GetProfile("test-profile")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if profile.Metadata.Template != "restrictive" {
		t.Errorf("profile.Metadata.Template = %q, want 'restrictive'", profile.Metadata.Template)
	}

	// Verify settings file has template content
	settingsPath := filepath.Join(profile.Path, ClaudeSettingsFile)
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings: %v", err)
	}

	if len(data) < 10 {
		t.Error("Settings file should contain template content")
	}
}

func TestCreateProfileWithTemplate_InvalidTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	Init()
	cfg, _ := Load()
	pm := NewProfileManager(cfg)

	err := pm.CreateProfileWithTemplate("test-profile", "Test", "nonexistent")
	if err == nil {
		t.Error("CreateProfileWithTemplate() with invalid template should return error")
	}
}
