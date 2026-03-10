package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return path
}

func TestLoad_FullConfig(t *testing.T) {
	yaml := `
server: "irc.libera.chat"
port: 6697
password: "secret"
nick: "mybot"
user: "mybot"
realname: "My Bot"
tls: true
tls_ca: "/etc/ssl/ca.pem"
tls_skip_verify: true
channels:
  - "#test"
  - "#dev"
db_path: "/tmp/test.db"
max_notes: 20
max_note_size: 2048
`
	cfg, err := Load(writeTempConfig(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server != "irc.libera.chat" {
		t.Errorf("Server = %q, want %q", cfg.Server, "irc.libera.chat")
	}
	if cfg.Port != 6697 {
		t.Errorf("Port = %d, want %d", cfg.Port, 6697)
	}
	if cfg.Password != "secret" {
		t.Errorf("Password = %q, want %q", cfg.Password, "secret")
	}
	if cfg.Nick != "mybot" {
		t.Errorf("Nick = %q, want %q", cfg.Nick, "mybot")
	}
	if cfg.User != "mybot" {
		t.Errorf("User = %q, want %q", cfg.User, "mybot")
	}
	if cfg.RealName != "My Bot" {
		t.Errorf("RealName = %q, want %q", cfg.RealName, "My Bot")
	}
	if !cfg.TLS {
		t.Error("TLS = false, want true")
	}
	if cfg.TLSCA != "/etc/ssl/ca.pem" {
		t.Errorf("TLSCA = %q, want %q", cfg.TLSCA, "/etc/ssl/ca.pem")
	}
	if !cfg.TLSSkipVerify {
		t.Error("TLSSkipVerify = false, want true")
	}
	if len(cfg.Channels) != 2 || cfg.Channels[0] != "#test" || cfg.Channels[1] != "#dev" {
		t.Errorf("Channels = %v, want [#test #dev]", cfg.Channels)
	}
	if cfg.DBPath != "/tmp/test.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "/tmp/test.db")
	}
	if cfg.MaxNotes != 20 {
		t.Errorf("MaxNotes = %d, want %d", cfg.MaxNotes, 20)
	}
	if cfg.MaxNoteSize != 2048 {
		t.Errorf("MaxNoteSize = %d, want %d", cfg.MaxNoteSize, 2048)
	}
}

func TestLoad_Defaults(t *testing.T) {
	yaml := `
server: "irc.example.com"
channels:
  - "#general"
`
	cfg, err := Load(writeTempConfig(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 6667 {
		t.Errorf("Port = %d, want default %d", cfg.Port, 6667)
	}
	if cfg.Nick != "notesbot" {
		t.Errorf("Nick = %q, want default %q", cfg.Nick, "notesbot")
	}
	if cfg.User != "notesbot" {
		t.Errorf("User = %q, want default %q", cfg.User, "notesbot")
	}
	if cfg.RealName != "IRC Notes Bot" {
		t.Errorf("RealName = %q, want default %q", cfg.RealName, "IRC Notes Bot")
	}
	if cfg.DBPath != "notes.db" {
		t.Errorf("DBPath = %q, want default %q", cfg.DBPath, "notes.db")
	}
	if cfg.MaxNotes != 15 {
		t.Errorf("MaxNotes = %d, want default %d", cfg.MaxNotes, 15)
	}
	if cfg.MaxNoteSize != 4096 {
		t.Errorf("MaxNoteSize = %d, want default %d", cfg.MaxNoteSize, 4096)
	}
	if cfg.TLS {
		t.Error("TLS = true, want default false")
	}
	if cfg.Password != "" {
		t.Errorf("Password = %q, want default empty", cfg.Password)
	}
}

func TestLoad_MissingServer(t *testing.T) {
	yaml := `
channels:
  - "#general"
`
	_, err := Load(writeTempConfig(t, yaml))
	if err == nil {
		t.Fatal("expected error for missing server, got nil")
	}
}

func TestLoad_MissingChannels(t *testing.T) {
	yaml := `
server: "irc.example.com"
`
	_, err := Load(writeTempConfig(t, yaml))
	if err == nil {
		t.Fatal("expected error for missing channels, got nil")
	}
}

func TestLoad_EmptyChannels(t *testing.T) {
	yaml := `
server: "irc.example.com"
channels: []
`
	_, err := Load(writeTempConfig(t, yaml))
	if err == nil {
		t.Fatal("expected error for empty channels, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	content := `
server: "test"
channels:
  - "#foo"
  invalid_indent
`
	_, err := Load(writeTempConfig(t, content))
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}
