package projectinfo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Daniel-A-C/kingfisher/internal/board"
	"github.com/Daniel-A-C/kingfisher/internal/registry"
)

func TestCollectGitStatsAndMakeup(t *testing.T) {
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Tester",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=Tester",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
		}
	}

	run("init", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() { println(\"hello kingfisher\") }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "App.svelte"), []byte("<script>let x = 1;</script><h1>Hi</h1>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(`{"lockfileVersion": 3}`), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "Initial Kingfisher commit")

	// Add an untracked file to test dirty detection
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("wip"), 0o644); err != nil {
		t.Fatal(err)
	}

	stats := CollectGitStats(context.Background(), dir)
	if !stats.IsGitRepo {
		t.Fatalf("expected IsGitRepo=true")
	}
	if stats.Branch != "main" {
		t.Fatalf("expected branch main, got %q", stats.Branch)
	}
	if stats.CommitCount != 1 {
		t.Fatalf("expected 1 commit, got %d", stats.CommitCount)
	}
	if stats.DirtyCount != 1 {
		t.Fatalf("expected 1 dirty file, got %d", stats.DirtyCount)
	}
	if stats.LastCommitMsg != "Initial Kingfisher commit" {
		t.Fatalf("unexpected last commit msg: %q", stats.LastCommitMsg)
	}
	if stats.TrackedFiles != 3 {
		t.Fatalf("expected 3 tracked files, got %d", stats.TrackedFiles)
	}
	if stats.TotalBytes == 0 {
		t.Fatalf("expected non-zero TotalBytes")
	}
	if len(stats.Languages) != 2 {
		t.Fatalf("expected 2 languages (Go + Svelte, lockfile excluded), got %+v", stats.Languages)
	}
	if stats.Languages[0].Name != "Go" || stats.Languages[1].Name != "Svelte" {
		t.Fatalf("unexpected language order: %+v", stats.Languages)
	}
}

func TestCollectGitStatsNonRepo(t *testing.T) {
	dir := t.TempDir()
	stats := CollectGitStats(context.Background(), dir)
	if stats.IsGitRepo {
		t.Fatalf("expected IsGitRepo=false for plain temp dir")
	}
}

func TestCollectAllAndBoardSummary(t *testing.T) {
	dir := t.TempDir()
	b, _, err := board.InitProject(dir, "Atlas", "ATL", false)
	if err != nil {
		t.Fatal(err)
	}
	b.AddTask("Low ready task", "ready", "low", "", nil)
	b.AddTask("High ready task", "ready", "high", "", nil)
	b.AddTask("Active work", "active", "medium", "", nil)
	b.AddTask("Finished work", "done", "critical", "", nil)
	if _, err := board.SaveAtomic(dir, b); err != nil {
		t.Fatal(err)
	}

	overviews := CollectAll(
		context.Background(),
		[]registry.ProjectEntry{{Slug: "atlas", Name: "Atlas", Prefix: "ATL", Path: dir}},
		[]string{dir},
		dir,
		map[string]*board.Board{dir: b},
	)
	if len(overviews) != 1 {
		t.Fatalf("expected 1 overview, got %d", len(overviews))
	}
	ov := overviews[0]
	if !ov.IsOpen || !ov.IsActive {
		t.Fatalf("expected IsOpen & IsActive true")
	}
	if ov.Tasks.Total != 4 || ov.Tasks.Ready != 2 || ov.Tasks.Active != 1 || ov.Tasks.Done != 1 {
		t.Fatalf("unexpected task summary: %+v", ov.Tasks)
	}
	if ov.Tasks.HighPriorityOpen != 1 {
		t.Fatalf("expected 1 high-priority open task, got %d", ov.Tasks.HighPriorityOpen)
	}
	if ov.Tasks.NextTaskTitle != "High ready task" {
		t.Fatalf("expected next task 'High ready task', got %q", ov.Tasks.NextTaskTitle)
	}
}
