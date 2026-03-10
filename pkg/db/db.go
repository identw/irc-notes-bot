package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Note represents a single notification.
type Note struct {
	ID        int64
	Channel   string
	Author    string
	Text      string
	CreatedAt time.Time
}

// NoteStore manages notes storage in SQLite.
type NoteStore struct {
	db       *sql.DB
	maxNotes int
}

// NewNoteStore creates a new notes store.
func NewNoteStore(dbPath string, maxNotes int) (*NoteStore, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	store := &NoteStore{
		db:       db,
		maxNotes: maxNotes,
	}

	if err := store.migrate(); err != nil {
		return nil, err
	}

	return store, nil
}

// migrate creates tables if they do not exist yet.
func (s *NoteStore) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		channel TEXT NOT NULL,
		author TEXT NOT NULL,
		text TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT (datetime('now'))
	);
	CREATE INDEX IF NOT EXISTS idx_notes_channel ON notes(channel);
	`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}

// AddNote adds a note for a channel. Implements a ring buffer:
// if the number of notes for a channel >= maxNotes, the oldest ones are deleted.
func (s *NoteStore) AddNote(channel, author, text string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert new note
	_, err = tx.Exec(
		"INSERT INTO notes (channel, author, text, created_at) VALUES (?, ?, ?, ?)",
		channel, author, text, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to add note: %w", err)
	}

	// Delete old entries if the limit is exceeded (ring buffer)
	_, err = tx.Exec(`
		DELETE FROM notes WHERE channel = ? AND id NOT IN (
			SELECT id FROM notes WHERE channel = ? ORDER BY id DESC LIMIT ?
		)
	`, channel, channel, s.maxNotes)
	if err != nil {
		return fmt.Errorf("failed to clean up old notes: %w", err)
	}

	return tx.Commit()
}

// ListNotes returns all notes for a channel, sorted by creation date.
func (s *NoteStore) ListNotes(channel string) ([]Note, error) {
	rows, err := s.db.Query(
		"SELECT id, channel, author, text, created_at FROM notes WHERE channel = ? ORDER BY id ASC",
		channel,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve notes: %w", err)
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Channel, &n.Author, &n.Text, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// CountNotes returns the number of notes for a channel.
func (s *NoteStore) CountNotes(channel string) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM notes WHERE channel = ?", channel).Scan(&count)
	return count, err
}

// Close closes the database connection.
func (s *NoteStore) Close() error {
	return s.db.Close()
}
