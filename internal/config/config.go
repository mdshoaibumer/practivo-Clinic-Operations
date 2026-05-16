package config

import (
	"os"
	"path/filepath"
)

// Config holds all application configuration values including file paths,
// security settings, and session parameters.
type Config struct {
	AppName          string
	Version          string
	DataDir          string
	DBPath           string
	BackupDir        string
	LogDir           string
	MaxLoginAttempts int
	LockoutMinutes   int
	SessionHours     int
	BcryptCost       int
}

// NewConfig creates a new Config with default values and ensures all
// required directories (data, backup, logs) exist on disk.
func NewConfig() (*Config, error) {
	dataDir, err := getDataDir()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		AppName:          "Clinmitra Dental",
		Version:          "1.0.0",
		DataDir:          dataDir,
		DBPath:           filepath.Join(dataDir, "clinmitra.db"),
		BackupDir:        filepath.Join(dataDir, "backups"),
		LogDir:           filepath.Join(dataDir, "logs"),
		MaxLoginAttempts: 5,
		LockoutMinutes:   15,
		SessionHours:     8,
		BcryptCost:       12,
	}

	if err := cfg.ensureDirectories(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// getDataDir returns the platform-specific application data directory
// (e.g., %APPDATA%/Clinmitra Dental on Windows).
func getDataDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "ClinmitraDental"), nil
}

// ensureDirectories creates the data, backup, and log directories with
// restrictive permissions (0700) if they do not already exist.
func (c *Config) ensureDirectories() error {
	dirs := []string{c.DataDir, c.BackupDir, c.LogDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	return nil
}
