# md-notes

<!-- workflow-docs:start:tagline -->
> A small CLI for terminal-resident developers that creates and finds plain-`.md` notes on disk so they can take scratch notes and recover them in seconds without an app lock-in.
<!-- workflow-docs:end:tagline -->

<!-- workflow-docs:start:overview -->
## Overview

md-notes replaces the "sticky notes and random `.txt` scratch files" workflow with a single CLI that writes dated markdown files into a user-owned folder and finds them again by half-remembered phrase. Notes are plain `.md` files — sync them with iCloud, git, or Syncthing (or none); edit them in any editor; the tool stays out of the way.
<!-- workflow-docs:end:overview -->

<!-- workflow-docs:start:who-for -->
## Who it's for

The author — a terminal-resident developer who takes frequent scratch notes during the day and currently loses them in folders. md-notes is not designed for a broader audience; adoption by others is incidental.
<!-- workflow-docs:end:who-for -->

<!-- workflow-docs:start:scope -->
## Scope

### In scope (v0.1)

- `notes new "<title>"` — create a new note with date + empty-tags frontmatter, opened in `$EDITOR`.
- `notes find <query>` — full-text search across titles and bodies, fast at thousands-of-notes scale.
- `notes ls` — list notes by modification recency, most-recent first.
- Configurable notes directory (default `~/notes/`).
- Empty `tags: []` written into every new note's frontmatter — so v0.2 tagging does not require a migration of older notes.

### Out of scope

Product-level non-goals (permanent):

- Web UI.
- Built-in sync (iCloud / git / Syncthing remain the user's responsibility).
- Encryption at rest or in transit.
- Plugin system.
- Cross-note linking, backlinks, or graph view.
- Built-in TUI viewer for search results.
- Background daemon or long-running index-maintenance service.

Deferred to v0.2 candidates:

- `notes find --tag <tag>` and `#tag` filtering.
- `notes tag <note> <tag>` mutator command.
- `notes today` daily-note shortcut.
<!-- workflow-docs:end:scope -->

<!-- workflow-docs:start:key-decisions -->
## Key decisions

- **ADR-005:** in-process scan in Go for full-text search; preserves single-binary distribution and the no-daemon constraint.
- **ADR-004:** minimal YAML frontmatter (`title`, `date`, `tags: []`) — stable shape so v0.2 tagging adds keys without migrating older notes.
- **ADR-003:** filenames as `YYYY-MM-DD-slug.md` with numeric collision suffix; sorts naturally and prevents same-title-same-day overwrites.
- **ADR-002:** configuration via `--dir` flag > `MD_NOTES_DIR` env > `~/.config/md-notes/config.toml` > `$HOME/notes` default.
- **ADR-001:** Go as the implementation language — matches the single-binary distribution constraint.

See `Design/adr/` for full ADRs.
<!-- workflow-docs:end:key-decisions -->

<!-- workflow-docs:start:more -->
## More

- Workflow rules and conventions: [`CLAUDE.md`](CLAUDE.md)
- Design artefacts: [`Design/`](Design/) — PRD, MVP, build-out plan, ADRs
- Per-issue prompts: [`prompts/`](prompts/)
<!-- workflow-docs:end:more -->
