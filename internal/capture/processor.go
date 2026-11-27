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

// Result is the structured task extracted from raw user context.
type Result struct {
	Title       string   `json:"title"`
	Status      string   `json:"status"`
	Priority    string   `json:"priority"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	UsedAI      bool     `json:"used_ai"`
	Model       string   `json:"model,omitempty"`
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
