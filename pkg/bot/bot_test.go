package bot

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/identw/irc-notes-bot/pkg/config"
	"github.com/identw/irc-notes-bot/pkg/db"
)

func newTestBot(t *testing.T) *Bot {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := db.NewNoteStore(dbPath, 15)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	cfg := &config.Config{
		Server:      "irc.example.com",
		Port:        6667,
		Nick:        "testbot",
		User:        "testbot",
		RealName:    "Test Bot",
		Channels:    []string{"#test"},
		DBPath:      dbPath,
		MaxNotes:    15,
		MaxNoteSize: 4096,
	}

	b, err := New(cfg, store)
	if err != nil {
		t.Fatalf("failed to create bot: %v", err)
	}
	return b
}

func TestNew_Basic(t *testing.T) {
	b := newTestBot(t)
	if b == nil {
		t.Fatal("bot is nil")
	}
	if b.Client == nil {
		t.Fatal("Client is nil")
	}
}

func TestNew_WithPassword(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := db.NewNoteStore(dbPath, 15)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	cfg := &config.Config{
		Server:      "irc.example.com",
		Port:        6667,
		Password:    "secret",
		Nick:        "testbot",
		User:        "testbot",
		RealName:    "Test Bot",
		Channels:    []string{"#test"},
		MaxNotes:    15,
		MaxNoteSize: 4096,
	}

	b, err := New(cfg, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b == nil {
		t.Fatal("bot is nil")
	}
}

func TestNew_WithTLS(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := db.NewNoteStore(dbPath, 15)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	cfg := &config.Config{
		Server:        "irc.example.com",
		Port:          6697,
		Nick:          "testbot",
		User:          "testbot",
		RealName:      "Test Bot",
		TLS:           true,
		TLSSkipVerify: true,
		Channels:      []string{"#test"},
		MaxNotes:      15,
		MaxNoteSize:   4096,
	}

	b, err := New(cfg, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b == nil {
		t.Fatal("bot is nil")
	}
}

func TestNew_WithInvalidCA(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := db.NewNoteStore(dbPath, 15)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	cfg := &config.Config{
		Server:      "irc.example.com",
		Port:        6697,
		Nick:        "testbot",
		User:        "testbot",
		RealName:    "Test Bot",
		TLS:         true,
		TLSCA:       "/nonexistent/ca.pem",
		Channels:    []string{"#test"},
		MaxNotes:    15,
		MaxNoteSize: 4096,
	}

	_, err = New(cfg, store)
	if err == nil {
		t.Fatal("expected error for invalid CA path, got nil")
	}
}

func TestNew_WithBadCACert(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "bad-ca.pem")
	if err := writeFile(caPath, "not a real certificate"); err != nil {
		t.Fatalf("failed to write fake CA: %v", err)
	}

	dbPath := filepath.Join(dir, "test.db")
	store, err := db.NewNoteStore(dbPath, 15)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	cfg := &config.Config{
		Server:      "irc.example.com",
		Port:        6697,
		Nick:        "testbot",
		User:        "testbot",
		RealName:    "Test Bot",
		TLS:         true,
		TLSCA:       caPath,
		Channels:    []string{"#test"},
		MaxNotes:    15,
		MaxNoteSize: 4096,
	}

	_, err = New(cfg, store)
	if err == nil {
		t.Fatal("expected error for bad CA cert, got nil")
	}
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
