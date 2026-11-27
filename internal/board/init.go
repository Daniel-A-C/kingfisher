package board

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// DerivePrefix computes a concise 2-4 character uppercase ticket prefix from a project name.
func DerivePrefix(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "KF"
	}

	// Split on separators or camelCase transitions
	var words []string
	var cur []rune
	runes := []rune(trimmed)
	for i, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			if len(cur) > 0 {
				words = append(words, string(cur))
				cur = nil
			}
			continue
		}
		if i > 0 && unicode.IsUpper(r) && len(cur) > 0 && unicode.IsLower(runes[i-1]) {
			words = append(words, string(cur))
			cur = nil
		}
		cur = append(cur, r)
	}
	if len(cur) > 0 {
		words = append(words, string(cur))
	}

	if len(words) >= 2 {
		var b strings.Builder
		for _, w := range words {
			for _, r := range w {
				b.WriteRune(unicode.ToUpper(r))
				break
			}
			if b.Len() >= 4 {
				break
			}
		}
		return b.String()
	}

	if len(words) == 1 {
		upper := strings.ToUpper(words[0])
		rs := []rune(upper)
		if len(rs) <= 3 {
			return upper
		}
		// Keep first letter + next 2 consonants (e.g. Corvidae -> CRV, Rookery -> RKR)
		var out []rune
		out = append(out, rs[0])
		for _, r := range rs[1:] {
			if !strings.ContainsRune("AEIOU", r) {
				out = append(out, r)
				if len(out) == 3 {
					break
				}
			}
		}
		for i := 1; len(out) < 3 && i < len(rs); i++ {
			out = append(out, rs[i])
		}
		return string(out)
	}

	return "KF"
}

// InitProject creates a fresh .kingfisher.yaml in projectDir and ensures it is gitignored.
// Returns the created board, the ignore file path updated (if any), and any error.
func InitProject(projectDir, name, prefix string, useGitignore bool) (*Board, string, error) {
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, "", fmt.Errorf("resolve project dir: %w", err)
	}

	if Exists(absDir) {
		b, _, err := Load(absDir)
		if err != nil {
			return nil, "", err
		}
		ignoredFile, _ := EnsureIgnored(absDir, useGitignore)
		return b, ignoredFile, nil
	}

	if strings.TrimSpace(name) == "" {
		name = filepath.Base(absDir)
	}
	if strings.TrimSpace(prefix) == "" {
		prefix = DerivePrefix(name)
	} else {
		prefix = strings.ToUpper(strings.TrimSpace(prefix))
	}

	b := &Board{
		Project: ProjectMeta{
			Name:   name,
			Prefix: prefix,
			NextID: 1,
		},
		Columns: DefaultColumns(),
		Tasks:   []Task{},
	}

	if _, err := SaveAtomic(absDir, b); err != nil {
		return nil, "", err
	}

	ignoredFile, err := EnsureIgnored(absDir, useGitignore)
	if err != nil {
		// Non-fatal: board was created, return warning via error only if critical
		return b, "", fmt.Errorf("board created, but failed to update ignore file: %w", err)
	}

	return b, ignoredFile, nil
}

// EnsureIgnored idempotently adds .kingfisher.yaml to .git/info/exclude (preferred when
// projectDir is a git repo) or .gitignore if useGitignore is true or no .git directory exists.
func EnsureIgnored(projectDir string, useGitignore bool) (string, error) {
	gitDir := filepath.Join(projectDir, ".git")
	var ignorePath string

	if !useGitignore {
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			infoDir := filepath.Join(gitDir, "info")
			if err := os.MkdirAll(infoDir, 0o755); err == nil {
				ignorePath = filepath.Join(infoDir, "exclude")
			}
		}
	}

	if ignorePath == "" {
		ignorePath = filepath.Join(projectDir, ".gitignore")
	}

	existing, err := os.ReadFile(ignorePath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	for _, line := range strings.Split(string(existing), "\n") {
		if strings.TrimSpace(line) == BoardFilename {
			return ignorePath, nil
		}
	}

	var prefix string
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		prefix = "\n"
	}
	entry := prefix + "# Kingfisher local project board\n" + BoardFilename + "\n"

	f, err := os.OpenFile(ignorePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return "", err
	}
	return ignorePath, nil
}
