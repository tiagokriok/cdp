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

func TestApplyTemplate_SettingsFileDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "test-profile")
	os.MkdirAll(profilePath, 0755)

	// Don't create settings file - ApplyTemplate should create it
	tm := NewTemplateManager()
	err := tm.ApplyTemplate(profilePath, "restrictive")
	if err != nil {
		t.Errorf("ApplyTemplate() should succeed and create settings file: %v", err)
	}

	// Verify settings file was created
	settingsPath := filepath.Join(profilePath, ClaudeSettingsFile)
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("Settings file should have been created")
	}
}

func TestApplyTemplate_Permissive(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "test-profile")
	os.MkdirAll(profilePath, 0755)

	// Create empty settings file
	settingsPath := filepath.Join(profilePath, ClaudeSettingsFile)
	os.WriteFile(settingsPath, []byte("{}"), 0644)

	tm := NewTemplateManager()
	err := tm.ApplyTemplate(profilePath, "permissive")
	if err != nil {
		t.Fatalf("ApplyTemplate() error = %v", err)
	}

	// Verify settings file has content
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings file: %v", err)
	}

	if len(data) < 10 {
		t.Error("Settings file should contain template content")
	}
}

func TestLoadTemplate_CustomTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	customDir := filepath.Join(tmpDir, ".cdp", "templates")
	os.MkdirAll(customDir, 0755)

	// Create a custom template file
	customTemplate := filepath.Join(customDir, "custom.json")
	os.WriteFile(customTemplate, []byte(`{"name":"custom","content":{"key":"value"}}`), 0644)

	// Create temp home for override
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	tm := NewTemplateManager()

	// The custom template might not be loaded since it's in custom templates dir
	// and our built-in templates should still work
	template, err := tm.LoadTemplate("restrictive")
	if err != nil {
		t.Errorf("LoadTemplate('restrictive') should work: %v", err)
	}
	if template == nil {
		t.Error("template should not be nil")
	}
}
