package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all bot configuration.
type Config struct {
	Server   string `yaml:"server"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`

	Nick     string `yaml:"nick"`
	User     string `yaml:"user"`
	RealName string `yaml:"realname"`

	TLS           bool   `yaml:"tls"`
	TLSCA         string `yaml:"tls_ca"`
	TLSSkipVerify bool   `yaml:"tls_skip_verify"`

	Channels []string `yaml:"channels"`

	DBPath      string `yaml:"db_path"`
	MaxNotes    int    `yaml:"max_notes"`
	MaxNoteSize int    `yaml:"max_note_size"`
}

// Load loads configuration from a YAML file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	cfg := &Config{
		// Default values
		Port:        6667,
		Nick:        "notesbot",
		User:        "notesbot",
		RealName:    "IRC Notes Bot",
		DBPath:      "notes.db",
		MaxNotes:    15,
		MaxNoteSize: 4096,
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	if cfg.Server == "" {
		return nil, fmt.Errorf("server is not specified in configuration")
	}
	if len(cfg.Channels) == 0 {
		return nil, fmt.Errorf("no channels specified in configuration")
	}

	return cfg, nil
}
