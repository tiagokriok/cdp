package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ConfigVersion   = "1.0"
	ConfigDirName   = ".cdp"
	ConfigFileName  = "config.yaml"
	ProfilesDirName = ".claude-profiles"
)

// Config represents the global CDP configuration
type Config struct {
	Version        string `yaml:"version"`
	ProfilesDir    string `yaml:"profilesDir"`
	CurrentProfile string `yaml:"currentProfile,omitempty"`
}

// GetConfigDir returns the CDP configuration directory path
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ConfigDirName), nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ConfigFileName), nil
}

// GetDefaultProfilesDir returns the default profiles directory path
func GetDefaultProfilesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ProfilesDirName), nil
}

// Init initializes the CDP configuration directory and creates a default config
func Init() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create profiles directory if it doesn't exist
	profilesDir, err := GetDefaultProfilesDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	// Check if config file already exists
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); err == nil {
		// Config already exists
		return nil
	}

	// Create default config
	cfg := &Config{
		Version:     ConfigVersion,
		ProfilesDir: profilesDir,
	}

	return cfg.Save()
}

// Load reads the configuration from disk
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found, run 'cdp init' first")
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to disk
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetProfilesDir returns the configured profiles directory
func (c *Config) GetProfilesDir() string {
	return c.ProfilesDir
}

// SetCurrentProfile sets the current active profile
func (c *Config) SetCurrentProfile(name string) error {
	c.CurrentProfile = name
	return c.Save()
}

// GetCurrentProfile returns the current active profile name
func (c *Config) GetCurrentProfile() string {
	return c.CurrentProfile
}

// Exists checks if the CDP configuration has been initialized
func Exists() bool {
	configPath, err := GetConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(configPath)
	return err == nil
}
