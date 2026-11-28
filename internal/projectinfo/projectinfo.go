package projectinfo

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Daniel-A-C/kingfisher/internal/board"
	"github.com/Daniel-A-C/kingfisher/internal/registry"
)

// LanguageStat represents a single language's share of a project's codebase.
type LanguageStat struct {
	Name    string  `json:"name"`
	Color   string  `json:"color"`
	Bytes   int64   `json:"bytes"`
	Files   int     `json:"files"`
	Percent float64 `json:"percent"`
}

// GitStats holds Git repository size, activity, and language makeup metrics.
type GitStats struct {
	IsGitRepo     bool           `json:"is_git_repo"`
	Branch        string         `json:"branch"`
	DirtyCount    int            `json:"dirty_count"`
	CommitCount   int            `json:"commit_count"`
	LastCommitAt  string         `json:"last_commit_at"`
	LastCommitMsg string         `json:"last_commit_msg"`
	TrackedFiles  int            `json:"tracked_files"`
	TotalBytes    int64          `json:"total_bytes"`
	Languages     []LanguageStat `json:"languages"`
}

// TaskSummary summarizes a project's Kingfisher board task counts.
type TaskSummary struct {
	Total             int    `json:"total"`
	Backlog           int    `json:"backlog"`
	Ready             int    `json:"ready"`
	Active            int    `json:"active"`
	Done              int    `json:"done"`
	HighPriorityOpen  int    `json:"high_priority_open"`
	NextTaskID        string `json:"next_task_id,omitempty"`
	NextTaskTitle     string `json:"next_task_title,omitempty"`
}

// Overview is the complete Home Menu summary card payload for a single project.
type Overview struct {
	Slug       string      `json:"slug"`
	Name       string      `json:"name"`
	Prefix     string      `json:"prefix"`
	Path       string      `json:"path"`
	LastOpened string      `json:"last_opened"`
	IsOpen     bool        `json:"is_open"`
	IsActive   bool        `json:"is_active"`
	Tasks      TaskSummary `json:"tasks"`
	Git        GitStats    `json:"git"`
}

type langMeta struct {
	name  string
	color string
}

// Jewel-toned Art Nouveau harmonious palette for language makeup bars.
var extLanguages = map[string]langMeta{
	".go":     {name: "Go", color: "#46d9ec"},
	".svelte": {name: "Svelte", color: "#e65d38"},
	".ts":     {name: "TypeScript", color: "#569cd6"},
	".tsx":    {name: "TypeScript", color: "#569cd6"},
	".mts":    {name: "TypeScript", color: "#569cd6"},
	".js":     {name: "JavaScript", color: "#f6d388"},
	".jsx":    {name: "JavaScript", color: "#f6d388"},
	".mjs":    {name: "JavaScript", color: "#f6d388"},
	".cjs":    {name: "JavaScript", color: "#f6d388"},
	".py":     {name: "Python", color: "#6bb594"},
	".rs":     {name: "Rust", color: "#c46d34"},
	".swift":  {name: "Swift", color: "#eb7b46"},
	".rb":     {name: "Ruby", color: "#d65a6f"},
	".java":   {name: "Java", color: "#dca248"},
	".kt":     {name: "Kotlin", color: "#9f79c8"},
	".kts":    {name: "Kotlin", color: "#9f79c8"},
	".c":      {name: "C/C++", color: "#7ca5c8"},
	".h":      {name: "C/C++", color: "#7ca5c8"},
	".cc":     {name: "C/C++", color: "#7ca5c8"},
	".cpp":    {name: "C/C++", color: "#7ca5c8"},
	".hpp":    {name: "C/C++", color: "#7ca5c8"},
	".cs":     {name: "C#", color: "#47b881"},
	".php":    {name: "PHP", color: "#8c7ae6"},
	".css":    {name: "CSS", color: "#9f79c8"},
	".scss":   {name: "CSS", color: "#9f79c8"},
	".less":   {name: "CSS", color: "#9f79c8"},
	".html":   {name: "HTML", color: "#d97757"},
	".htm":    {name: "HTML", color: "#d97757"},
	".vue":    {name: "Vue", color: "#42b883"},
	".sh":     {name: "Shell", color: "#2f9e7a"},
	".zsh":    {name: "Shell", color: "#2f9e7a"},
	".bash":   {name: "Shell", color: "#2f9e7a"},
	".sql":    {name: "SQL", color: "#3b9ab2"},
	".yaml":   {name: "YAML", color: "#b89f6d"},
	".yml":    {name: "YAML", color: "#b89f6d"},
	".toml":   {name: "TOML", color: "#b89f6d"},
	".json":   {name: "JSON", color: "#9db4ab"},
	".md":     {name: "Markdown", color: "#8aa79f"},
	".mdx":    {name: "Markdown", color: "#8aa79f"},
	".zig":    {name: "Zig", color: "#e5a93c"},
	".ex":     {name: "Elixir", color: "#a57bd6"},
	".exs":    {name: "Elixir", color: "#a57bd6"},
	".lua":    {name: "Lua", color: "#5c85d6"},
	".dart":   {name: "Dart", color: "#29b6f6"},
	".r":      {name: "R", color: "#5b8fd9"},
}

var skipMakeupFiles = map[string]bool{
	"package-lock.json": true,
	"yarn.lock":         true,
	"pnpm-lock.yaml":    true,
	"go.sum":            true,
	"Cargo.lock":        true,
	"poetry.lock":       true,
	"composer.lock":     true,
	"Gemfile.lock":      true,
}

var skipMakeupExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".ico": true, ".icns": true, ".svg": true, ".pdf": true, ".zip": true,
	".tar": true, ".gz": true, ".woff": true, ".woff2": true, ".ttf": true,
	".otf": true, ".eot": true, ".mp3": true, ".mp4": true, ".mov": true,
	".wav": true, ".wasm": true, ".exe": true, ".dll": true, ".so": true,
	".dylib": true, ".bin": true, ".lock": true,
}

// CollectAll gathers Overviews for all registered projects concurrently.
func CollectAll(
	ctx context.Context,
	projects []registry.ProjectEntry,
	openTabs []string,
	activeTab string,
	loadedBoards map[string]*board.Board,
) []Overview {
	if len(projects) == 0 {
		return []Overview{}
	}

	openSet := make(map[string]bool, len(openTabs))
	for _, t := range openTabs {
		openSet[t] = true
	}

	results := make([]Overview, len(projects))
	sem := make(chan struct{}, 6)
	var wg sync.WaitGroup

	for i, proj := range projects {
		wg.Add(1)
		go func(idx int, p registry.ProjectEntry) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			var b *board.Board
			if cached, ok := loadedBoards[p.Path]; ok && cached != nil {
				b = cached
			} else if loaded, _, err := board.Load(p.Path); err == nil {
				b = loaded
			}

			results[idx] = Overview{
				Slug:       p.Slug,
				Name:       p.Name,
				Prefix:     p.Prefix,
				Path:       p.Path,
				LastOpened: p.LastOpened,
				IsOpen:     openSet[p.Path],
				IsActive:   p.Path == activeTab,
				Tasks:      SummarizeBoard(b),
				Git:        CollectGitStats(ctx, p.Path),
			}
		}(i, proj)
	}

	wg.Wait()
	return results
}

// SummarizeBoard computes task counts and the highest-priority ready task for a board.
func SummarizeBoard(b *board.Board) TaskSummary {
	if b == nil {
		return TaskSummary{}
	}
	var s TaskSummary
	s.Total = len(b.Tasks)

	for _, t := range b.Tasks {
		switch t.Status {
		case "backlog":
			s.Backlog++
		case "ready":
			s.Ready++
		case "active":
			s.Active++
		case "done":
			s.Done++
		}
		if t.Status != "done" && (t.Priority == board.PriorityHigh || t.Priority == board.PriorityCritical) {
			s.HighPriorityOpen++
		}
	}

	if next, ok := b.NextReadyTask(); ok && next != nil {
		s.NextTaskID = next.ID
		s.NextTaskTitle = next.Title
	}
	return s
}

// CollectGitStats inspects a directory's Git metadata, tracked file size, and language makeup.
func CollectGitStats(ctx context.Context, projectDir string) GitStats {
	stats := GitStats{
		Languages: []LanguageStat{},
	}

	if info, err := os.Stat(projectDir); err != nil || !info.IsDir() {
		return stats
	}

	// Verify this directory is inside a git working tree
	inside, err := runGit(ctx, projectDir, "rev-parse", "--is-inside-work-tree")
	if err != nil || strings.TrimSpace(inside) != "true" {
		return stats
	}
	stats.IsGitRepo = true

	// Branch name (falls back to short commit hash on detached HEAD)
	if branch, err := runGit(ctx, projectDir, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		b := strings.TrimSpace(branch)
		if b == "HEAD" {
			if sha, shaErr := runGit(ctx, projectDir, "rev-parse", "--short", "HEAD"); shaErr == nil {
				b = strings.TrimSpace(sha)
			}
		}
		stats.Branch = b
	}

	// Dirty working tree count
	if porcelain, err := runGit(ctx, projectDir, "status", "--porcelain"); err == nil {
		trimmed := strings.TrimRight(porcelain, "\r\n")
		if trimmed != "" {
			stats.DirtyCount = len(strings.Split(trimmed, "\n"))
		}
	}

	// Commit count
	if countStr, err := runGit(ctx, projectDir, "rev-list", "--count", "HEAD"); err == nil {
		if n, parseErr := strconv.Atoi(strings.TrimSpace(countStr)); parseErr == nil {
			stats.CommitCount = n
		}
	}

	// Last commit timestamp & subject
	if logLine, err := runGit(ctx, projectDir, "log", "-1", "--format=%cI%x1f%s"); err == nil {
		parts := strings.SplitN(strings.TrimSpace(logLine), "\x1f", 2)
		if len(parts) >= 1 {
			stats.LastCommitAt = parts[0]
		}
		if len(parts) == 2 {
			stats.LastCommitMsg = parts[1]
		}
	}

	// Tracked files -> total bytes & language makeup
	if lsOut, err := runGitRaw(ctx, projectDir, "ls-files", "-z"); err == nil && len(lsOut) > 0 {
		entries := bytes.Split(lsOut, []byte{0})
		stats.TrackedFiles, stats.TotalBytes, stats.Languages = computeFileMakeup(projectDir, entries)
	}

	return stats
}

func computeFileMakeup(projectDir string, relPaths [][]byte) (int, int64, []LanguageStat) {
	type bucket struct {
		color string
		bytes int64
		files int
	}
	byLang := make(map[string]*bucket)
	var trackedCount int
	var totalBytes int64
	var makeupBytes int64

	for _, raw := range relPaths {
		if len(raw) == 0 {
			continue
		}
		rel := string(raw)
		full := filepath.Join(projectDir, rel)
		fi, err := os.Stat(full)
		if err != nil || fi.IsDir() {
			continue
		}
		sz := fi.Size()
		trackedCount++
		totalBytes += sz

		base := filepath.Base(rel)
		if skipMakeupFiles[base] {
			continue
		}
		ext := strings.ToLower(filepath.Ext(base))
		if skipMakeupExts[ext] {
			continue
		}

		meta, ok := classifyFile(base, ext)
		if !ok {
			continue
		}
		b, exists := byLang[meta.name]
		if !exists {
			b = &bucket{color: meta.color}
			byLang[meta.name] = b
		}
		b.bytes += sz
		b.files++
		makeupBytes += sz
	}

	if makeupBytes == 0 || len(byLang) == 0 {
		return trackedCount, totalBytes, []LanguageStat{}
	}

	all := make([]LanguageStat, 0, len(byLang))
	for name, b := range byLang {
		pct := float64(b.bytes) * 100.0 / float64(makeupBytes)
		all = append(all, LanguageStat{
			Name:    name,
			Color:   b.color,
			Bytes:   b.bytes,
			Files:   b.files,
			Percent: roundOneDecimal(pct),
		})
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].Bytes != all[j].Bytes {
			return all[i].Bytes > all[j].Bytes
		}
		return all[i].Name < all[j].Name
	})

	const maxSegments = 5
	if len(all) > maxSegments {
		top := all[:maxSegments]
		var otherBytes int64
		var otherFiles int
		for _, rest := range all[maxSegments:] {
			otherBytes += rest.Bytes
			otherFiles += rest.Files
		}
		if otherBytes > 0 {
			top = append(top, LanguageStat{
				Name:    "Other",
				Color:   "#6b847e",
				Bytes:   otherBytes,
				Files:   otherFiles,
				Percent: roundOneDecimal(float64(otherBytes) * 100.0 / float64(makeupBytes)),
			})
		}
		all = top
	}

	return trackedCount, totalBytes, all
}

func classifyFile(base, ext string) (langMeta, bool) {
	if meta, ok := extLanguages[ext]; ok {
		return meta, true
	}
	switch strings.ToLower(base) {
	case "makefile", "gnumakefile":
		return langMeta{name: "Makefile", color: "#7ca59a"}, true
	case "dockerfile":
		return langMeta{name: "Docker", color: "#3b9ab2"}, true
	}
	return langMeta{}, false
}

func roundOneDecimal(v float64) float64 {
	return float64(int(v*10+0.5)) / 10.0
}

func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	raw, err := runGitRaw(ctx, dir, args...)
	return string(raw), err
}

func runGitRaw(ctx context.Context, dir string, args ...string) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cmdCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "git", args...)
	cmd.Dir = dir
	return cmd.Output()
}
