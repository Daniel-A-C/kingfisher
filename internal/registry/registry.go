package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Daniel-A-C/kingfisher/internal/board"
	"github.com/Daniel-A-C/kingfisher/internal/capture"
)

// ProjectEntry is a lightweight pointer to a local repo containing .kingfisher.yaml.
type ProjectEntry struct {
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	Prefix     string `json:"prefix"`
	Path       string `json:"path"`
	LastOpened string `json:"last_opened"`
	Pinned     bool   `json:"pinned"`
}

// Preferences holds global UI, accessibility, auto-discovery, and local AI capture settings.
type Preferences struct {
	Opacity           float64        `json:"opacity"`
	Blur              float64        `json:"blur"`
	HighContrast      bool           `json:"high_contrast"`
	ReduceMotion      bool           `json:"reduce_motion"`
	ShowHeaderOpacity *bool          `json:"show_header_opacity,omitempty"`
	ScanDirs          []string       `json:"scan_dirs"`
	AI                capture.Config `json:"ai"`
}

// Registry stores known project paths, open tab order, and user preferences.
type Registry struct {
	mu          sync.Mutex
	filePath    string
	Projects    []ProjectEntry `json:"projects"`
	OpenTabs    []string       `json:"open_tabs"`
	ActiveTab   string         `json:"active_tab"`
	Preferences Preferences    `json:"preferences"`
}

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

// DefaultPath returns ~/.config/kingfisher/registry.json (or KINGFISHER_CONFIG_DIR/registry.json).
func DefaultPath() (string, error) {
	if custom := strings.TrimSpace(os.Getenv("KINGFISHER_CONFIG_DIR")); custom != "" {
		return filepath.Join(custom, "registry.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "kingfisher", "registry.json"), nil
}

func boolPtr(v bool) *bool { return &v }

// DefaultPreferences returns initial Kingfisher defaults.
func DefaultPreferences() Preferences {
	return Preferences{
		Opacity:           0.35,
		Blur:              0.25,
		HighContrast:      false,
		ReduceMotion:      false,
		ShowHeaderOpacity: boolPtr(false),
		ScanDirs:          []string{"~/Documents/programs"},
		AI:                capture.DefaultConfig(),
	}
}

// Load reads the registry from path, or creates a default in-memory registry if it does not exist.
func Load(path string) (*Registry, error) {
	r := &Registry{
		filePath:    path,
		Projects:    []ProjectEntry{},
		OpenTabs:    []string{},
		Preferences: DefaultPreferences(),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return r, nil
		}
		return nil, fmt.Errorf("read registry: %w", err)
	}

	if err := json.Unmarshal(data, r); err != nil {
		return nil, fmt.Errorf("parse registry: %w", err)
	}
	if r.Projects == nil {
		r.Projects = []ProjectEntry{}
	}
	if r.OpenTabs == nil {
		r.OpenTabs = []string{}
	}
	if r.Preferences.ScanDirs == nil {
		r.Preferences.ScanDirs = DefaultPreferences().ScanDirs
	}
	if r.Preferences.AI.Endpoint == "" && r.Preferences.AI.Model == "" && !r.Preferences.AI.Enabled {
		r.Preferences.AI = capture.DefaultConfig()
	}
	if r.Preferences.ShowHeaderOpacity == nil {
		r.Preferences.ShowHeaderOpacity = boolPtr(true)
	}
	if r.Preferences.Opacity < 0 || r.Preferences.Opacity > 1.0 {
		r.Preferences.Opacity = 0.35
	}
	if r.Preferences.Blur < 0 || r.Preferences.Blur > 1.0 {
		r.Preferences.Blur = 0.25
	}
	return r, nil
}

// Save atomically persists the registry JSON to disk.
func (r *Registry) Save() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.saveLocked()
}

func (r *Registry) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(r.filePath), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}
	tmp := r.filePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write temp registry: %w", err)
	}
	if err := os.Rename(tmp, r.filePath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename temp registry: %w", err)
	}
	return nil
}

// Slugify converts a project name into a clean lowercase URL/CLI slug.
func Slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugSanitizer.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "project"
	}
	return s
}

// RegisterProject adds or updates a project pointer in the registry.
func (r *Registry) RegisterProject(projectDir, name, prefix string) ProjectEntry {
	r.mu.Lock()
	defer r.mu.Unlock()

	abs, err := filepath.Abs(projectDir)
	if err != nil {
		abs = filepath.Clean(projectDir)
	}
	now := time.Now().UTC().Format(time.RFC3339)

	for i := range r.Projects {
		if r.Projects[i].Path == abs {
			if name != "" {
				r.Projects[i].Name = name
			}
			if prefix != "" {
				r.Projects[i].Prefix = prefix
			}
			r.Projects[i].LastOpened = now
			return r.Projects[i]
		}
	}

	if name == "" {
		name = filepath.Base(abs)
	}
	if prefix == "" {
		prefix = board.DerivePrefix(name)
	}

	baseSlug := Slugify(name)
	slug := baseSlug
	suffix := 2
	for r.slugExistsLocked(slug) {
		slug = fmt.Sprintf("%s-%d", baseSlug, suffix)
		suffix++
	}

	entry := ProjectEntry{
		Slug:       slug,
		Name:       name,
		Prefix:     prefix,
		Path:       abs,
		LastOpened: now,
	}
	r.Projects = append(r.Projects, entry)
	return entry
}

func (r *Registry) slugExistsLocked(slug string) bool {
	for _, p := range r.Projects {
		if p.Slug == slug {
			return true
		}
	}
	return false
}

// FindBySlugOrPath looks up a registered project by slug, exact path, or case-insensitive name.
func (r *Registry) FindBySlugOrPath(query string) (*ProjectEntry, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	q := strings.TrimSpace(query)
	if q == "" {
		return nil, false
	}
	abs, _ := filepath.Abs(ExpandHome(q))

	for i := range r.Projects {
		p := &r.Projects[i]
		if strings.EqualFold(p.Slug, q) || p.Path == abs || strings.EqualFold(p.Name, q) {
			cp := *p
			return &cp, true
		}
	}
	return nil, false
}

// OpenTab ensures path is in OpenTabs and sets it as ActiveTab.
func (r *Registry) OpenTab(projectDir string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	abs, err := filepath.Abs(projectDir)
	if err != nil {
		abs = filepath.Clean(projectDir)
	}

	found := false
	for _, t := range r.OpenTabs {
		if t == abs {
			found = true
			break
		}
	}
	if !found {
		r.OpenTabs = append(r.OpenTabs, abs)
	}
	r.ActiveTab = abs
}

// CloseTab removes path from OpenTabs and updates ActiveTab sensibly.
func (r *Registry) CloseTab(projectDir string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	abs, err := filepath.Abs(projectDir)
	if err != nil {
		abs = filepath.Clean(projectDir)
	}

	idx := -1
	for i, t := range r.OpenTabs {
		if t == abs {
			idx = i
			break
		}
	}
	if idx == -1 {
		return
	}

	r.OpenTabs = append(r.OpenTabs[:idx], r.OpenTabs[idx+1:]...)
	if r.ActiveTab == abs {
		if len(r.OpenTabs) == 0 {
			r.ActiveTab = ""
		} else if idx >= len(r.OpenTabs) {
			r.ActiveTab = r.OpenTabs[len(r.OpenTabs)-1]
		} else {
			r.ActiveTab = r.OpenTabs[idx]
		}
	}
}

// SetActiveTab updates ActiveTab if it is currently open.
func (r *Registry) SetActiveTab(projectDir string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	abs, err := filepath.Abs(projectDir)
	if err != nil {
		abs = filepath.Clean(projectDir)
	}
	for _, t := range r.OpenTabs {
		if t == abs {
			r.ActiveTab = abs
			return
		}
	}
}

// ReorderTabs replaces OpenTabs with a validated order.
func (r *Registry) ReorderTabs(paths []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing := make(map[string]bool, len(r.OpenTabs))
	for _, t := range r.OpenTabs {
		existing[t] = true
	}

	next := make([]string, 0, len(r.OpenTabs))
	seen := make(map[string]bool, len(paths))
	for _, p := range paths {
		if existing[p] && !seen[p] {
			next = append(next, p)
			seen[p] = true
		}
	}
	for _, t := range r.OpenTabs {
		if !seen[t] {
			next = append(next, t)
		}
	}
	r.OpenTabs = next
}

// ScanProjects walks configured ScanDirs (up to depth 2) and registers any folder containing .kingfisher.yaml.
func (r *Registry) ScanProjects() []ProjectEntry {
	r.mu.Lock()
	dirs := append([]string(nil), r.Preferences.ScanDirs...)
	r.mu.Unlock()

	skipDirs := map[string]bool{
		".git": true, "node_modules": true, "dist": true, "build": true,
		"target": true, ".venv": true, "vendor": true, ".next": true,
	}

	var discovered []string
	for _, rawDir := range dirs {
		root := ExpandHome(rawDir)
		info, err := os.Stat(root)
		if err != nil || !info.IsDir() {
			continue
		}

		// Depth 0
		if board.Exists(root) {
			discovered = append(discovered, root)
		}

		// Depth 1 & 2
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, e1 := range entries {
			if !e1.IsDir() || skipDirs[e1.Name()] || strings.HasPrefix(e1.Name(), ".") {
				continue
			}
			p1 := filepath.Join(root, e1.Name())
			if board.Exists(p1) {
				discovered = append(discovered, p1)
				continue
			}
			subEntries, err := os.ReadDir(p1)
			if err != nil {
				continue
			}
			for _, e2 := range subEntries {
				if !e2.IsDir() || skipDirs[e2.Name()] || strings.HasPrefix(e2.Name(), ".") {
					continue
				}
				p2 := filepath.Join(p1, e2.Name())
				if board.Exists(p2) {
					discovered = append(discovered, p2)
				}
			}
		}
	}

	var added []ProjectEntry
	for _, dir := range discovered {
		b, _, err := board.Load(dir)
		if err != nil {
			continue
		}
		entry := r.RegisterProject(dir, b.Project.Name, b.Project.Prefix)
		added = append(added, entry)
	}
	return added
}

// ExpandHome replaces a leading ~/ with the user's home directory.
func ExpandHome(p string) string {
	if strings.HasPrefix(p, "~/") || p == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}
