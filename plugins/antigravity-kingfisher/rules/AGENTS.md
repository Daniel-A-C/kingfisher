# Kingfisher Task Synchronization Protocol

Whenever a workspace contains a `.kingfisher.yaml` file in its root, treat it as the live local task board for that repository and keep it automatically synchronized without waiting for the user to remind you.

## 1. CLI Discovery
Prefer `kf` if available on `PATH`; otherwise use `~/go/bin/kf`. If neither binary is available, read and update `.kingfisher.yaml` directly while preserving its YAML schema and setting `updated` to the current UTC ISO-8601 timestamp (`YYYY-MM-DDTHH:MM:SSZ`).

## 2. Starting Work
- **When asked for the "next task" or to pick up work**:
  1. Run `kf next --json` (or `~/go/bin/kf next --json`) to identify the highest-priority `ready` task.
  2. Immediately move that task to `active` (`kf move <ID> active --note "Started <ID>: <short plan>"`) before editing code.
- **When asked to work on a specific task ID or topic**:
  1. Check `.kingfisher.yaml` (or `kf list`) to see if an existing ticket matches the request.
  2. If a matching ticket exists and is in `backlog` or `ready`, move it to `active` before starting implementation.

## 3. During Work
- If you uncover an out-of-scope bug or follow-up item during implementation, mention it to the user and offer to file it (or file it directly when requested) via `kf capture "<context>"` or `kf add "<title>" --priority <low|medium|high|critical> --tags <t1,t2>`.

## 4. Completing Work
- As soon as all requested edits, tests, and builds for a Kingfisher task pass:
  1. **Automatically** move the task to `done` using:
     `kf move <ID> done --note "<concise summary of what was implemented and verified>"`
  2. Do **not** wait for the user to remind you to close or update the ticket.
  3. Include the task ID and updated status (`done`) in your response.
- If work is blocked or paused mid-task, leave the task in `active` and append a brief status note in `agent_notes` explaining what remains.
