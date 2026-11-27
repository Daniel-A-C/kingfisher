package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Daniel-A-C/kingfisher/internal/board"
	"github.com/Daniel-A-C/kingfisher/internal/ipc"
	"github.com/Daniel-A-C/kingfisher/internal/registry"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return cmdOpen(nil, stdout, stderr)
	}

	switch args[0] {
	case "init":
		return cmdInit(args[1:], stdout, stderr)
	case "open":
		return cmdOpen(args[1:], stdout, stderr)
	case "next":
		return cmdNext(args[1:], stdout, stderr)
	case "list", "ls":
		return cmdList(args[1:], stdout, stderr)
	case "add":
		return cmdAdd(args[1:], stdout, stderr)
	case "move", "mv":
		return cmdMove(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		// Treat unknown single arg as `kf open <slug-or-path>` for fast jumping
		if !strings.HasPrefix(args[0], "-") {
			return cmdOpen(args, stdout, stderr)
		}
		fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `kf — Kingfisher CLI companion

Usage:
  kf                          Open current directory's board in Kingfisher
  kf open [slug-or-path]      Focus Kingfisher on a project by slug or path
  kf init [flags]             Initialize .kingfisher.yaml (ignored via .git/info/exclude)
  kf next [--json]            Print the highest-priority ready task
  kf list [--status S] [--json] List tasks on the current project board
  kf add <title> [flags]      Add a task to .kingfisher.yaml
  kf move <ID> <status> [--note "..."]  Move a task and append agent_notes`)
}

func loadRegistry() (*registry.Registry, error) {
	p, err := registry.DefaultPath()
	if err != nil {
		return nil, err
	}
	return registry.Load(p)
}

func cmdInit(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	name := fs.String("name", "", "Project display name (defaults to folder name)")
	prefix := fs.String("prefix", "", "Ticket ID prefix (e.g. CRV, KF)")
	useGitignore := fs.Bool("gitignore", false, "Add .kingfisher.yaml to .gitignore instead of .git/info/exclude")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	b, ignoredPath, err := board.InitProject(dir, *name, *prefix, *useGitignore)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	if reg, err := loadRegistry(); err == nil {
		reg.RegisterProject(dir, b.Project.Name, b.Project.Prefix)
		_ = reg.Save()
	}

	fmt.Fprintf(stdout, "Initialized %s (%s) in %s\n", b.Project.Name, b.Project.Prefix, board.BoardPath(dir))
	if ignoredPath != "" {
		fmt.Fprintf(stdout, "Ignored via %s\n", ignoredPath)
	}
	return 0
}

func cmdOpen(args []string, stdout, stderr io.Writer) int {
	var msg ipc.Message
	msg.Action = "open"

	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		target := args[0]
		if info, err := os.Stat(target); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(target)
			msg.Path = abs
		} else {
			msg.Slug = target
		}
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		if board.Exists(cwd) {
			msg.Path = cwd
		}
	}

	sock := ipc.SocketPath()
	if err := ipc.Send(sock, msg); err == nil {
		return 0
	}

	// Kingfisher isn't running yet — try launching Kingfisher.app on macOS and retry IPC
	_ = exec.Command("open", "-a", "Kingfisher").Start()
	for i := 0; i < 15; i++ {
		time.Sleep(150 * time.Millisecond)
		if err := ipc.Send(sock, msg); err == nil {
			return 0
		}
	}

	fmt.Fprintf(stderr, "Kingfisher desktop app is not running (could not connect to %s).\n", sock)
	fmt.Fprintln(stdout, "Tip: Run `make install` to install Kingfisher.app into /Applications.")
	return 1
}

func cmdNext(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("next", flag.ContinueOnError)
	fs.SetOutput(stderr)
	asJSON := fs.Bool("json", false, "Output as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	b, _, err := board.Load(dir)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	task, ok := b.NextReadyTask()
	if !ok {
		if *asJSON {
			fmt.Fprintln(stdout, "null")
		} else {
			fmt.Fprintln(stdout, "No ready or backlog tasks.")
		}
		return 0
	}

	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(task)
		return 0
	}

	fmt.Fprintf(stdout, "%s [%s | %s] %s\n", task.ID, task.Status, task.Priority, task.Title)
	if len(task.Tags) > 0 {
		fmt.Fprintf(stdout, "Tags: %s\n", strings.Join(task.Tags, ", "))
	}
	if task.Description != "" {
		fmt.Fprintf(stdout, "\n%s\n", task.Description)
	}
	if task.AgentNotes != "" {
		fmt.Fprintf(stdout, "\nAgent Notes:\n%s\n", task.AgentNotes)
	}
	return 0
}

func cmdList(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	statusFilter := fs.String("status", "", "Filter by column status ID")
	asJSON := fs.Bool("json", false, "Output as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	b, _, err := board.Load(dir)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	tasks := make([]board.Task, 0, len(b.Tasks))
	for _, t := range b.Tasks {
		if *statusFilter != "" && !strings.EqualFold(t.Status, *statusFilter) {
			continue
		}
		tasks = append(tasks, t)
	}

	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(tasks)
		return 0
	}

	if len(tasks) == 0 {
		fmt.Fprintln(stdout, "No tasks found.")
		return 0
	}
	for _, t := range tasks {
		fmt.Fprintf(stdout, "%-8s %-8s %-8s %s\n", t.ID, t.Status, t.Priority, t.Title)
	}
	return 0
}

func cmdAdd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(stderr)
	status := fs.String("status", "ready", "Column status ID (backlog|ready|active|done)")
	priority := fs.String("priority", "medium", "Priority (low|medium|high|critical)")
	tagsRaw := fs.String("tags", "", "Comma-separated tags")
	desc := fs.String("desc", "", "Task description")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	title := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if title == "" {
		fmt.Fprintln(stderr, "error: task title is required (e.g. kf add \"Fix parser\")")
		return 2
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	b, _, err := board.Load(dir)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	var tags []string
	if strings.TrimSpace(*tagsRaw) != "" {
		for _, tg := range strings.Split(*tagsRaw, ",") {
			if s := strings.TrimSpace(tg); s != "" {
				tags = append(tags, s)
			}
		}
	}

	task := b.AddTask(title, *status, *priority, *desc, tags)
	if _, err := board.SaveAtomic(dir, b); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Added %s [%s | %s] %s\n", task.ID, task.Status, task.Priority, task.Title)
	return 0
}

func cmdMove(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(stderr, "usage: kf move <ID> <status> [--note \"...\"]")
		return 2
	}
	id := args[0]
	targetStatus := args[1]

	fs := flag.NewFlagSet("move", flag.ContinueOnError)
	fs.SetOutput(stderr)
	note := fs.String("note", "", "Append note to agent_notes")
	if err := fs.Parse(args[2:]); err != nil {
		return 2
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	b, _, err := board.Load(dir)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	task, err := b.MoveTask(id, targetStatus, -1, *note)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	if _, err := board.SaveAtomic(dir, b); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Moved %s -> %s\n", task.ID, task.Status)
	return 0
}
