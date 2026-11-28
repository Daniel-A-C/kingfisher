package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Daniel-A-C/kingfisher/internal/board"
	"github.com/Daniel-A-C/kingfisher/internal/capture"
	"github.com/Daniel-A-C/kingfisher/internal/ipc"
	"github.com/Daniel-A-C/kingfisher/internal/projectinfo"
	"github.com/Daniel-A-C/kingfisher/internal/registry"
	"github.com/Daniel-A-C/kingfisher/internal/watcher"
	"github.com/wailsapp/wails/v2/pkg/options"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// AppState is the complete snapshot pushed to or requested by the Svelte frontend.
type AppState struct {
	Projects    []registry.ProjectEntry `json:"projects"`
	OpenTabs    []string                `json:"open_tabs"`
	ActiveTab   string                  `json:"active_tab"`
	Boards      map[string]*board.Board `json:"boards"`
	Errors      map[string]string       `json:"errors"`
	Preferences registry.Preferences    `json:"preferences"`
}

// BoardUpdateEvent is emitted over Wails events when a watched .kingfisher.yaml changes.
type BoardUpdateEvent struct {
	Path  string       `json:"path"`
	Board *board.Board `json:"board"`
}

// BoardErrorEvent is emitted when an external edit writes malformed YAML.
type BoardErrorEvent struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

// App is the Wails-bound backend controller for Kingfisher.
type App struct {
	ctx       context.Context
	mu        sync.Mutex
	reg       *registry.Registry
	watcher   *watcher.Manager
	ipcServer *ipc.Server
	boards    map[string]*board.Board
	errors    map[string]string
}

// NewApp constructs an App instance.
func NewApp() *App {
	return &App{
		boards: make(map[string]*board.Board),
		errors: make(map[string]string),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	regPath, err := registry.DefaultPath()
	if err == nil {
		if r, loadErr := registry.Load(regPath); loadErr == nil {
			a.reg = r
		}
	}
	if a.reg == nil {
		a.reg, _ = registry.Load(filepath.Join(os.TempDir(), "kingfisher-registry.json"))
	}

	mgr, err := watcher.New(watcher.Callbacks{
		OnUpdate: a.handleExternalBoardUpdate,
		OnError:  a.handleExternalBoardError,
	})
	if err == nil {
		a.watcher = mgr
	}

	// Auto-discover any projects in configured scan directories
	a.reg.ScanProjects()

	// Load and watch all previously open tabs that still exist
	validTabs := make([]string, 0, len(a.reg.OpenTabs))
	for _, tabPath := range a.reg.OpenTabs {
		if !board.Exists(tabPath) {
			continue
		}
		if b, hash, err := board.Load(tabPath); err == nil {
			a.boards[tabPath] = b
			validTabs = append(validTabs, tabPath)
			if a.watcher != nil {
				_ = a.watcher.Watch(tabPath, hash)
			}
		}
	}
	a.reg.OpenTabs = validTabs
	if len(validTabs) > 0 {
		foundActive := false
		for _, t := range validTabs {
			if t == a.reg.ActiveTab {
				foundActive = true
				break
			}
		}
		if !foundActive {
			a.reg.ActiveTab = validTabs[0]
		}
	} else {
		a.reg.ActiveTab = ""
	}
	_ = a.reg.Save()

	// Start single-instance Unix socket IPC server for kf CLI
	if srv, err := ipc.StartServer(ipc.SocketPath(), a.handleIPCMessage); err == nil {
		a.ipcServer = srv
	}

	hideNativeWindowButtons()
}

func (a *App) domReady(_ context.Context) {
	hideNativeWindowButtons()
}

func (a *App) shutdown(_ context.Context) {
	if a.ipcServer != nil {
		_ = a.ipcServer.Close()
	}
	if a.watcher != nil {
		_ = a.watcher.Close()
	}
	if a.reg != nil {
		_ = a.reg.Save()
	}
}

func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	a.focusWindow()
	for _, arg := range secondInstanceData.Args {
		if strings.HasPrefix(arg, "kingfisher://") {
			a.handleDeepLink(arg)
			return
		}
	}
}

func (a *App) handleIPCMessage(msg ipc.Message) {
	a.focusWindow()

	var targetPath string
	if msg.Path != "" {
		targetPath = msg.Path
	} else if msg.Slug != "" {
		if entry, ok := a.reg.FindBySlugOrPath(msg.Slug); ok {
			targetPath = entry.Path
		}
	}

	if targetPath != "" {
		if _, err := a.OpenProjectTab(targetPath); err == nil && a.ctx != nil {
			wailsruntime.EventsEmit(a.ctx, "state:changed", a.GetState())
		}
	}
}

func (a *App) handleDeepLink(rawURL string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	slug := u.Query().Get("project")
	if slug == "" {
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) > 0 && parts[0] != "" {
			slug = parts[len(parts)-1]
		}
	}
	if slug != "" {
		a.handleIPCMessage(ipc.Message{Action: "open", Slug: slug})
	}
}

func (a *App) focusWindow() {
	if a.ctx == nil {
		return
	}
	wailsruntime.WindowUnminimise(a.ctx)
	wailsruntime.WindowShow(a.ctx)
	wailsruntime.WindowSetAlwaysOnTop(a.ctx, true)
	wailsruntime.WindowSetAlwaysOnTop(a.ctx, false)
}

func (a *App) handleExternalBoardUpdate(projectDir string, b *board.Board) {
	a.mu.Lock()
	a.boards[projectDir] = b
	delete(a.errors, projectDir)
	a.reg.RegisterProject(projectDir, b.Project.Name, b.Project.Prefix)
	a.mu.Unlock()

	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "board:updated", BoardUpdateEvent{
			Path:  projectDir,
			Board: b,
		})
	}
}

func (a *App) handleExternalBoardError(projectDir string, errMsg string) {
	a.mu.Lock()
	a.errors[projectDir] = errMsg
	a.mu.Unlock()

	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "board:error", BoardErrorEvent{
			Path:  projectDir,
			Error: errMsg,
		})
	}
}

// GetState returns the current AppState snapshot.
func (a *App) GetState() AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.snapshotLocked()
}

func (a *App) snapshotLocked() AppState {
	boardsCopy := make(map[string]*board.Board, len(a.boards))
	for k, v := range a.boards {
		boardsCopy[k] = v
	}
	errorsCopy := make(map[string]string, len(a.errors))
	for k, v := range a.errors {
		errorsCopy[k] = v
	}
	return AppState{
		Projects:    append([]registry.ProjectEntry(nil), a.reg.Projects...),
		OpenTabs:    append([]string(nil), a.reg.OpenTabs...),
		ActiveTab:   a.reg.ActiveTab,
		Boards:      boardsCopy,
		Errors:      errorsCopy,
		Preferences: a.reg.Preferences,
	}
}

// ScanProjects rescans configured scan directories and returns the refreshed AppState.
func (a *App) ScanProjects() AppState {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.reg.ScanProjects()
	_ = a.reg.Save()
	return a.snapshotLocked()
}

// GetProjectOverviews returns Home Menu summaries (task counts + Git size & language makeup)
// for all registered Kingfisher projects.
func (a *App) GetProjectOverviews() []projectinfo.Overview {
	a.mu.Lock()
	projectsCopy := append([]registry.ProjectEntry(nil), a.reg.Projects...)
	tabsCopy := append([]string(nil), a.reg.OpenTabs...)
	activeTab := a.reg.ActiveTab
	boardsCopy := make(map[string]*board.Board, len(a.boards))
	for k, v := range a.boards {
		boardsCopy[k] = v
	}
	ctx := a.ctx
	a.mu.Unlock()

	if ctx == nil {
		ctx = context.Background()
	}
	return projectinfo.CollectAll(ctx, projectsCopy, tabsCopy, activeTab, boardsCopy)
}

// OpenProjectTab loads a project's .kingfisher.yaml, registers it, starts watching it, and switches to its tab.
func (a *App) OpenProjectTab(projectDir string) (AppState, error) {
	abs, err := filepath.Abs(registry.ExpandHome(projectDir))
	if err != nil {
		return AppState{}, err
	}

	b, hash, err := board.Load(abs)
	if err != nil {
		return AppState{}, err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.boards[abs] = b
	delete(a.errors, abs)
	a.reg.RegisterProject(abs, b.Project.Name, b.Project.Prefix)
	a.reg.OpenTab(abs)
	_ = a.reg.Save()

	if a.watcher != nil {
		_ = a.watcher.Watch(abs, hash)
	}

	return a.snapshotLocked(), nil
}

// CloseProjectTab closes a project tab and stops watching its directory.
func (a *App) CloseProjectTab(projectDir string) AppState {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.reg.CloseTab(projectDir)
	delete(a.boards, projectDir)
	delete(a.errors, projectDir)
	if a.watcher != nil {
		a.watcher.Unwatch(projectDir)
	}
	_ = a.reg.Save()
	return a.snapshotLocked()
}

// SetActiveTab switches the active project tab.
func (a *App) SetActiveTab(projectDir string) AppState {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.reg.SetActiveTab(projectDir)
	_ = a.reg.Save()
	return a.snapshotLocked()
}

// ReorderTabs persists a new tab ordering.
func (a *App) ReorderTabs(paths []string) AppState {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.reg.ReorderTabs(paths)
	_ = a.reg.Save()
	return a.snapshotLocked()
}

// PickFolder opens a native macOS folder selection dialog.
func (a *App) PickFolder() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("app context unavailable")
	}
	return wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Select Project Folder",
	})
}

// InitProject initializes .kingfisher.yaml in projectDir (ignoring it in .git/info/exclude) and opens its tab.
func (a *App) InitProject(projectDir, name, prefix string, useGitignore bool) (AppState, error) {
	abs, err := filepath.Abs(registry.ExpandHome(projectDir))
	if err != nil {
		return AppState{}, err
	}
	if _, _, err := board.InitProject(abs, name, prefix, useGitignore); err != nil {
		return AppState{}, err
	}
	return a.OpenProjectTab(abs)
}

func (a *App) saveBoardLocked(projectPath string, b *board.Board) error {
	hash, err := board.SaveAtomic(projectPath, b)
	if err != nil {
		return err
	}
	if a.watcher != nil {
		a.watcher.RecordSelfWrite(projectPath, hash)
	}
	a.boards[projectPath] = b
	delete(a.errors, projectPath)
	return nil
}

// CaptureResponse is returned after local LLM (or heuristic) capture creates a task.
type CaptureResponse struct {
	Board  *board.Board `json:"board"`
	Task   board.Task   `json:"task"`
	UsedAI bool         `json:"used_ai"`
	Model  string       `json:"model,omitempty"`
}

// CapturedTicket identifies a single ticket created during project-agnostic batch capture.
type CapturedTicket struct {
	ProjectPath string     `json:"project_path"`
	ProjectName string     `json:"project_name"`
	ProjectSlug string     `json:"project_slug"`
	Task        board.Task `json:"task"`
}

// CaptureBatchResponse is returned after project-agnostic capture creates one or more tickets.
type CaptureBatchResponse struct {
	State   AppState         `json:"state"`
	Tickets []CapturedTicket `json:"tickets"`
	UsedAI  bool             `json:"used_ai"`
	Model   string           `json:"model,omitempty"`
}

// CaptureContext processes freeform user context across all known Kingfisher projects,
// automatically routing each item to the matching project and splitting distinct ideas into
// separate tickets.
func (a *App) CaptureContext(rawContext, activeProjectPath, defaultStatus string) (CaptureBatchResponse, error) {
	rawContext = strings.TrimSpace(rawContext)
	if rawContext == "" {
		return CaptureBatchResponse{}, fmt.Errorf("capture context cannot be empty")
	}

	a.mu.Lock()
	projectsCopy := append([]registry.ProjectEntry(nil), a.reg.Projects...)
	aiCfg := a.reg.Preferences.AI
	ctx := a.ctx
	loadedBoards := make(map[string]*board.Board, len(a.boards))
	for k, v := range a.boards {
		loadedBoards[k] = v
	}
	a.mu.Unlock()

	if ctx == nil {
		ctx = context.Background()
	}

	var contexts []capture.ProjectContext
	defaultSlug := ""
	for _, p := range projectsCopy {
		if !board.Exists(p.Path) {
			continue
		}
		var b *board.Board
		if cached, ok := loadedBoards[p.Path]; ok {
			b = cached
		} else if diskBoard, _, err := board.Load(p.Path); err == nil {
			b = diskBoard
		}
		pc := capture.ProjectContext{
			Slug:   p.Slug,
			Name:   p.Name,
			Prefix: p.Prefix,
			Path:   p.Path,
		}
		if b != nil {
			pc.Columns = append([]board.Column(nil), b.Columns...)
			seen := map[string]bool{}
			for _, t := range b.Tasks {
				for _, tg := range t.Tags {
					low := strings.ToLower(strings.TrimSpace(tg))
					if low != "" && !seen[low] {
						seen[low] = true
						pc.Tags = append(pc.Tags, low)
					}
				}
			}
		}
		contexts = append(contexts, pc)
		if p.Path == activeProjectPath {
			defaultSlug = p.Slug
		}
	}

	if len(contexts) == 0 {
		return CaptureBatchResponse{}, fmt.Errorf("no Kingfisher projects registered yet — open or initialize a folder first")
	}
	if defaultSlug == "" {
		defaultSlug = contexts[0].Slug
	}

	batch := capture.ProcessBatch(ctx, aiCfg, rawContext, contexts, defaultSlug, defaultStatus)

	pathBySlug := make(map[string]capture.ProjectContext, len(contexts))
	for _, c := range contexts {
		pathBySlug[c.Slug] = c
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	tickets := make([]CapturedTicket, 0, len(batch.Items))
	for _, item := range batch.Items {
		target, ok := pathBySlug[item.ProjectSlug]
		if !ok {
			target = contexts[0]
		}

		b, isOpen := a.boards[target.Path]
		if !isOpen {
			var err error
			b, _, err = board.Load(target.Path)
			if err != nil {
				continue
			}
		}

		task := b.AddTaskWithRaw(item.Title, item.Status, item.Priority, item.Description, item.RawInput, item.Tags)
		hash, err := board.SaveAtomic(target.Path, b)
		if err != nil {
			return CaptureBatchResponse{}, err
		}
		if isOpen {
			if a.watcher != nil {
				a.watcher.RecordSelfWrite(target.Path, hash)
			}
			a.boards[target.Path] = b
			delete(a.errors, target.Path)
		}

		tickets = append(tickets, CapturedTicket{
			ProjectPath: target.Path,
			ProjectName: target.Name,
			ProjectSlug: target.Slug,
			Task:        task,
		})
	}

	if len(tickets) == 0 {
		return CaptureBatchResponse{}, fmt.Errorf("failed to file captured items")
	}

	return CaptureBatchResponse{
		State:   a.snapshotLocked(),
		Tickets: tickets,
		UsedAI:  batch.UsedAI,
		Model:   batch.Model,
	}, nil
}

// CaptureTask processes raw freeform context with the configured local small LLM
// (falling back to heuristic parsing when offline) and adds the resulting task to the board.
func (a *App) CaptureTask(projectPath, rawContext, defaultStatus string) (CaptureResponse, error) {
	rawContext = strings.TrimSpace(rawContext)
	if rawContext == "" {
		return CaptureResponse{}, fmt.Errorf("capture context cannot be empty")
	}

	a.mu.Lock()
	b, ok := a.boards[projectPath]
	if !ok {
		a.mu.Unlock()
		return CaptureResponse{}, fmt.Errorf("project not loaded: %s", projectPath)
	}
	// Snapshot board metadata & AI config before releasing lock during LLM call
	boardSnapshot := &board.Board{
		Project: b.Project,
		Columns: append([]board.Column(nil), b.Columns...),
		Tasks:   append([]board.Task(nil), b.Tasks...),
	}
	aiCfg := a.reg.Preferences.AI
	ctx := a.ctx
	a.mu.Unlock()

	if ctx == nil {
		ctx = context.Background()
	}

	res := capture.Process(ctx, aiCfg, rawContext, boardSnapshot, defaultStatus)

	a.mu.Lock()
	defer a.mu.Unlock()

	b, ok = a.boards[projectPath]
	if !ok {
		return CaptureResponse{}, fmt.Errorf("project not loaded: %s", projectPath)
	}
	task := b.AddTaskWithRaw(res.Title, res.Status, res.Priority, res.Description, res.RawInput, res.Tags)
	if err := a.saveBoardLocked(projectPath, b); err != nil {
		return CaptureResponse{}, err
	}
	return CaptureResponse{
		Board:  b,
		Task:   task,
		UsedAI: res.UsedAI,
		Model:  res.Model,
	}, nil
}

// GetAIStatus probes the configured local LLM endpoint and returns live availability.
func (a *App) GetAIStatus() capture.Status {
	a.mu.Lock()
	cfg := a.reg.Preferences.AI
	ctx := a.ctx
	a.mu.Unlock()
	if ctx == nil {
		ctx = context.Background()
	}
	return capture.CheckStatus(ctx, cfg)
}

// UpdateAIConfig updates and persists the local LLM configuration.
func (a *App) UpdateAIConfig(cfg capture.Config) registry.Preferences {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.reg.Preferences.AI = cfg
	_ = a.reg.Save()
	return a.reg.Preferences
}

// UpdatePreferences persists full accessibility, appearance, scan directory, and AI settings.
func (a *App) UpdatePreferences(prefs registry.Preferences) AppState {
	a.mu.Lock()
	defer a.mu.Unlock()

	if prefs.Opacity < 0 {
		prefs.Opacity = 0
	}
	if prefs.Opacity > 0.85 {
		prefs.Opacity = 0.85
	}
	cleanDirs := make([]string, 0, len(prefs.ScanDirs))
	for _, d := range prefs.ScanDirs {
		if s := strings.TrimSpace(d); s != "" {
			cleanDirs = append(cleanDirs, s)
		}
	}
	prefs.ScanDirs = cleanDirs
	a.reg.Preferences = prefs
	a.reg.ScanProjects()
	_ = a.reg.Save()
	return a.snapshotLocked()
}

// CreateTask adds a new task to the specified project board and saves atomically.
func (a *App) CreateTask(projectPath, title, status, priority, description string, tags []string) (*board.Board, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	b, ok := a.boards[projectPath]
	if !ok {
		return nil, fmt.Errorf("project not loaded: %s", projectPath)
	}
	b.AddTask(title, status, priority, description, tags)
	if err := a.saveBoardLocked(projectPath, b); err != nil {
		return nil, err
	}
	return b, nil
}

// UpdateTask modifies an existing task and saves atomically.
func (a *App) UpdateTask(projectPath, taskID string, patch board.TaskPatch) (*board.Board, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	b, ok := a.boards[projectPath]
	if !ok {
		return nil, fmt.Errorf("project not loaded: %s", projectPath)
	}
	if _, err := b.UpdateTask(taskID, patch); err != nil {
		return nil, err
	}
	if err := a.saveBoardLocked(projectPath, b); err != nil {
		return nil, err
	}
	return b, nil
}

// MoveTask moves a task across columns or reorders it within a column.
func (a *App) MoveTask(projectPath, taskID, targetStatus string, targetIndex int) (*board.Board, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	b, ok := a.boards[projectPath]
	if !ok {
		return nil, fmt.Errorf("project not loaded: %s", projectPath)
	}
	if _, err := b.MoveTask(taskID, targetStatus, targetIndex, ""); err != nil {
		return nil, err
	}
	if err := a.saveBoardLocked(projectPath, b); err != nil {
		return nil, err
	}
	return b, nil
}

// DeleteTask deletes a task from the project board.
func (a *App) DeleteTask(projectPath, taskID string) (*board.Board, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	b, ok := a.boards[projectPath]
	if !ok {
		return nil, fmt.Errorf("project not loaded: %s", projectPath)
	}
	if err := b.DeleteTask(taskID); err != nil {
		return nil, err
	}
	if err := a.saveBoardLocked(projectPath, b); err != nil {
		return nil, err
	}
	return b, nil
}

// SetOpacity saves the user's stained-glass opacity preference (0.0 to 0.85).
func (a *App) SetOpacity(opacity float64) registry.Preferences {
	a.mu.Lock()
	defer a.mu.Unlock()

	if opacity < 0 {
		opacity = 0
	}
	if opacity > 0.85 {
		opacity = 0.85
	}
	a.reg.Preferences.Opacity = opacity
	_ = a.reg.Save()
	return a.reg.Preferences
}

// UpdateScanDirs saves scan directories and rescans for projects.
func (a *App) UpdateScanDirs(dirs []string) AppState {
	a.mu.Lock()
	defer a.mu.Unlock()

	clean := make([]string, 0, len(dirs))
	for _, d := range dirs {
		d = strings.TrimSpace(d)
		if d != "" {
			clean = append(clean, d)
		}
	}
	a.reg.Preferences.ScanDirs = clean
	a.reg.ScanProjects()
	_ = a.reg.Save()
	return a.snapshotLocked()
}

// HideWindow hides Kingfisher in the background for instant Raycast Super+F re-summon.
func (a *App) HideWindow() {
	if a.ctx != nil {
		wailsruntime.WindowHide(a.ctx)
	}
}

// MinimizeWindow minimizes the window.
func (a *App) MinimizeWindow() {
	if a.ctx != nil {
		wailsruntime.WindowMinimise(a.ctx)
	}
}

// ToggleMaximizeWindow maximizes or unmaximizes the window.
func (a *App) ToggleMaximizeWindow() {
	if a.ctx != nil {
		wailsruntime.WindowToggleMaximise(a.ctx)
	}
}
