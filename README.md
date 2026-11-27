# Kingfisher

A fast, local-first macOS project manager with an **Art Nouveau stained-glass aesthetic**, **native transparency**, **browser-style project tabs**, and **single-file per-repo `.kingfisher.yaml` storage** synced live with your coding agents.

## Highlights

- **One Gitignored File Per Repo (`.kingfisher.yaml`)**: Task data lives directly inside each repository root and is automatically ignored via `.git/info/exclude` (zero `.gitignore` diff noise).
- **Instant Summon (`Super + F` via Raycast)**: Kingfisher stays warm in the background (`HideWindowOnClose: true`), so summoning it via Raycast feels near-instant.
- **Browser-Style Project Tabs**: Jump directly to projects with `⌘1`–`⌘8`, `⌘9` for the last tab, `⌘K` / `⌘T` for the fuzzy Project Aviary switcher, and `⌘N` for Quick Capture.
- **Real-Time Agent Sync**: Live `fsnotify` watchers pick up edits from coding agents or the `kf` CLI in real time, with graceful recovery if malformed YAML is saved mid-edit.
- **Stained-Glass Transparency**: Native macOS window translucency with an in-app opacity slider (`0%`–`85%`) and `⌘⇧T` preset cycling so your Rookery wallpapers shine through.

## Build & Install

```sh
# Run unit tests + frontend typecheck & build
make test

# Build Kingfisher.app and the kf CLI binary
make build

# Install /Applications/Kingfisher.app and ~/go/bin/kf
make install
```

## Raycast Instant Summon Setup

1. Open **Raycast Settings → Extensions → Applications**.
2. Find **Kingfisher** and assign your preferred hotkey (e.g. `Super + F`).
3. Pressing `Super + F` brings Kingfisher straight to the front; pressing `Esc` tucks it away cleanly while keeping it warm in memory.

## Companion CLI (`kf`)

```sh
kf init                          # Create .kingfisher.yaml & add to .git/info/exclude
kf                               # Open current repo's board in Kingfisher
kf next                          # Show the highest-priority ready task
kf next --json                   # Machine-readable next task for coding agents
kf list [--status ready]         # List tasks
kf add "Implement fuzzy search" --priority high --tags cli,ux
kf move KF-1 active --note "Started working on fuzzy matcher"
kf move KF-1 done --note "Completed and verified with go test ./..."
```

## Drop-In Coding Agent Instruction

Add this snippet to your global Cursor / Claude / Gemini / Antigravity rules:

> If `.kingfisher.yaml` exists in the project root, it is the local task board. When asked to work on a Kingfisher task, read `.kingfisher.yaml` (or run `kf next --json`), move the task's status to `active`, implement the changes, then update `status` to `done` (or run `kf move <ID> done --note "..."`), set `updated` to the current ISO timestamp, and summarize changes in `agent_notes`.
