package watcher

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Daniel-A-C/kingfisher/internal/board"
	"github.com/fsnotify/fsnotify"
)

// Callbacks receives live updates and parse errors from watched .kingfisher.yaml files.
type Callbacks struct {
	OnUpdate func(projectDir string, b *board.Board)
	OnError  func(projectDir string, errMsg string)
}

// Manager watches multiple project directories for changes to .kingfisher.yaml.
type Manager struct {
	mu       sync.Mutex
	fsw      *fsnotify.Watcher
	watched  map[string]bool
	lastHash map[string]string
	timers   map[string]*time.Timer
	cb       Callbacks
	debounce time.Duration
	done     chan struct{}
}

// New creates a new file watcher Manager.
func New(cb Callbacks) (*Manager, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	m := &Manager{
		fsw:      fsw,
		watched:  make(map[string]bool),
		lastHash: make(map[string]string),
		timers:   make(map[string]*time.Timer),
		cb:       cb,
		debounce: 120 * time.Millisecond,
		done:     make(chan struct{}),
	}
	go m.loop()
	return m, nil
}

// Watch starts watching projectDir for edits or atomic renames to .kingfisher.yaml.
func (m *Manager) Watch(projectDir, initialHash string) error {
	abs, err := filepath.Abs(projectDir)
	if err != nil {
		abs = filepath.Clean(projectDir)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if initialHash != "" {
		m.lastHash[abs] = initialHash
	}
	if m.watched[abs] {
		return nil
	}
	if err := m.fsw.Add(abs); err != nil {
		return err
	}
	m.watched[abs] = true
	return nil
}

// Unwatch stops watching projectDir.
func (m *Manager) Unwatch(projectDir string) {
	abs, err := filepath.Abs(projectDir)
	if err != nil {
		abs = filepath.Clean(projectDir)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.watched[abs] {
		return
	}
	_ = m.fsw.Remove(abs)
	delete(m.watched, abs)
	delete(m.lastHash, abs)
	if t, ok := m.timers[abs]; ok {
		t.Stop()
		delete(m.timers, abs)
	}
}

// RecordSelfWrite records the SHA-256 hash of a write initiated by Kingfisher itself
// so the corresponding fsnotify event does not trigger a redundant UI reload.
func (m *Manager) RecordSelfWrite(projectDir, hash string) {
	abs, err := filepath.Abs(projectDir)
	if err != nil {
		abs = filepath.Clean(projectDir)
	}
	m.mu.Lock()
	m.lastHash[abs] = hash
	m.mu.Unlock()
}

// Close shuts down the underlying fsnotify watcher.
func (m *Manager) Close() error {
	close(m.done)
	m.mu.Lock()
	for _, t := range m.timers {
		t.Stop()
	}
	m.timers = make(map[string]*time.Timer)
	m.mu.Unlock()
	return m.fsw.Close()
}

func (m *Manager) loop() {
	for {
		select {
		case <-m.done:
			return
		case ev, ok := <-m.fsw.Events:
			if !ok {
				return
			}
			if filepath.Base(ev.Name) != board.BoardFilename {
				continue
			}
			if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}
			dir := filepath.Dir(ev.Name)
			m.scheduleReload(dir)
		case _, ok := <-m.fsw.Errors:
			if !ok {
				return
			}
		}
	}
}

func (m *Manager) scheduleReload(projectDir string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.watched[projectDir] {
		return
	}
	if t, ok := m.timers[projectDir]; ok {
		t.Stop()
	}
	m.timers[projectDir] = time.AfterFunc(m.debounce, func() {
		m.handleReload(projectDir)
	})
}

func (m *Manager) handleReload(projectDir string) {
	raw, err := os.ReadFile(board.BoardPath(projectDir))
	if err != nil {
		// File might be mid-rename or temporarily absent; skip quietly unless persistent
		return
	}

	hash := board.HashBytes(raw)
	m.mu.Lock()
	if !m.watched[projectDir] {
		m.mu.Unlock()
		return
	}
	if m.lastHash[projectDir] == hash {
		m.mu.Unlock()
		return
	}
	m.mu.Unlock()

	b, parsedHash, err := board.Load(projectDir)
	if err != nil {
		if m.cb.OnError != nil {
			m.cb.OnError(projectDir, err.Error())
		}
		return
	}

	m.mu.Lock()
	m.lastHash[projectDir] = parsedHash
	m.mu.Unlock()

	if m.cb.OnUpdate != nil {
		m.cb.OnUpdate(projectDir, b)
	}
}
