package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Daniel-A-C/kingfisher/internal/board"
)

func TestRegistryPersistenceAndScan(t *testing.T) {
	tmp := t.TempDir()
	regPath := filepath.Join(tmp, "config", "registry.json")

	scanRoot := filepath.Join(tmp, "programs")
	projA := filepath.Join(scanRoot, "Corvidae")
	projB := filepath.Join(scanRoot, "Rookery")
	if err := os.MkdirAll(projA, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projB, 0o755); err != nil {
		t.Fatal(err)
	}

	if _, _, err := board.InitProject(projA, "Corvidae", "CRV", true); err != nil {
		t.Fatal(err)
	}
	if _, _, err := board.InitProject(projB, "Rookery", "RKR", true); err != nil {
		t.Fatal(err)
	}

	r, err := Load(regPath)
	if err != nil {
		t.Fatal(err)
	}
	r.Preferences.ScanDirs = []string{scanRoot}

	found := r.ScanProjects()
	if len(found) != 2 {
		t.Fatalf("expected 2 discovered projects, got %d", len(found))
	}

	r.OpenTab(projA)
	r.OpenTab(projB)
	if r.ActiveTab != projB {
		t.Fatalf("expected active tab %s, got %s", projB, r.ActiveTab)
	}

	if err := r.Save(); err != nil {
		t.Fatal(err)
	}

	reloaded, err := Load(regPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.Projects) != 2 || len(reloaded.OpenTabs) != 2 {
		t.Fatalf("unexpected reloaded registry: %+v", reloaded)
	}

	p, ok := reloaded.FindBySlugOrPath("corvidae")
	if !ok || p.Prefix != "CRV" {
		t.Fatalf("expected to find corvidae by slug, got %+v (ok=%v)", p, ok)
	}

	reloaded.CloseTab(projB)
	if reloaded.ActiveTab != projA {
		t.Fatalf("expected active tab fallback to %s, got %s", projA, reloaded.ActiveTab)
	}
}
