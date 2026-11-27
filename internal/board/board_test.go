package board

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDerivePrefix(t *testing.T) {
	cases := map[string]string{
		"Corvidae":        "CRV",
		"Rookery":         "RKR",
		"Kingfisher":      "KNG",
		"my-cool-app":     "MCA",
		"MusicVisualizer": "MV",
		"BIS":             "BIS",
	}
	for input, want := range cases {
		if got := DerivePrefix(input); got != want {
			t.Errorf("DerivePrefix(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestInitAndGitExclude(t *testing.T) {
	dir := t.TempDir()
	// Simulate a git repository
	if err := os.MkdirAll(filepath.Join(dir, ".git", "info"), 0o755); err != nil {
		t.Fatal(err)
	}

	b, ignoredPath, err := InitProject(dir, "Corvidae", "CRV", false)
	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}
	if b.Project.Name != "Corvidae" || b.Project.Prefix != "CRV" || len(b.Columns) != 4 {
		t.Fatalf("unexpected board: %+v", b)
	}
	if !strings.HasSuffix(ignoredPath, filepath.Join(".git", "info", "exclude")) {
		t.Fatalf("expected .git/info/exclude, got %s", ignoredPath)
	}

	// Ensure .gitignore was NOT created
	if _, err := os.Stat(filepath.Join(dir, ".gitignore")); !os.IsNotExist(err) {
		t.Fatalf("expected .gitignore to remain untouched")
	}

	// Idempotency check
	if _, err := EnsureIgnored(dir, false); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(ignoredPath)
	if err != nil {
		t.Fatal(err)
	}
	if count := strings.Count(string(content), BoardFilename); count != 1 {
		t.Fatalf("expected %s once in exclude file, got %d", BoardFilename, count)
	}
}

func TestCRUDAndAtomicRoundTrip(t *testing.T) {
	dir := t.TempDir()
	b, _, err := InitProject(dir, "Kingfisher", "KF", true)
	if err != nil {
		t.Fatal(err)
	}

	t1 := b.AddTask("Low priority item", "ready", "low", "desc 1", []string{"ux"})
	t2 := b.AddTask("Critical dive", "ready", "critical", "desc 2", []string{"core", "cli"})
	if t1.ID != "KF-1" || t2.ID != "KF-2" {
		t.Fatalf("unexpected IDs: %s, %s", t1.ID, t2.ID)
	}

	next, ok := b.NextReadyTask()
	if !ok || next.ID != "KF-2" {
		t.Fatalf("expected KF-2 as next ready task, got %+v (ok=%v)", next, ok)
	}

	moved, err := b.MoveTask("KF-2", "active", 0, "Started work on core parser")
	if err != nil {
		t.Fatal(err)
	}
	if moved.Status != "active" || moved.AgentNotes != "Started work on core parser" {
		t.Fatalf("unexpected moved task: %+v", moved)
	}

	hash1, err := SaveAtomic(dir, b)
	if err != nil {
		t.Fatal(err)
	}

	loaded, hash2, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if hash1 != hash2 {
		t.Fatalf("hash mismatch: %s != %s", hash1, hash2)
	}
	if len(loaded.Tasks) != 2 || loaded.Project.NextID != 3 {
		t.Fatalf("unexpected loaded board: %+v", loaded)
	}
}
