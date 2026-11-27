package board

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	PriorityLow      = "low"
	PriorityMedium   = "medium"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)

// Board represents a single project's .kingfisher.yaml file.
type Board struct {
	Project ProjectMeta `yaml:"project" json:"project"`
	Columns []Column    `yaml:"columns" json:"columns"`
	Tasks   []Task      `yaml:"tasks"   json:"tasks"`
}

// ProjectMeta holds project-level metadata and ticket ID sequencing.
type ProjectMeta struct {
	Name   string `yaml:"name"    json:"name"`
	Prefix string `yaml:"prefix"  json:"prefix"`
	NextID int    `yaml:"next_id" json:"next_id"`
}

// Column represents a Kanban column on the board.
type Column struct {
	ID   string `yaml:"id"   json:"id"`
	Name string `yaml:"name" json:"name"`
}

// Task represents a single card/issue in the project board.
type Task struct {
	ID          string   `yaml:"id"                    json:"id"`
	Title       string   `yaml:"title"                 json:"title"`
	Status      string   `yaml:"status"                json:"status"`
	Priority    string   `yaml:"priority"              json:"priority"`
	Tags        []string `yaml:"tags,omitempty"        json:"tags"`
	Created     string   `yaml:"created"               json:"created"`
	Updated     string   `yaml:"updated"               json:"updated"`
	Description string   `yaml:"description,omitempty" json:"description"`
	AgentNotes  string   `yaml:"agent_notes"           json:"agent_notes"`
}

// TaskPatch holds optional updates for an existing task.
type TaskPatch struct {
	Title       *string  `json:"title,omitempty"`
	Status      *string  `json:"status,omitempty"`
	Priority    *string  `json:"priority,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	HasTags     bool     `json:"has_tags,omitempty"`
	Description *string  `json:"description,omitempty"`
	AgentNotes  *string  `json:"agent_notes,omitempty"`
}

// DefaultColumns returns the 4 default Kingfisher Kanban columns.
func DefaultColumns() []Column {
	return []Column{
		{ID: "backlog", Name: "Backlog"},
		{ID: "ready", Name: "Ready to Dive"},
		{ID: "active", Name: "In Flight"},
		{ID: "done", Name: "Landed"},
	}
}

// PriorityRank returns a numeric sort rank where higher means more urgent.
func PriorityRank(p string) int {
	switch strings.ToLower(strings.TrimSpace(p)) {
	case PriorityCritical:
		return 4
	case PriorityHigh:
		return 3
	case PriorityMedium:
		return 2
	case PriorityLow:
		return 1
	default:
		return 0
	}
}

// NormalizePriority ensures a valid priority string, defaulting to medium.
func NormalizePriority(p string) string {
	switch strings.ToLower(strings.TrimSpace(p)) {
	case PriorityLow:
		return PriorityLow
	case PriorityMedium:
		return PriorityMedium
	case PriorityHigh:
		return PriorityHigh
	case PriorityCritical:
		return PriorityCritical
	default:
		return PriorityMedium
	}
}

// Validate checks board integrity and normalizes minor missing defaults.
func (b *Board) Validate() error {
	if strings.TrimSpace(b.Project.Name) == "" {
		return fmt.Errorf("project.name is required")
	}
	if strings.TrimSpace(b.Project.Prefix) == "" {
		b.Project.Prefix = DerivePrefix(b.Project.Name)
	}
	if len(b.Columns) == 0 {
		b.Columns = DefaultColumns()
	}

	colIDs := make(map[string]bool, len(b.Columns))
	for i, c := range b.Columns {
		if strings.TrimSpace(c.ID) == "" {
			return fmt.Errorf("columns[%d].id is required", i)
		}
		if colIDs[c.ID] {
			return fmt.Errorf("duplicate column id %q", c.ID)
		}
		colIDs[c.ID] = true
	}

	if b.Tasks == nil {
		b.Tasks = []Task{}
	}

	maxNum := 0
	seenIDs := make(map[string]bool, len(b.Tasks))
	for i := range b.Tasks {
		t := &b.Tasks[i]
		if strings.TrimSpace(t.ID) == "" {
			return fmt.Errorf("tasks[%d].id is required", i)
		}
		if seenIDs[t.ID] {
			return fmt.Errorf("duplicate task id %q", t.ID)
		}
		seenIDs[t.ID] = true

		if strings.TrimSpace(t.Title) == "" {
			return fmt.Errorf("task %q title is required", t.ID)
		}
		if !colIDs[t.Status] {
			// Gracefully fall back to first column if status is unknown
			t.Status = b.Columns[0].ID
		}
		t.Priority = NormalizePriority(t.Priority)
		if t.Tags == nil {
			t.Tags = []string{}
		}

		if n := extractIDNumber(t.ID); n > maxNum {
			maxNum = n
		}
	}

	if b.Project.NextID <= maxNum {
		b.Project.NextID = maxNum + 1
	}
	if b.Project.NextID < 1 {
		b.Project.NextID = 1
	}

	return nil
}

// HasColumn reports whether a column ID exists on the board.
func (b *Board) HasColumn(colID string) bool {
	for _, c := range b.Columns {
		if c.ID == colID {
			return true
		}
	}
	return false
}

func extractIDNumber(id string) int {
	idx := strings.LastIndex(id, "-")
	if idx == -1 || idx+1 >= len(id) {
		return 0
	}
	n, err := strconv.Atoi(id[idx+1:])
	if err != nil {
		return 0
	}
	return n
}
