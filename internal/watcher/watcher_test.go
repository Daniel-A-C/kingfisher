package watcher

import (
	"os"
	"testing"
	"time"

	"github.com/Daniel-A-C/kingfisher/internal/board"
)

func TestWatcherExternalEditSelfWriteAndRecovery(t *testing.T) {
	dir := t.TempDir()
	b, _, err := board.InitProject(dir, "TestProj", "TST", true)
	if err != nil {
		t.Fatal(err)
	}
	_, initialHash, err := board.Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	updates := make(chan *board.Board, 4)
	errs := make(chan string, 4)

	mgr, err := New(Callbacks{
		OnUpdate: func(_ string, updated *board.Board) { updates <- updated },
		OnError:  func(_ string, msg string) { errs <- msg },
	})
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	if err := mgr.Watch(dir, initialHash); err != nil {
		t.Fatal(err)
	}

	// 1. Self-write should be suppressed
	b.AddTask("Self write task", "ready", "medium", "", nil)
	h, err := board.SaveAtomic(dir, b)
	if err != nil {
		t.Fatal(err)
	}
	mgr.RecordSelfWrite(dir, h)

	select {
	case u := <-updates:
		t.Fatalf("expected self-write to be suppressed, got update with %d tasks", len(u.Tasks))
	case <-time.After(250 * time.Millisecond):
	}

	// 2. External write (agent edit) should fire OnUpdate
	b.AddTask("Agent added task", "active", "high", "", nil)
	if _, err := board.SaveAtomic(dir, b); err != nil {
		t.Fatal(err)
	}

	select {
	case u := <-updates:
		if len(u.Tasks) != 2 {
			t.Fatalf("expected 2 tasks after external edit, got %d", len(u.Tasks))
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for external edit update")
	}

	// 3. Malformed YAML should trigger OnError
	if err := os.WriteFile(board.BoardPath(dir), []byte("project:\n  name: [unclosed"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case msg := <-errs:
		if msg == "" {
			t.Fatalf("expected non-empty error message")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for malformed YAML error")
	}
}
