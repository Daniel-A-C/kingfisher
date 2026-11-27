package capture

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Daniel-A-C/kingfisher/internal/board"
)

func TestHeuristicParse(t *testing.T) {
	b := &board.Board{
		Project: board.ProjectMeta{Name: "Kingfisher", Prefix: "KNG", NextID: 1},
		Columns: board.DefaultColumns(),
	}

	res := HeuristicParse("urgent: fix drag-and-drop glitch on empty columns #board #ux. Cards bounce back when dropped into still waters.", b, "ready")
	if res.UsedAI {
		t.Fatalf("expected UsedAI=false for heuristic parse")
	}
	if res.Priority != board.PriorityCritical {
		t.Fatalf("expected critical priority, got %s", res.Priority)
	}
	if res.Status != "ready" {
		t.Fatalf("expected ready status, got %s", res.Status)
	}
	if res.Title == "" || len(res.Tags) == 0 {
		t.Fatalf("expected non-empty title and tags, got %+v", res)
	}
}

func TestOllamaStructuredCapture(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"models": []map[string]string{{"name": "qwen2.5:1.5b"}},
			})
		case "/api/chat":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": map[string]string{
					"content": `{"title":"Simplify capture with local LLM","status":"ready","priority":"high","tags":["ux","ai","capture"],"description":"Single input box processed by a small local open-source model."}`,
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	b := &board.Board{
		Project: board.ProjectMeta{Name: "Kingfisher", Prefix: "KNG", NextID: 1},
		Columns: board.DefaultColumns(),
	}
	cfg := Config{
		Enabled:        true,
		Provider:       "ollama",
		Endpoint:       srv.URL,
		Model:          "qwen2.5:1.5b",
		TimeoutSeconds: 5,
	}

	st := CheckStatus(context.Background(), cfg)
	if !st.Available || st.Model != "qwen2.5:1.5b" {
		t.Fatalf("unexpected status: %+v", st)
	}

	res := Process(context.Background(), cfg, "Capture should be simplified with a tiny local model", b, "backlog")
	if !res.UsedAI {
		t.Fatalf("expected UsedAI=true, got %+v", res)
	}
	if res.Title != "Simplify capture with local LLM" || res.Status != "ready" || res.Priority != "high" {
		t.Fatalf("unexpected structured result: %+v", res)
	}
}

func TestBatchMultiProjectCapture(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"models": []map[string]string{{"name": "qwen2.5:1.5b"}},
			})
		case "/api/chat":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": map[string]string{
					"content": `{"items":[{"project":"rookery","title":"Fix wallpaper transition on wake","status":"ready","priority":"high","tags":["mac","wallpaper"],"description":"Ensure wallpaper fades smoothly after sleep."},{"project":"kingfisher","title":"Add keyboard navigation to cards","status":"backlog","priority":"medium","tags":["ux"],"description":"Support j/k/h/l navigation across Kanban columns."}]}`,
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	projects := []ProjectContext{
		{Slug: "kingfisher", Name: "Kingfisher", Prefix: "KNG", Columns: board.DefaultColumns()},
		{Slug: "rookery", Name: "Rookery", Prefix: "RKR", Columns: board.DefaultColumns()},
	}
	cfg := Config{Enabled: true, Provider: "ollama", Endpoint: srv.URL, Model: "qwen2.5:1.5b", TimeoutSeconds: 5}

	res := ProcessBatch(context.Background(), cfg, "1. Rookery: fix wallpaper wake bug\n2. Kingfisher: add j/k card nav", projects, "kingfisher", "ready")
	if !res.UsedAI || len(res.Items) != 2 {
		t.Fatalf("expected 2 AI items, got %+v", res)
	}
	if res.Items[0].ProjectSlug != "rookery" || res.Items[1].ProjectSlug != "kingfisher" {
		t.Fatalf("unexpected project routing: %+v", res.Items)
	}

	// Verify heuristic multi-item + project-routing fallback too
	fallback := HeuristicParseBatch("- Rookery: fix wallpaper fade\n- Corvidae: add spaced repetition", []ProjectContext{
		{Slug: "kingfisher", Name: "Kingfisher", Prefix: "KNG", Columns: board.DefaultColumns()},
		{Slug: "rookery", Name: "Rookery", Prefix: "RKR", Columns: board.DefaultColumns()},
		{Slug: "corvidae", Name: "Corvidae", Prefix: "CRV", Columns: board.DefaultColumns()},
	}, "kingfisher", "ready")
	if len(fallback.Items) != 2 {
		t.Fatalf("expected 2 heuristic items, got %+v", fallback)
	}
	if fallback.Items[0].ProjectSlug != "rookery" || fallback.Items[1].ProjectSlug != "corvidae" {
		t.Fatalf("unexpected heuristic routing: %+v", fallback.Items)
	}
}
