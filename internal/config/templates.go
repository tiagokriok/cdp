package config

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/*.json
var embeddedTemplates embed.FS

// Template represents a profile template
type Template struct {
	Name    string                 `json:"name"`
	Content map[string]interface{} `json:"content"`
}

// TemplateManager handles template operations
type TemplateManager struct {
	customTemplatesDir string
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	homeDir, _ := os.UserHomeDir()
	return &TemplateManager{
		customTemplatesDir: filepath.Join(homeDir, ".cdp", "templates"),
	}
}

// ListTemplates returns all available templates (built-in + custom)
func (tm *TemplateManager) ListTemplates() ([]string, error) {
	templates := []string{}

	// List built-in templates
	entries, err := embeddedTemplates.ReadDir("templates")
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				name := strings.TrimSuffix(entry.Name(), ".json")
				templates = append(templates, name)
			}
		}
	}

	// List custom templates
	if _, err := os.Stat(tm.customTemplatesDir); err == nil {
		entries, err := os.ReadDir(tm.customTemplatesDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
					name := strings.TrimSuffix(entry.Name(), ".json")
					// Avoid duplicates
					found := false
					for _, t := range templates {
						if t == name {
							found = true
							break
						}
					}
					if !found {
						templates = append(templates, name)
					}
				}
			}
		}
	}

	return templates, nil
}

// LoadTemplate loads a template by name
func (tm *TemplateManager) LoadTemplate(name string) (*Template, error) {
	// Try custom templates first
	customPath := filepath.Join(tm.customTemplatesDir, name+".json")
	if _, err := os.Stat(customPath); err == nil {
		return tm.loadTemplateFromFile(customPath, name)
	}

	// Try built-in templates
	data, err := embeddedTemplates.ReadFile("templates/" + name + ".json")
	if err != nil {
		return nil, fmt.Errorf("template '%s' not found", name)
	}

	var content map[string]interface{}
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("failed to parse template '%s': %w", name, err)
	}

	return &Template{
		Name:    name,
		Content: content,
	}, nil
}

// loadTemplateFromFile loads a template from a file path
func (tm *TemplateManager) loadTemplateFromFile(path, name string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var content map[string]interface{}
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &Template{
		Name:    name,
		Content: content,
	}, nil
}

// ApplyTemplate applies a template to a profile's settings.json
func (tm *TemplateManager) ApplyTemplate(profilePath, templateName string) error {
	template, err := tm.LoadTemplate(templateName)
	if err != nil {
		return err
	}

	settingsPath := filepath.Join(profilePath, ClaudeSettingsFile)

	// Load existing settings if any
	var settings map[string]interface{}
	if data, err := os.ReadFile(settingsPath); err == nil {
		json.Unmarshal(data, &settings)
	}
	if settings == nil {
		settings = make(map[string]interface{})
	}

	// Merge template content into settings
	for key, value := range template.Content {
		settings[key] = value
	}

	// Write back
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// TemplateExists checks if a template exists
func (tm *TemplateManager) TemplateExists(name string) bool {
	_, err := tm.LoadTemplate(name)
	return err == nil
}
