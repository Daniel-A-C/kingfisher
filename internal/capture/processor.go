package capture

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/Daniel-A-C/kingfisher/internal/board"
)

// Config controls the local open-source LLM used to structure raw captured context.
type Config struct {
	Enabled        bool   `json:"enabled"`
	Provider       string `json:"provider"` // "ollama" | "openai"
	Endpoint       string `json:"endpoint"`
	Model          string `json:"model"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
}

// Status describes the live availability of the configured local LLM runtime.
type Status struct {
	Available bool     `json:"available"`
	Provider  string   `json:"provider"`
	Model     string   `json:"model"`
	Endpoint  string   `json:"endpoint"`
	Models    []string `json:"models"`
	Message   string   `json:"message"`
}

// Result is a single structured task extracted from raw user context.
type Result struct {
	ProjectSlug string   `json:"project,omitempty"`
	Title       string   `json:"title"`
	Status      string   `json:"status"`
	Priority    string   `json:"priority"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	UsedAI      bool     `json:"used_ai"`
	Model       string   `json:"model,omitempty"`
}

// ProjectContext describes a known Kingfisher project for project-agnostic routing.
type ProjectContext struct {
	Slug    string
	Name    string
	Prefix  string
	Path    string
	Columns []board.Column
	Tags    []string
}

// BatchResult holds one or more structured tickets routed to their respective projects.
type BatchResult struct {
	Items  []Result `json:"items"`
	UsedAI bool     `json:"used_ai"`
	Model  string   `json:"model,omitempty"`
}

// DefaultConfig returns sensible defaults for a very small, fast local Ollama model.
func DefaultConfig() Config {
	return Config{
		Enabled:        true,
		Provider:       "ollama",
		Endpoint:       "http://127.0.0.1:11434",
		Model:          "qwen2.5:1.5b",
		TimeoutSeconds: 12,
	}
}

func (c Config) normalized() Config {
	def := DefaultConfig()
	if strings.TrimSpace(c.Provider) == "" {
		c.Provider = def.Provider
	}
	if strings.TrimSpace(c.Endpoint) == "" {
		c.Endpoint = def.Endpoint
	}
	c.Endpoint = strings.TrimRight(strings.TrimSpace(c.Endpoint), "/")
	if strings.TrimSpace(c.Model) == "" {
		c.Model = def.Model
	}
	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = def.TimeoutSeconds
	}
	return c
}

// CheckStatus probes the configured local model endpoint quickly (600ms timeout).
func CheckStatus(ctx context.Context, cfg Config) Status {
	cfg = cfg.normalized()
	st := Status{
		Provider: cfg.Provider,
		Model:    cfg.Model,
		Endpoint: cfg.Endpoint,
		Models:   []string{},
	}
	if !cfg.Enabled {
		st.Message = "Local AI disabled (using fast heuristic parser)"
		return st
	}

	probeCtx, cancel := context.WithTimeout(ctx, 600*time.Millisecond)
	defer cancel()
	client := &http.Client{Timeout: 600 * time.Millisecond}

	if strings.EqualFold(cfg.Provider, "openai") {
		req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, cfg.Endpoint+"/v1/models", nil)
		if err != nil {
			st.Message = "Local OpenAI-compatible server unavailable"
			return st
		}
		resp, err := client.Do(req)
		if err != nil {
			st.Message = "Local server offline — heuristic fallback active"
			return st
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			var parsed struct {
				Data []struct {
					ID string `json:"id"`
				} `json:"data"`
			}
			_ = json.NewDecoder(resp.Body).Decode(&parsed)
			for _, m := range parsed.Data {
				st.Models = append(st.Models, m.ID)
			}
			st.Available = true
			st.Message = fmt.Sprintf("Connected to %s (%s)", cfg.Provider, cfg.Model)
			return st
		}
		st.Message = fmt.Sprintf("Server returned HTTP %d", resp.StatusCode)
		return st
	}

	// Default: Ollama (/api/tags)
	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, cfg.Endpoint+"/api/tags", nil)
	if err != nil {
		st.Message = "Ollama unavailable"
		return st
	}
	resp, err := client.Do(req)
	if err != nil {
		st.Message = "Ollama not running — heuristic fallback active"
		return st
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		st.Message = fmt.Sprintf("Ollama returned HTTP %d", resp.StatusCode)
		return st
	}

	var tagsResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		st.Message = "Could not parse Ollama models"
		return st
	}

	hasConfigured := false
	for _, m := range tagsResp.Models {
		st.Models = append(st.Models, m.Name)
		if m.Name == cfg.Model || strings.HasPrefix(m.Name, cfg.Model+":") {
			hasConfigured = true
		}
	}

	if hasConfigured {
		st.Available = true
		st.Message = fmt.Sprintf("Local AI ready (%s)", cfg.Model)
		return st
	}
	if len(st.Models) > 0 {
		st.Available = true
		st.Model = st.Models[0]
		st.Message = fmt.Sprintf("Local AI ready (%s)", st.Model)
		return st
	}

	st.Message = fmt.Sprintf("Ollama running, pull %s to enable AI capture", cfg.Model)
	return st
}

// Process turns raw freeform user context into a structured board task using the local LLM,
// falling back seamlessly to HeuristicParse if the local model is unavailable.
func Process(ctx context.Context, cfg Config, rawContext string, b *board.Board, defaultStatus string) Result {
	rawContext = strings.TrimSpace(rawContext)
	cfg = cfg.normalized()

	if defaultStatus == "" || (b != nil && !b.HasColumn(defaultStatus)) {
		if b != nil && b.HasColumn("ready") {
			defaultStatus = "ready"
		} else if b != nil && len(b.Columns) > 0 {
			defaultStatus = b.Columns[0].ID
		} else {
			defaultStatus = "ready"
		}
	}

	if cfg.Enabled && rawContext != "" {
		if res, err := callLocalLLM(ctx, cfg, rawContext, b, defaultStatus); err == nil {
			return res
		}
	}

	return HeuristicParse(rawContext, b, defaultStatus)
}

func buildPrompt(rawContext string, b *board.Board, defaultStatus string) (string, string) {
	var cols []string
	var existingTags []string
	seenTag := map[string]bool{}
	projectName := "Project"

	if b != nil {
		if b.Project.Name != "" {
			projectName = b.Project.Name
		}
		for _, c := range b.Columns {
			cols = append(cols, fmt.Sprintf("%s (%s)", c.ID, c.Name))
		}
		for _, t := range b.Tasks {
			for _, tg := range t.Tags {
				low := strings.ToLower(strings.TrimSpace(tg))
				if low != "" && !seenTag[low] {
					seenTag[low] = true
					existingTags = append(existingTags, low)
				}
			}
		}
	}
	if len(cols) == 0 {
		cols = []string{"backlog (Backlog)", "ready (Ready to Dive)", "active (In Flight)", "done (Landed)"}
	}

	sys := fmt.Sprintf(`You are a fast, concise task-triage assistant for the local project board "%s".
Convert the user's raw note into a single well-structured Kanban ticket.
Return ONLY valid JSON with these exact keys:
- "title": string (concise, imperative, <= 70 characters)
- "status": string (must be one of the column IDs: %s; default to "%s" unless the text implies another stage like backlog/someday or active/in-progress)
- "priority": string (one of: "low", "medium", "high", "critical")
- "tags": array of 1 to 4 short lowercase kebab-case tags (reuse existing tags when relevant: [%s])
- "description": string (cleanly formatted, concise markdown capturing all details/acceptance criteria from the user's input without inventing requirements)`,
		projectName,
		strings.Join(cols, ", "),
		defaultStatus,
		strings.Join(existingTags, ", "),
	)

	return sys, rawContext
}

func callLocalLLM(ctx context.Context, cfg Config, rawContext string, b *board.Board, defaultStatus string) (Result, error) {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// For Ollama, if configured model isn't pulled yet, try first available model
	modelName := cfg.Model
	if strings.EqualFold(cfg.Provider, "ollama") {
		st := CheckStatus(reqCtx, cfg)
		if !st.Available {
			return Result{}, fmt.Errorf("ollama unavailable: %s", st.Message)
		}
		if st.Model != "" {
			modelName = st.Model
		}
	}

	sysPrompt, userPrompt := buildPrompt(rawContext, b, defaultStatus)
	client := &http.Client{Timeout: timeout}

	var content string
	var err error
	if strings.EqualFold(cfg.Provider, "openai") {
		content, err = callOpenAICompat(reqCtx, client, cfg.Endpoint, modelName, sysPrompt, userPrompt)
	} else {
		content, err = callOllama(reqCtx, client, cfg.Endpoint, modelName, sysPrompt, userPrompt)
	}
	if err != nil {
		return Result{}, err
	}

	return parseModelJSON(content, rawContext, b, defaultStatus, modelName)
}

func callOllama(ctx context.Context, client *http.Client, endpoint, model, sysPrompt, userPrompt string) (string, error) {
	payload := map[string]any{
		"model":  model,
		"stream": false,
		"format": "json",
		"options": map[string]any{
			"temperature": 0.2,
		},
		"messages": []map[string]string{
			{"role": "system", "content": sysPrompt},
			{"role": "user", "content": userPrompt},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("ollama HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	var out struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Message.Content, nil
}

func callOpenAICompat(ctx context.Context, client *http.Client, endpoint, model, sysPrompt, userPrompt string) (string, error) {
	payload := map[string]any{
		"model":           model,
		"temperature":     0.2,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": sysPrompt},
			{"role": "user", "content": userPrompt},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("openai-compat HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("empty choices from local model")
	}
	return out.Choices[0].Message.Content, nil
}

func parseModelJSON(raw, fallbackContext string, b *board.Board, defaultStatus, modelName string) (Result, error) {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	// Extract outermost JSON object if wrapped in stray prose
	if start := strings.Index(cleaned, "{"); start >= 0 {
		if end := strings.LastIndex(cleaned, "}"); end > start {
			cleaned = cleaned[start : end+1]
		}
	}

	var parsed struct {
		Title       string   `json:"title"`
		Status      string   `json:"status"`
		Priority    string   `json:"priority"`
		Tags        []string `json:"tags"`
		Description string   `json:"description"`
	}
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		return Result{}, fmt.Errorf("parse model json: %w", err)
	}

	title := strings.TrimSpace(parsed.Title)
	if title == "" {
		return Result{}, fmt.Errorf("model returned empty title")
	}

	status := strings.ToLower(strings.TrimSpace(parsed.Status))
	if b != nil && !b.HasColumn(status) {
		status = defaultStatus
	}
	if status == "" {
		status = defaultStatus
	}

	desc := strings.TrimSpace(parsed.Description)
	if desc == "" {
		desc = strings.TrimSpace(fallbackContext)
	}

	return Result{
		Title:       title,
		Status:      status,
		Priority:    board.NormalizePriority(parsed.Priority),
		Tags:        cleanTags(parsed.Tags),
		Description: desc,
		UsedAI:      true,
		Model:       modelName,
	}, nil
}

var hashtagRe = regexp.MustCompile(`#([a-zA-Z0-9][a-zA-Z0-9_-]{0,24})`)

// HeuristicParse extracts a clean title, priority, column, tags, and description locally
// when no LLM server is running.
func HeuristicParse(rawContext string, b *board.Board, defaultStatus string) Result {
	raw := strings.TrimSpace(rawContext)
	if defaultStatus == "" {
		defaultStatus = "ready"
	}

	lower := strings.ToLower(raw)

	// 1. Extract explicit #hashtags
	var tags []string
	for _, m := range hashtagRe.FindAllStringSubmatch(raw, -1) {
		tags = append(tags, strings.ToLower(m[1]))
	}
	cleanedText := strings.TrimSpace(hashtagRe.ReplaceAllString(raw, ""))

	// 2. Infer priority
	priority := board.PriorityMedium
	switch {
	case containsAny(lower, "critical", "urgent", "asap", "blocker", "p0", "crash", "broken"):
		priority = board.PriorityCritical
	case containsAny(lower, "high priority", "important", "high:", "p1", "soon"):
		priority = board.PriorityHigh
	case containsAny(lower, "low priority", "someday", "maybe", "nit", "minor", "nice to have"):
		priority = board.PriorityLow
	}

	// 3. Infer status/column
	status := defaultStatus
	switch {
	case containsAny(lower, "backlog", "someday", "later", "eventually", "idea:"):
		if b == nil || b.HasColumn("backlog") {
			status = "backlog"
		}
	case containsAny(lower, "in progress", "working on", "currently", "in flight"):
		if b == nil || b.HasColumn("active") {
			status = "active"
		}
	}

	// 4. Infer helpful domain tags if few explicit hashtags were given
	keywordTags := []struct {
		tag      string
		keywords []string
	}{
		{"bug", []string{"bug", "fix", "crash", "broken", "glitch", "error"}},
		{"ux", []string{"ui", "ux", "css", "layout", "modal", "button", "design", "style"}},
		{"cli", []string{"cli", "terminal", "command", "flag"}},
		{"ai", []string{"llm", "model", "ollama", "prompt", "agent"}},
		{"perf", []string{"perf", "slow", "fast", "latency", "memory", "optimize"}},
		{"docs", []string{"readme", "docs", "documentation"}},
	}
	for _, kt := range keywordTags {
		if len(tags) >= 4 {
			break
		}
		if containsAny(lower, kt.keywords...) {
			tags = append(tags, kt.tag)
		}
	}

	// 5. Derive crisp title from first line/sentence
	title, desc := splitTitleAndDescription(cleanedText)
	if title == "" {
		title = "Untitled note"
	}

	return Result{
		Title:       title,
		Status:      status,
		Priority:    priority,
		Tags:        cleanTags(tags),
		Description: desc,
		UsedAI:      false,
	}
}

func splitTitleAndDescription(text string) (string, string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", ""
	}

	lines := strings.Split(text, "\n")
	firstLine := strings.TrimSpace(lines[0])
	// Strip leading priority/type prefixes like "urgent: " or "todo: "
	for _, pfx := range []string{"urgent:", "critical:", "high:", "low:", "todo:", "idea:", "bug:"} {
		if strings.HasPrefix(strings.ToLower(firstLine), pfx) {
			firstLine = strings.TrimSpace(firstLine[len(pfx):])
		}
	}

	// If first line has multiple sentences, use the first sentence for the title
	title := firstLine
	if idx := strings.IndexAny(firstLine, ".!?"); idx > 12 && idx < 80 {
		title = strings.TrimSpace(firstLine[:idx])
	}
	if len([]rune(title)) > 72 {
		rs := []rune(title)
		cut := 72
		for cut > 40 && !unicode.IsSpace(rs[cut]) {
			cut--
		}
		title = strings.TrimSpace(string(rs[:cut])) + "…"
	}

	// Capitalize first letter
	if rs := []rune(title); len(rs) > 0 {
		rs[0] = unicode.ToUpper(rs[0])
		title = string(rs)
	}

	if title == text {
		return title, ""
	}
	return title, text
}

func cleanTags(tags []string) []string {
	seen := make(map[string]bool, len(tags))
	out := make([]string, 0, len(tags))
	for _, tg := range tags {
		s := strings.ToLower(strings.TrimSpace(tg))
		s = strings.TrimPrefix(s, "#")
		s = strings.ReplaceAll(s, " ", "-")
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
		if len(out) >= 5 {
			break
		}
	}
	return out
}

func containsAny(haystack string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(haystack, n) {
			return true
		}
	}
	return false
}

// ProcessBatch processes freeform context across all known projects, automatically routing each
// item to the right project and splitting multiple distinct ideas into separate tickets.
func ProcessBatch(ctx context.Context, cfg Config, rawContext string, projects []ProjectContext, defaultProjectSlug, defaultStatus string) BatchResult {
	rawContext = strings.TrimSpace(rawContext)
	cfg = cfg.normalized()

	if len(projects) == 0 {
		single := Process(ctx, cfg, rawContext, nil, defaultStatus)
		return BatchResult{Items: []Result{single}, UsedAI: single.UsedAI, Model: single.Model}
	}

	if defaultProjectSlug == "" {
		defaultProjectSlug = projects[0].Slug
	}
	if defaultStatus == "" {
		defaultStatus = "ready"
	}

	if cfg.Enabled && rawContext != "" {
		if res, err := callLocalLLMBatch(ctx, cfg, rawContext, projects, defaultProjectSlug, defaultStatus); err == nil && len(res.Items) > 0 {
			return res
		}
	}

	return HeuristicParseBatch(rawContext, projects, defaultProjectSlug, defaultStatus)
}

func buildBatchPrompt(rawContext string, projects []ProjectContext, defaultProjectSlug, defaultStatus string) (string, string) {
	var projLines []string
	for _, p := range projects {
		var colIDs []string
		for _, c := range p.Columns {
			colIDs = append(colIDs, c.ID)
		}
		if len(colIDs) == 0 {
			colIDs = []string{"backlog", "ready", "active", "done"}
		}
		line := fmt.Sprintf(`- slug: "%s" | name: "%s" | prefix: "%s" | columns: [%s]`,
			p.Slug, p.Name, p.Prefix, strings.Join(colIDs, ", "))
		if len(p.Tags) > 0 {
			line += fmt.Sprintf(" | existing_tags: [%s]", strings.Join(p.Tags, ", "))
		}
		projLines = append(projLines, line)
	}

	sys := fmt.Sprintf(`You are a fast, precise project-routing and task-triage assistant for Kingfisher.
Available projects:
%s

Rules:
1. Route EACH task to the matching project's "slug" based on the user's text. If a task does not mention a specific project, assign it to the default project "%s".
2. If the user describes MULTIPLE distinct tasks, bugs, or ideas (for the same project or different projects), split them into separate objects in the "items" array. If it is one task with supporting details, return 1 object in "items".
3. Return ONLY valid JSON with this exact shape:
{
  "items": [
    {
      "project": "<project slug>",
      "title": "<concise imperative title, <= 70 chars>",
      "status": "<valid column id for that project; default %s>",
      "priority": "low | medium | high | critical",
      "tags": ["tag1", "tag2"],
      "description": "<cleanly formatted concise markdown preserving all user details>"
    }
  ]
}`, strings.Join(projLines, "\n"), defaultProjectSlug, defaultStatus)

	return sys, rawContext
}

func callLocalLLMBatch(ctx context.Context, cfg Config, rawContext string, projects []ProjectContext, defaultProjectSlug, defaultStatus string) (BatchResult, error) {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout < 15*time.Second {
		timeout = 15 * time.Second
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	modelName := cfg.Model
	if strings.EqualFold(cfg.Provider, "ollama") {
		st := CheckStatus(reqCtx, cfg)
		if !st.Available {
			return BatchResult{}, fmt.Errorf("ollama unavailable: %s", st.Message)
		}
		if st.Model != "" {
			modelName = st.Model
		}
	}

	sysPrompt, userPrompt := buildBatchPrompt(rawContext, projects, defaultProjectSlug, defaultStatus)
	client := &http.Client{Timeout: timeout}

	var content string
	var err error
	if strings.EqualFold(cfg.Provider, "openai") {
		content, err = callOpenAICompat(reqCtx, client, cfg.Endpoint, modelName, sysPrompt, userPrompt)
	} else {
		content, err = callOllama(reqCtx, client, cfg.Endpoint, modelName, sysPrompt, userPrompt)
	}
	if err != nil {
		return BatchResult{}, err
	}

	return parseBatchModelJSON(content, rawContext, projects, defaultProjectSlug, defaultStatus, modelName)
}

type rawModelItem struct {
	Project     string   `json:"project"`
	Title       string   `json:"title"`
	Status      string   `json:"status"`
	Priority    string   `json:"priority"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
}

func parseBatchModelJSON(raw, fallbackContext string, projects []ProjectContext, defaultProjectSlug, defaultStatus, modelName string) (BatchResult, error) {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var rawItems []rawModelItem

	// Shape 1: {"items": [...]}
	var wrapper struct {
		Items []rawModelItem `json:"items"`
	}
	if err := json.Unmarshal([]byte(cleaned), &wrapper); err == nil && len(wrapper.Items) > 0 {
		rawItems = wrapper.Items
	}

	// Shape 2: bare JSON array [...]
	if len(rawItems) == 0 {
		var arr []rawModelItem
		if err := json.Unmarshal([]byte(cleaned), &arr); err == nil && len(arr) > 0 {
			rawItems = arr
		}
	}

	// Shape 3: single JSON object {...}
	if len(rawItems) == 0 {
		if start := strings.Index(cleaned, "{"); start >= 0 {
			if end := strings.LastIndex(cleaned, "}"); end > start {
				sub := cleaned[start : end+1]
				if err := json.Unmarshal([]byte(sub), &wrapper); err == nil && len(wrapper.Items) > 0 {
					rawItems = wrapper.Items
				} else {
					var single rawModelItem
					if err := json.Unmarshal([]byte(sub), &single); err == nil && strings.TrimSpace(single.Title) != "" {
						rawItems = []rawModelItem{single}
					}
				}
			}
		}
	}

	if len(rawItems) == 0 {
		return BatchResult{}, fmt.Errorf("model returned no valid items")
	}

	results := make([]Result, 0, len(rawItems))
	for _, it := range rawItems {
		title := strings.TrimSpace(it.Title)
		if title == "" {
			continue
		}
		proj := resolveProject(it.Project, projects, defaultProjectSlug)
		status := resolveColumnStatus(it.Status, proj.Columns, defaultStatus)
		desc := strings.TrimSpace(it.Description)
		if desc == "" && len(rawItems) == 1 {
			desc = strings.TrimSpace(fallbackContext)
		}
		results = append(results, Result{
			ProjectSlug: proj.Slug,
			Title:       title,
			Status:      status,
			Priority:    board.NormalizePriority(it.Priority),
			Tags:        cleanTags(it.Tags),
			Description: desc,
			UsedAI:      true,
			Model:       modelName,
		})
	}

	if len(results) == 0 {
		return BatchResult{}, fmt.Errorf("no valid tasks after normalization")
	}
	return BatchResult{Items: results, UsedAI: true, Model: modelName}, nil
}

func resolveProject(hint string, projects []ProjectContext, defaultSlug string) ProjectContext {
	q := strings.ToLower(strings.TrimSpace(hint))
	if q != "" {
		for _, p := range projects {
			if strings.EqualFold(p.Slug, q) || strings.EqualFold(p.Name, q) || strings.EqualFold(p.Prefix, q) {
				return p
			}
		}
		// Partial substring match on name/slug
		for _, p := range projects {
			if strings.Contains(strings.ToLower(p.Name), q) || strings.Contains(strings.ToLower(p.Slug), q) {
				return p
			}
		}
	}
	for _, p := range projects {
		if strings.EqualFold(p.Slug, defaultSlug) {
			return p
		}
	}
	return projects[0]
}

func resolveColumnStatus(status string, cols []board.Column, defaultStatus string) string {
	s := strings.ToLower(strings.TrimSpace(status))
	for _, c := range cols {
		if strings.EqualFold(c.ID, s) || strings.EqualFold(c.Name, s) {
			return c.ID
		}
	}
	for _, c := range cols {
		if strings.EqualFold(c.ID, defaultStatus) {
			return c.ID
		}
	}
	if len(cols) > 0 {
		return cols[0].ID
	}
	return "ready"
}

var listItemPrefixRe = regexp.MustCompile(`^\s*(?:[-*•]|\d+[.)])\s+`)

// HeuristicParseBatch splits multi-item lists/paragraphs and routes each item to a matching project
// when no local LLM is available.
func HeuristicParseBatch(rawContext string, projects []ProjectContext, defaultProjectSlug, defaultStatus string) BatchResult {
	chunks := splitMultiItemChunks(rawContext)
	items := make([]Result, 0, len(chunks))

	for _, chunk := range chunks {
		proj := matchProjectInText(chunk, projects, defaultProjectSlug)
		b := &board.Board{Columns: proj.Columns}
		if len(b.Columns) == 0 {
			b.Columns = board.DefaultColumns()
		}
		res := HeuristicParse(chunk, b, defaultStatus)
		res.ProjectSlug = proj.Slug
		items = append(items, res)
	}

	return BatchResult{Items: items, UsedAI: false}
}

func splitMultiItemChunks(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	lines := strings.Split(raw, "\n")
	listLineCount := 0
	for _, l := range lines {
		if listItemPrefixRe.MatchString(l) {
			listLineCount++
		}
	}

	// If 2+ bullet/numbered lines exist, split per list item
	if listLineCount >= 2 {
		var chunks []string
		var cur []string
		for _, l := range lines {
			trimmed := strings.TrimSpace(l)
			if trimmed == "" {
				continue
			}
			if listItemPrefixRe.MatchString(l) {
				if len(cur) > 0 {
					chunks = append(chunks, strings.Join(cur, "\n"))
				}
				cur = []string{strings.TrimSpace(listItemPrefixRe.ReplaceAllString(l, ""))}
			} else {
				cur = append(cur, trimmed)
			}
		}
		if len(cur) > 0 {
			chunks = append(chunks, strings.Join(cur, "\n"))
		}
		if len(chunks) > 0 {
			return chunks
		}
	}

	return []string{raw}
}

func matchProjectInText(text string, projects []ProjectContext, defaultSlug string) ProjectContext {
	lower := strings.ToLower(text)
	for _, p := range projects {
		name := strings.ToLower(strings.TrimSpace(p.Name))
		slug := strings.ToLower(strings.TrimSpace(p.Slug))
		if len(name) >= 3 && containsWord(lower, name) {
			return p
		}
		if len(slug) >= 3 && slug != name && containsWord(lower, slug) {
			return p
		}
	}
	return resolveProject("", projects, defaultSlug)
}

func containsWord(haystack, word string) bool {
	idx := strings.Index(haystack, word)
	for idx >= 0 {
		beforeOK := idx == 0 || !unicode.IsLetter(rune(haystack[idx-1]))
		afterIdx := idx + len(word)
		afterOK := afterIdx >= len(haystack) || !unicode.IsLetter(rune(haystack[afterIdx]))
		if beforeOK && afterOK {
			return true
		}
		next := strings.Index(haystack[idx+1:], word)
		if next < 0 {
			break
		}
		idx = idx + 1 + next
	}
	return false
}
