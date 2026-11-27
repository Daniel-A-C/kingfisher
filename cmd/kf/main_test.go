package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Daniel-A-C/kingfisher/internal/board"
)

func TestCLIWorkflow(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KINGFISHER_CONFIG_DIR", t.TempDir())

	var out, errBuf bytes.Buffer
	if code := run([]string{"init", "--name", "Corvidae", "--prefix", "CRV", "--gitignore"}, &out, &errBuf); code != 0 {
		t.Fatalf("init failed (%d): %s", code, errBuf.String())
	}

	out.Reset()
	errBuf.Reset()
	if code := run([]string{"add", "--priority", "high", "--tags", "cli,ux", "Add fuzzy deck picker"}, &out, &errBuf); code != 0 {
		t.Fatalf("add failed (%d): %s", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "CRV-1") {
		t.Fatalf("expected CRV-1 in output, got: %s", out.String())
	}

	out.Reset()
	errBuf.Reset()
	if code := run([]string{"next", "--json"}, &out, &errBuf); code != 0 {
		t.Fatalf("next failed (%d): %s", code, errBuf.String())
	}
	var task board.Task
	if err := json.Unmarshal(out.Bytes(), &task); err != nil {
		t.Fatalf("invalid json from next: %v (%s)", err, out.String())
	}
	if task.ID != "CRV-1" || task.Priority != "high" {
		t.Fatalf("unexpected next task: %+v", task)
	}

	out.Reset()
	errBuf.Reset()
	if code := run([]string{"move", "CRV-1", "done", "--note", "Implemented in picker.go"}, &out, &errBuf); code != 0 {
		t.Fatalf("move failed (%d): %s", code, errBuf.String())
	}

	loaded, _, err := board.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Tasks[0].Status != "done" || !strings.Contains(loaded.Tasks[0].AgentNotes, "Implemented in picker.go") {
		t.Fatalf("unexpected task after move: %+v", loaded.Tasks[0])
	}
}
