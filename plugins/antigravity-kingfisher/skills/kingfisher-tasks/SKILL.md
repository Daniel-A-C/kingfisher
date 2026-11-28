---
name: kingfisher-tasks
description: Manage local .kingfisher.yaml task boards using the kf CLI (pick next tasks, move tasks across columns, capture ideas, and record agent notes). Use when working in any repository that has or needs a .kingfisher.yaml board.
---

# Kingfisher Task Management

Kingfisher stores a single gitignored `.kingfisher.yaml` file at the root of each project repository (excluded via `.git/info/exclude`). The desktop app watches this file in real time via `fsnotify`, so CLI or direct YAML updates appear live on the board.

## Locating the CLI

Check `kf` on `PATH`, falling back to `~/go/bin/kf`:

```sh
KF=$(command -v kf || echo "$HOME/go/bin/kf")
```

## Core Commands

```sh
# Initialize a new board in the current repo (adds .kingfisher.yaml to .git/info/exclude)
$KF init

# Get the highest-priority ready task as JSON
$KF next --json

# List tasks (optionally filtered by status: backlog, ready, active, done)
$KF list
$KF list --status ready

# Move a task and append an agent note (automatically updates timestamp)
$KF move KNG-12 active --note "Started implementing feature X"
$KF move KNG-12 done --note "Implemented feature X and verified with unit tests"

# Capture freeform context (uses local LLM or heuristic fallback; routes across projects)
$KF capture "high: fix race condition in watcher debounce"

# Add a structured task directly to the current repo
$KF add "Add keyboard navigation" --priority high --tags ux,keyboard --status ready
```

## Board Schema Reference

Columns default to:
- `backlog` — Backlog
- `ready` — Ready to Dive
- `active` — In Flight
- `done` — Landed

Priorities (highest to lowest): `critical`, `high`, `medium`, `low`.

Each task entry contains:
- `id`: `<PREFIX>-<N>` (e.g. `KNG-11`)
- `title`: short imperative summary
- `status`: column ID
- `priority`: `low | medium | high | critical`
- `tags`: list of lowercase tags
- `created` / `updated`: UTC RFC3339 timestamps
- `description`: structured details
- `raw_input`: optional verbatim original human capture text
- `agent_notes`: newline-delimited progress/completion notes from coding agents

## Best Practices

1. Always move a ticket to `active` as soon as you begin working on it so the live board reflects in-flight work.
2. Always move a ticket to `done` with a concise `--note` summarizing changes and verification immediately after tests/builds succeed.
3. Preserve `raw_input` verbatim if editing `.kingfisher.yaml` manually.
