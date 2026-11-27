package board

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	BoardFilename = ".kingfisher.yaml"
	fileHeader    = "# Kingfisher Project Board — Local Only (gitignored)\n"
)

// BoardPath returns the full path to .kingfisher.yaml inside projectDir.
func BoardPath(projectDir string) string {
	return filepath.Join(projectDir, BoardFilename)
}

// Exists reports whether projectDir contains a .kingfisher.yaml file.
func Exists(projectDir string) bool {
	info, err := os.Stat(BoardPath(projectDir))
	return err == nil && !info.IsDir()
}

// HashBytes returns the hex-encoded SHA-256 hash of data.
func HashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// Load reads, parses, and validates .kingfisher.yaml from projectDir.
// Returns the parsed Board and the SHA-256 hash of the raw file bytes.
func Load(projectDir string) (*Board, string, error) {
	path := BoardPath(projectDir)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", path, err)
	}

	var b Board
	if err := yaml.Unmarshal(data, &b); err != nil {
		return nil, "", fmt.Errorf("parse %s: %w", path, err)
	}
	if err := b.Validate(); err != nil {
		return nil, "", fmt.Errorf("validate %s: %w", path, err)
	}

	return &b, HashBytes(data), nil
}

// Marshal serializes a Board into clean YAML with the Kingfisher header.
func Marshal(b *Board) ([]byte, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString(fileHeader)

	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(b); err != nil {
		_ = enc.Close()
		return nil, fmt.Errorf("marshal yaml: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("close yaml encoder: %w", err)
	}

	return buf.Bytes(), nil
}

// SaveAtomic writes .kingfisher.yaml via a temporary file + fsync + os.Rename
// so external readers (like coding agents) never observe a partial write.
// Returns the SHA-256 hash of the written bytes.
func SaveAtomic(projectDir string, b *Board) (string, error) {
	data, err := Marshal(b)
	if err != nil {
		return "", err
	}

	target := BoardPath(projectDir)
	tmp := target + ".tmp"

	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return "", fmt.Errorf("open temp board file: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return "", fmt.Errorf("write temp board file: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return "", fmt.Errorf("sync temp board file: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("close temp board file: %w", err)
	}

	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("rename temp board file: %w", err)
	}

	return HashBytes(data), nil
}

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// AddTask appends a new task with an auto-incremented ID and returns it.
func (b *Board) AddTask(title, status, priority, description string, tags []string) Task {
	title = strings.TrimSpace(title)
	if !b.HasColumn(status) {
		status = b.Columns[0].ID
	}
	if tags == nil {
		tags = []string{}
	}
	cleanTags := make([]string, 0, len(tags))
	for _, tg := range tags {
		tg = strings.TrimSpace(tg)
		if tg != "" {
			cleanTags = append(cleanTags, tg)
		}
	}

	now := nowISO()
	t := Task{
		ID:          fmt.Sprintf("%s-%d", b.Project.Prefix, b.Project.NextID),
		Title:       title,
		Status:      status,
		Priority:    NormalizePriority(priority),
		Tags:        cleanTags,
		Created:     now,
		Updated:     now,
		Description: strings.TrimSpace(description),
		AgentNotes:  "",
	}
	b.Project.NextID++
	b.Tasks = append(b.Tasks, t)
	return t
}

// AddTaskWithRaw appends a new task including the verbatim original human capture input.
func (b *Board) AddTaskWithRaw(title, status, priority, description, rawInput string, tags []string) Task {
	t := b.AddTask(title, status, priority, description, tags)
	trimmedRaw := strings.TrimSpace(rawInput)
	if trimmedRaw != "" {
		b.Tasks[len(b.Tasks)-1].RawInput = trimmedRaw
		t.RawInput = trimmedRaw
	}
	return t
}

// UpdateTask applies partial updates to the task matching id.
func (b *Board) UpdateTask(id string, patch TaskPatch) (Task, error) {
	for i := range b.Tasks {
		if strings.EqualFold(b.Tasks[i].ID, id) {
			t := &b.Tasks[i]
			if patch.Title != nil {
				title := strings.TrimSpace(*patch.Title)
				if title == "" {
					return Task{}, fmt.Errorf("task title cannot be empty")
				}
				t.Title = title
			}
			if patch.Status != nil {
				if !b.HasColumn(*patch.Status) {
					return Task{}, fmt.Errorf("unknown column status %q", *patch.Status)
				}
				t.Status = *patch.Status
			}
			if patch.Priority != nil {
				t.Priority = NormalizePriority(*patch.Priority)
			}
			if patch.HasTags {
				clean := make([]string, 0, len(patch.Tags))
				for _, tg := range patch.Tags {
					tg = strings.TrimSpace(tg)
					if tg != "" {
						clean = append(clean, tg)
					}
				}
				t.Tags = clean
			}
			if patch.Description != nil {
				t.Description = strings.TrimSpace(*patch.Description)
			}
			if patch.RawInput != nil {
				t.RawInput = strings.TrimSpace(*patch.RawInput)
			}
			if patch.AgentNotes != nil {
				t.AgentNotes = strings.TrimSpace(*patch.AgentNotes)
			}
			t.Updated = nowISO()
			return *t, nil
		}
	}
	return Task{}, fmt.Errorf("task %q not found", id)
}

// MoveTask changes a task's status (and optionally reorders it within the target column)
// and optionally appends a note to AgentNotes.
func (b *Board) MoveTask(id, targetStatus string, targetIndex int, appendNote string) (Task, error) {
	if !b.HasColumn(targetStatus) {
		return Task{}, fmt.Errorf("unknown column status %q", targetStatus)
	}

	foundIdx := -1
	for i := range b.Tasks {
		if strings.EqualFold(b.Tasks[i].ID, id) {
			foundIdx = i
			break
		}
	}
	if foundIdx == -1 {
		return Task{}, fmt.Errorf("task %q not found", id)
	}

	task := b.Tasks[foundIdx]
	task.Status = targetStatus
	task.Updated = nowISO()

	if note := strings.TrimSpace(appendNote); note != "" {
		if strings.TrimSpace(task.AgentNotes) == "" {
			task.AgentNotes = note
		} else {
			task.AgentNotes = strings.TrimRight(task.AgentNotes, "\n") + "\n" + note
		}
	}

	// Remove task from its old position
	remaining := make([]Task, 0, len(b.Tasks))
	remaining = append(remaining, b.Tasks[:foundIdx]...)
	remaining = append(remaining, b.Tasks[foundIdx+1:]...)

	// Insert at targetIndex within the tasks that share targetStatus
	if targetIndex < 0 {
		remaining = append(remaining, task)
		b.Tasks = remaining
		return task, nil
	}

	colPos := 0
	insertAt := len(remaining)
	for i, t := range remaining {
		if t.Status == targetStatus {
			if colPos == targetIndex {
				insertAt = i
				break
			}
			colPos++
		}
	}

	result := make([]Task, 0, len(remaining)+1)
	result = append(result, remaining[:insertAt]...)
	result = append(result, task)
	result = append(result, remaining[insertAt:]...)
	b.Tasks = result

	return task, nil
}

// DeleteTask removes the task matching id.
func (b *Board) DeleteTask(id string) error {
	for i := range b.Tasks {
		if strings.EqualFold(b.Tasks[i].ID, id) {
			b.Tasks = append(b.Tasks[:i], b.Tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task %q not found", id)
}

// NextReadyTask returns the highest-priority task in the "ready" column
// (falling back to "backlog" if no "ready" tasks exist).
func (b *Board) NextReadyTask() (*Task, bool) {
	pickBest := func(status string) *Task {
		var best *Task
		for i := range b.Tasks {
			t := &b.Tasks[i]
			if t.Status != status {
				continue
			}
			if best == nil || PriorityRank(t.Priority) > PriorityRank(best.Priority) {
				best = t
			}
		}
		return best
	}

	if t := pickBest("ready"); t != nil {
		cp := *t
		return &cp, true
	}
	if t := pickBest("backlog"); t != nil {
		cp := *t
		return &cp, true
	}
	return nil, false
}
