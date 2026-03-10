package db

import (
	"fmt"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T, maxNotes int) *NoteStore {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewNoteStore(dbPath, maxNotes)
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestNewNoteStore_CreatesDB(t *testing.T) {
	store := newTestStore(t, 15)
	if store == nil {
		t.Fatal("store is nil")
	}
}

func TestAddNote_And_ListNotes(t *testing.T) {
	store := newTestStore(t, 15)

	if err := store.AddNote("#test", "alice", "hello world"); err != nil {
		t.Fatalf("AddNote failed: %v", err)
	}

	notes, err := store.ListNotes("#test")
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
	if notes[0].Channel != "#test" {
		t.Errorf("Channel = %q, want %q", notes[0].Channel, "#test")
	}
	if notes[0].Author != "alice" {
		t.Errorf("Author = %q, want %q", notes[0].Author, "alice")
	}
	if notes[0].Text != "hello world" {
		t.Errorf("Text = %q, want %q", notes[0].Text, "hello world")
	}
	if notes[0].CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestListNotes_Empty(t *testing.T) {
	store := newTestStore(t, 15)

	notes, err := store.ListNotes("#empty")
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}
	if len(notes) != 0 {
		t.Fatalf("expected 0 notes, got %d", len(notes))
	}
}

func TestCountNotes(t *testing.T) {
	store := newTestStore(t, 15)

	count, err := store.CountNotes("#test")
	if err != nil {
		t.Fatalf("CountNotes failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	for i := 0; i < 5; i++ {
		if err := store.AddNote("#test", "bob", "note"); err != nil {
			t.Fatalf("AddNote failed: %v", err)
		}
	}

	count, err = store.CountNotes("#test")
	if err != nil {
		t.Fatalf("CountNotes failed: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5, got %d", count)
	}
}

func TestRingBuffer_ExceedLimit(t *testing.T) {
	maxNotes := 3
	store := newTestStore(t, maxNotes)

	// Add 5 notes, only last 3 should remain
	for i := 1; i <= 5; i++ {
		if err := store.AddNote("#ring", "user", fmt.Sprintf("note-%d", i)); err != nil {
			t.Fatalf("AddNote %d failed: %v", i, err)
		}
	}

	notes, err := store.ListNotes("#ring")
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}
	if len(notes) != maxNotes {
		t.Fatalf("expected %d notes, got %d", maxNotes, len(notes))
	}

	// The oldest notes (note-1, note-2) should be deleted
	if notes[0].Text != "note-3" {
		t.Errorf("notes[0].Text = %q, want %q", notes[0].Text, "note-3")
	}
	if notes[1].Text != "note-4" {
		t.Errorf("notes[1].Text = %q, want %q", notes[1].Text, "note-4")
	}
	if notes[2].Text != "note-5" {
		t.Errorf("notes[2].Text = %q, want %q", notes[2].Text, "note-5")
	}
}

func TestRingBuffer_ExactLimit(t *testing.T) {
	maxNotes := 3
	store := newTestStore(t, maxNotes)

	for i := 1; i <= 3; i++ {
		if err := store.AddNote("#exact", "user", fmt.Sprintf("note-%d", i)); err != nil {
			t.Fatalf("AddNote failed: %v", err)
		}
	}

	count, err := store.CountNotes("#exact")
	if err != nil {
		t.Fatalf("CountNotes failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3, got %d", count)
	}
}

func TestRingBuffer_SingleSlot(t *testing.T) {
	store := newTestStore(t, 1)

	if err := store.AddNote("#one", "user", "first"); err != nil {
		t.Fatalf("AddNote failed: %v", err)
	}
	if err := store.AddNote("#one", "user", "second"); err != nil {
		t.Fatalf("AddNote failed: %v", err)
	}

	notes, err := store.ListNotes("#one")
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
	if notes[0].Text != "second" {
		t.Errorf("Text = %q, want %q", notes[0].Text, "second")
	}
}

func TestChannelIsolation(t *testing.T) {
	store := newTestStore(t, 15)

	if err := store.AddNote("#chan1", "alice", "note for chan1"); err != nil {
		t.Fatalf("AddNote failed: %v", err)
	}
	if err := store.AddNote("#chan2", "bob", "note for chan2"); err != nil {
		t.Fatalf("AddNote failed: %v", err)
	}

	notes1, err := store.ListNotes("#chan1")
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}
	notes2, err := store.ListNotes("#chan2")
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}

	if len(notes1) != 1 {
		t.Errorf("chan1: expected 1 note, got %d", len(notes1))
	}
	if len(notes2) != 1 {
		t.Errorf("chan2: expected 1 note, got %d", len(notes2))
	}
	if notes1[0].Text != "note for chan1" {
		t.Errorf("chan1 text = %q, want %q", notes1[0].Text, "note for chan1")
	}
	if notes2[0].Text != "note for chan2" {
		t.Errorf("chan2 text = %q, want %q", notes2[0].Text, "note for chan2")
	}
}

func TestRingBuffer_ChannelIsolation(t *testing.T) {
	store := newTestStore(t, 2)

	// Fill #chan1 to the limit
	for i := 1; i <= 3; i++ {
		if err := store.AddNote("#chan1", "user", fmt.Sprintf("c1-%d", i)); err != nil {
			t.Fatalf("AddNote failed: %v", err)
		}
	}
	// Add one note to #chan2
	if err := store.AddNote("#chan2", "user", "c2-1"); err != nil {
		t.Fatalf("AddNote failed: %v", err)
	}

	count1, _ := store.CountNotes("#chan1")
	count2, _ := store.CountNotes("#chan2")

	if count1 != 2 {
		t.Errorf("chan1: expected 2 notes, got %d", count1)
	}
	if count2 != 1 {
		t.Errorf("chan2: expected 1 note, got %d", count2)
	}
}

func TestListNotes_OrderByIDAsc(t *testing.T) {
	store := newTestStore(t, 15)

	texts := []string{"first", "second", "third"}
	for _, txt := range texts {
		if err := store.AddNote("#order", "user", txt); err != nil {
			t.Fatalf("AddNote failed: %v", err)
		}
	}

	notes, err := store.ListNotes("#order")
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}

	for i, txt := range texts {
		if notes[i].Text != txt {
			t.Errorf("notes[%d].Text = %q, want %q", i, notes[i].Text, txt)
		}
	}

	// IDs should be ascending
	for i := 1; i < len(notes); i++ {
		if notes[i].ID <= notes[i-1].ID {
			t.Errorf("notes[%d].ID (%d) <= notes[%d].ID (%d)", i, notes[i].ID, i-1, notes[i-1].ID)
		}
	}
}
